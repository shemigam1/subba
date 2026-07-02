package httpapi_test

// Integration tests for the fully-wired router, driven over real HTTP against the
// local infra from `make up` (postgres :5433, redis, rabbitmq). When any dependency
// is unreachable the whole suite skips, so `go test ./...` stays green without Docker.
//
// Isolation notes:
//   - Every test signs up fresh tenants under @apitest.subba.dev; TestMain deletes
//     them afterwards (tenant deletes cascade to plans/customers/tokens).
//   - The rate limiter keys on client IP, so each test client sends a unique
//     X-Forwarded-For (gin's default trusted-proxy config honors it). That keeps
//     repeated runs out of each other's rate-limit buckets — except the rate-limit
//     test, which reuses one IP on purpose.

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"github.com/shamigam1/subba/internal/config"
	httpapi "github.com/shamigam1/subba/internal/http"
	"github.com/shamigam1/subba/internal/platform"
)

const testEmailDomain = "apitest.subba.dev"

var (
	testSrv  *httptest.Server
	testPlat *platform.Platform
	logBuf   *syncBuffer
	setupErr error
)

func TestMain(m *testing.M) {
	code := func() int {
		// Best-effort env: the repo .env first, then local-dev defaults so the
		// suite also runs where only the compose stack exists.
		_ = godotenv.Load("../../.env")
		setenvDefault("DATABASE_URL", "postgres://subba_app:subba_app@localhost:5433/subba?sslmode=disable")
		setenvDefault("ADMIN_DATABASE_URL", "postgres://postgres:postgres@localhost:5433/subba?sslmode=disable")
		if os.Getenv("MASTER_ENCRYPTION_KEY") == "" {
			key := make([]byte, 32)
			_, _ = rand.Read(key)
			os.Setenv("MASTER_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(key))
		}
		// Force the dev mail fallback: the magic-link tests read the link from logs.
		os.Setenv("RESEND_API_KEY", "")
		os.Setenv("APP_ENV", "development")

		cfg, err := config.Load()
		if err != nil {
			setupErr = err
			return m.Run()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		testPlat, err = platform.New(ctx, cfg)
		if err != nil {
			setupErr = err
			return m.Run()
		}
		defer testPlat.Close()
		defer cleanupTestTenants()

		logBuf = &syncBuffer{}
		testSrv = httptest.NewServer(httpapi.NewRouter(cfg, zerolog.New(logBuf), testPlat))
		defer testSrv.Close()

		return m.Run()
	}()
	os.Exit(code)
}

func setup(t *testing.T) {
	t.Helper()
	if setupErr != nil {
		t.Skipf("integration infra unavailable (run `make up` + `make migrate`): %v", setupErr)
	}
}

func cleanupTestTenants() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, _ = testPlat.AdminDB.Exec(ctx, "DELETE FROM tenants WHERE email LIKE '%@'||$1", testEmailDomain)
}

// ---------------------------------------------------------------- helpers

func setenvDefault(key, val string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, val)
	}
}

// syncBuffer lets the router's zerolog writes race safely with test-side reads.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func uniqueEmail(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s@%s", prefix, hex.EncodeToString(b), testEmailDomain)
}

func randomIP() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return fmt.Sprintf("10.%d.%d.%d", b[0], b[1], b[2])
}

// client is one browser-like caller: its own cookie jar and client IP.
type client struct {
	t    *testing.T
	http *http.Client
	ip   string
}

func newClient(t *testing.T) *client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &client{t: t, http: &http.Client{Jar: jar}, ip: randomIP()}
}

func (c *client) do(method, path string, body any) (int, []byte) {
	c.t.Helper()
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, testSrv.URL+path, rdr)
	if err != nil {
		c.t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Forwarded-For", c.ip)
	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatal(err)
	}
	return resp.StatusCode, raw
}

// signup registers a fresh tenant and returns its id (the session cookie lands in the jar).
func (c *client) signup(name, email, password string) string {
	c.t.Helper()
	status, body := c.do("POST", "/v1/auth/signup", map[string]string{
		"name": name, "email": email, "password": password,
	})
	if status != http.StatusCreated {
		c.t.Fatalf("signup = %d, want 201: %s", status, body)
	}
	var tenant struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &tenant); err != nil || tenant.ID == "" {
		c.t.Fatalf("signup response missing id: %s", body)
	}
	return tenant.ID
}

func decode[T any](t *testing.T, raw []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("decode %T from %s: %v", v, raw, err)
	}
	return v
}

// magicLinkToken scans the captured logs for the dev-fallback magic link sent to email.
func magicLinkToken(t *testing.T, email string) string {
	t.Helper()
	sc := bufio.NewScanner(strings.NewReader(logBuf.String()))
	token := ""
	for sc.Scan() {
		var entry struct {
			To   string `json:"to"`
			Link string `json:"link"`
		}
		if json.Unmarshal(sc.Bytes(), &entry) != nil || entry.To != email {
			continue
		}
		u, err := url.Parse(entry.Link)
		if err != nil {
			t.Fatalf("unparseable magic link %q: %v", entry.Link, err)
		}
		token = u.Query().Get("token") // keep scanning: last link wins
	}
	if token == "" {
		t.Fatalf("no magic link logged for %s", email)
	}
	return token
}

// ---------------------------------------------------------------- probes

func TestHealthAndReadiness(t *testing.T) {
	setup(t)
	c := newClient(t)

	status, body := c.do("GET", "/healthz", nil)
	if status != http.StatusOK {
		t.Fatalf("/healthz = %d: %s", status, body)
	}
	status, body = c.do("GET", "/readyz", nil)
	if status != http.StatusOK {
		t.Fatalf("/readyz = %d: %s", status, body)
	}
}

// ---------------------------------------------------------------- dashboard auth

func TestSignupLoginSession(t *testing.T) {
	setup(t)
	c := newClient(t)
	email := uniqueEmail("auth")

	c.signup("Auth Test Co", email, "correcthorsebattery")

	status, body := c.do("GET", "/v1/me", nil)
	if status != http.StatusOK {
		t.Fatalf("/me after signup = %d: %s", status, body)
	}
	me := decode[struct {
		Email string `json:"email"`
	}](t, body)
	if me.Email != email {
		t.Fatalf("/me email = %q, want %q", me.Email, email)
	}

	// Same email again → conflict.
	status, body = c.do("POST", "/v1/auth/signup", map[string]string{
		"name": "Dup", "email": email, "password": "correcthorsebattery",
	})
	if status != http.StatusConflict {
		t.Fatalf("duplicate signup = %d, want 409: %s", status, body)
	}

	// Logout kills the session.
	if status, body = c.do("POST", "/v1/auth/logout", nil); status != http.StatusNoContent {
		t.Fatalf("logout = %d: %s", status, body)
	}
	if status, _ = c.do("GET", "/v1/me", nil); status != http.StatusUnauthorized {
		t.Fatalf("/me after logout = %d, want 401", status)
	}

	// Login restores access.
	status, body = c.do("POST", "/v1/auth/login", map[string]string{"email": email, "password": "correcthorsebattery"})
	if status != http.StatusOK {
		t.Fatalf("login = %d: %s", status, body)
	}
	if status, _ = c.do("GET", "/v1/me", nil); status != http.StatusOK {
		t.Fatalf("/me after login = %d, want 200", status)
	}
}

func TestLoginDoesNotRevealAccounts(t *testing.T) {
	setup(t)
	c := newClient(t)
	email := uniqueEmail("enum")
	c.signup("Enum Test Co", email, "correcthorsebattery")

	fresh := newClient(t)
	status, wrongPw := fresh.do("POST", "/v1/auth/login", map[string]string{"email": email, "password": "wrongpassword"})
	if status != http.StatusUnauthorized {
		t.Fatalf("wrong-password login = %d, want 401: %s", status, wrongPw)
	}
	status, noAccount := fresh.do("POST", "/v1/auth/login", map[string]string{"email": uniqueEmail("ghost"), "password": "wrongpassword"})
	if status != http.StatusUnauthorized {
		t.Fatalf("unknown-email login = %d, want 401: %s", status, noAccount)
	}

	// Identical error message either way, so responses can't be used to probe for accounts.
	type errBody struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	a, b := decode[errBody](t, wrongPw), decode[errBody](t, noAccount)
	if a != b {
		t.Fatalf("login errors differ by account existence: %+v vs %+v", a, b)
	}
}

func TestSignupValidation(t *testing.T) {
	setup(t)
	c := newClient(t)

	status, body := c.do("POST", "/v1/auth/signup", map[string]string{
		"name": "Short PW Co", "email": uniqueEmail("shortpw"), "password": "short",
	})
	if status != 422 {
		t.Fatalf("short-password signup = %d, want 422: %s", status, body)
	}

	req, _ := http.NewRequest("POST", testSrv.URL+"/v1/auth/signup", strings.NewReader("{not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", c.ip)
	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("malformed-JSON signup = %d, want 400", resp.StatusCode)
	}
}

// ---------------------------------------------------------------- RLS isolation

func TestTenantIsolation(t *testing.T) {
	setup(t)
	tenantA := newClient(t)
	tenantB := newClient(t)
	tenantA.signup("Tenant A", uniqueEmail("tenant-a"), "correcthorsebattery")
	tenantB.signup("Tenant B", uniqueEmail("tenant-b"), "correcthorsebattery")

	status, body := tenantA.do("POST", "/v1/plans", map[string]any{
		"name": "Pro", "amount_minor": 500000, "currency": "NGN", "interval": "month",
	})
	if status != http.StatusCreated {
		t.Fatalf("create plan = %d: %s", status, body)
	}
	plan := decode[struct {
		ID string `json:"id"`
	}](t, body)

	// B's plan list must not contain A's plan.
	status, body = tenantB.do("GET", "/v1/plans", nil)
	if status != http.StatusOK {
		t.Fatalf("B list plans = %d: %s", status, body)
	}
	if plans := decode[[]json.RawMessage](t, body); len(plans) != 0 {
		t.Fatalf("tenant B sees %d foreign plans: %s", len(plans), body)
	}

	// Even with the exact UUID, B gets a 404 — RLS holds on direct-ID reads.
	if status, body = tenantB.do("GET", "/v1/plans/"+plan.ID, nil); status != http.StatusNotFound {
		t.Fatalf("B GET A's plan by id = %d, want 404: %s", status, body)
	}
	if status, body = tenantB.do("DELETE", "/v1/plans/"+plan.ID, nil); status != http.StatusNotFound {
		t.Fatalf("B DELETE A's plan = %d, want 404: %s", status, body)
	}

	// Same for customers, in the other direction.
	status, body = tenantB.do("POST", "/v1/customers", map[string]string{
		"name": "B Customer", "email": uniqueEmail("cust-b"),
	})
	if status != http.StatusCreated {
		t.Fatalf("B create customer = %d: %s", status, body)
	}
	cust := decode[struct {
		ID string `json:"id"`
	}](t, body)

	status, body = tenantA.do("GET", "/v1/customers", nil)
	if status != http.StatusOK {
		t.Fatalf("A list customers = %d: %s", status, body)
	}
	list := decode[struct {
		Data []json.RawMessage `json:"data"`
	}](t, body)
	if len(list.Data) != 0 {
		t.Fatalf("tenant A sees %d foreign customers: %s", len(list.Data), body)
	}
	if status, body = tenantA.do("GET", "/v1/customers/"+cust.ID, nil); status != http.StatusNotFound {
		t.Fatalf("A GET B's customer by id = %d, want 404: %s", status, body)
	}
}

// ---------------------------------------------------------------- plans CRUD

func TestPlanLifecycle(t *testing.T) {
	setup(t)
	c := newClient(t)
	c.signup("Plans Co", uniqueEmail("plans"), "correcthorsebattery")

	status, body := c.do("POST", "/v1/plans", map[string]any{
		"name": "Starter", "amount_minor": 150000, "interval": "month",
	})
	if status != http.StatusCreated {
		t.Fatalf("create plan = %d: %s", status, body)
	}
	plan := decode[struct {
		ID       string `json:"id"`
		Currency string `json:"currency"`
	}](t, body)
	if plan.Currency != "NGN" {
		t.Fatalf("currency default = %q, want NGN", plan.Currency)
	}

	// Invalid interval is rejected.
	status, body = c.do("POST", "/v1/plans", map[string]any{
		"name": "Weekly", "amount_minor": 1000, "interval": "week",
	})
	if status != 422 {
		t.Fatalf("bad-interval plan = %d, want 422: %s", status, body)
	}

	if status, body = c.do("GET", "/v1/plans/"+plan.ID, nil); status != http.StatusOK {
		t.Fatalf("get plan = %d: %s", status, body)
	}
	if status, body = c.do("DELETE", "/v1/plans/"+plan.ID, nil); status != http.StatusNoContent {
		t.Fatalf("delete plan = %d: %s", status, body)
	}

	// Soft-deleted: default list hides it, include_deleted surfaces it.
	status, body = c.do("GET", "/v1/plans", nil)
	if status != http.StatusOK {
		t.Fatalf("list plans = %d: %s", status, body)
	}
	if plans := decode[[]json.RawMessage](t, body); len(plans) != 0 {
		t.Fatalf("deleted plan still listed: %s", body)
	}
	status, body = c.do("GET", "/v1/plans?include_deleted=true", nil)
	if status != http.StatusOK {
		t.Fatalf("list include_deleted = %d: %s", status, body)
	}
	if plans := decode[[]json.RawMessage](t, body); len(plans) != 1 {
		t.Fatalf("include_deleted returned %d plans, want 1: %s", len(plans), body)
	}
}

// ---------------------------------------------------------------- portal magic links

func TestPortalMagicLinkFlow(t *testing.T) {
	setup(t)
	dash := newClient(t)
	tenantID := dash.signup("Portal Co", uniqueEmail("portal-tenant"), "correcthorsebattery")

	custEmail := uniqueEmail("portal-cust")
	status, body := dash.do("POST", "/v1/customers", map[string]string{"name": "Portal Customer", "email": custEmail})
	if status != http.StatusCreated {
		t.Fatalf("create customer = %d: %s", status, body)
	}

	// Real and fake emails must be indistinguishable from outside.
	anon := newClient(t)
	status, realBody := anon.do("POST", "/v1/portal/access-request", map[string]string{"tenant_id": tenantID, "email": custEmail})
	if status != http.StatusAccepted {
		t.Fatalf("access-request (real) = %d: %s", status, realBody)
	}
	ghostEmail := uniqueEmail("ghost-cust")
	status, fakeBody := anon.do("POST", "/v1/portal/access-request", map[string]string{"tenant_id": tenantID, "email": ghostEmail})
	if status != http.StatusAccepted {
		t.Fatalf("access-request (fake) = %d: %s", status, fakeBody)
	}
	if string(realBody) != string(fakeBody) {
		t.Fatalf("access-request bodies differ by account existence:\nreal: %s\nfake: %s", realBody, fakeBody)
	}
	if strings.Contains(logBuf.String(), ghostEmail) {
		t.Fatalf("a magic link was issued for a nonexistent customer %s", ghostEmail)
	}

	// The dev fallback logged the link; exchange its token for a portal session.
	token := magicLinkToken(t, custEmail)
	portalClient := newClient(t)
	status, body = portalClient.do("POST", "/v1/portal/session", map[string]string{"token": token})
	if status != http.StatusOK {
		t.Fatalf("portal session exchange = %d: %s", status, body)
	}
	ctx := decode[struct {
		Customer struct {
			Email string `json:"email"`
		} `json:"customer"`
	}](t, body)
	if ctx.Customer.Email != custEmail {
		t.Fatalf("portal context customer = %q, want %q", ctx.Customer.Email, custEmail)
	}

	// Single-use: replaying the same token fails even with a fresh client.
	if status, body = newClient(t).do("POST", "/v1/portal/session", map[string]string{"token": token}); status != http.StatusUnauthorized {
		t.Fatalf("token replay = %d, want 401: %s", status, body)
	}

	if status, body = portalClient.do("GET", "/v1/portal/me", nil); status != http.StatusOK {
		t.Fatalf("/portal/me = %d: %s", status, body)
	}

	// The two auth domains don't cross: dashboard cookie ≠ portal cookie.
	if status, _ = dash.do("GET", "/v1/portal/me", nil); status != http.StatusUnauthorized {
		t.Fatalf("dashboard session on /portal/me = %d, want 401", status)
	}
	if status, _ = portalClient.do("GET", "/v1/me", nil); status != http.StatusUnauthorized {
		t.Fatalf("portal session on /me = %d, want 401", status)
	}

	// Portal logout revokes the session server-side.
	if status, _ = portalClient.do("POST", "/v1/portal/logout", nil); status != http.StatusNoContent {
		t.Fatalf("portal logout = %d, want 204", status)
	}
	if status, _ = portalClient.do("GET", "/v1/portal/me", nil); status != http.StatusUnauthorized {
		t.Fatalf("/portal/me after logout = %d, want 401", status)
	}
}

func TestPortalSessionValidation(t *testing.T) {
	setup(t)
	c := newClient(t)

	if status, body := c.do("POST", "/v1/portal/session", map[string]string{}); status != http.StatusBadRequest {
		t.Fatalf("missing token = %d, want 400: %s", status, body)
	}
	if status, body := c.do("POST", "/v1/portal/session", map[string]string{"token": "garbage-token"}); status != http.StatusUnauthorized {
		t.Fatalf("garbage token = %d, want 401: %s", status, body)
	}
}

func TestAccessRequestRateLimit(t *testing.T) {
	setup(t)
	// One IP on purpose: the limiter allows 5/min per client IP on access-request.
	c := newClient(t)
	payload := map[string]string{
		"tenant_id": "00000000-0000-0000-0000-000000000001", "email": uniqueEmail("ratelimit"),
	}
	for i := 1; i <= 5; i++ {
		if status, body := c.do("POST", "/v1/portal/access-request", payload); status != http.StatusAccepted {
			t.Fatalf("request %d = %d, want 202: %s", i, status, body)
		}
	}
	if status, body := c.do("POST", "/v1/portal/access-request", payload); status != http.StatusTooManyRequests {
		t.Fatalf("request 6 = %d, want 429: %s", status, body)
	}
}

// ---------------------------------------------------------------- auth guard

func TestDashboardRequiresAuth(t *testing.T) {
	setup(t)
	c := newClient(t)
	for _, path := range []string{"/v1/me", "/v1/plans", "/v1/customers", "/v1/api-keys", "/v1/settings", "/v1/analytics/overview"} {
		if status, _ := c.do("GET", path, nil); status != http.StatusUnauthorized {
			t.Errorf("GET %s unauthenticated = %d, want 401", path, status)
		}
	}
	for _, path := range []string{"/v1/portal/me", "/v1/portal/subscription", "/v1/portal/invoices"} {
		if status, _ := c.do("GET", path, nil); status != http.StatusUnauthorized {
			t.Errorf("GET %s unauthenticated = %d, want 401", path, status)
		}
	}
}
