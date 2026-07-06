// Package dashboard implements the tenant-facing (developer) API: auth, plans,
// customers, subscriptions, API keys, analytics, settings, and portal-link minting.
package dashboard

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/shamigam1/subba/internal/auth"
	"github.com/shamigam1/subba/internal/config"
	"github.com/shamigam1/subba/internal/crypto"
	"github.com/shamigam1/subba/internal/http/dto"
	"github.com/shamigam1/subba/internal/http/middleware"
	"github.com/shamigam1/subba/internal/http/render"
	"github.com/shamigam1/subba/internal/nomba"
	"github.com/shamigam1/subba/internal/store"
	"github.com/shamigam1/subba/internal/store/db"
)

type Handler struct {
	cfg      *config.Config
	log      zerolog.Logger
	pool     *pgxpool.Pool // tenant-scoped (RLS)
	admin    *pgxpool.Pool // privileged (signup/auth)
	sessions *auth.Sessions
	nomba    *nomba.Client
}

func New(cfg *config.Config, log zerolog.Logger, pool, admin *pgxpool.Pool, sessions *auth.Sessions, nc *nomba.Client) *Handler {
	return &Handler{cfg: cfg, log: log, pool: pool, admin: admin, sessions: sessions, nomba: nc}
}

// tenantQ runs fn inside the calling tenant's RLS-scoped transaction.
func (h *Handler) tenantQ(c *gin.Context, fn func(*db.Queries) error) error {
	return store.WithTenant(c.Request.Context(), h.pool, middleware.TenantID(c).String(), func(tx pgx.Tx) error {
		return fn(db.New(tx))
	})
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

const sessionTTL = 24 * time.Hour

func (h *Handler) setSession(c *gin.Context, sid string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.CookieSession, sid, int(sessionTTL.Seconds()), "/", "", h.cfg.AppEnv != "development", true)
}

// ----------------------------------------------------------------- auth

func (h *Handler) Signup(c *gin.Context) {
	var req struct {
		Name, Email, Password string
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Name == "" || req.Email == "" || len(req.Password) < 8 {
		render.ValidationErr(c, map[string]string{"password": "name, email required; password min 8 chars"})
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not hash password")
		return
	}
	t, err := db.New(h.admin).CreateTenant(c.Request.Context(), db.CreateTenantParams{Name: req.Name, Email: req.Email, PasswordHash: &hash})
	if err != nil {
		if isUniqueViolation(err) {
			render.Err(c, http.StatusConflict, "conflict", "an account with that email already exists")
			return
		}
		render.Err(c, http.StatusInternalServerError, "internal", "could not create account")
		return
	}
	sid, _ := h.sessions.Create(c.Request.Context(), "tenant", t.ID.String(), sessionTTL)
	h.setSession(c, sid)
	render.JSON(c, http.StatusCreated, dto.FromTenant(t))
}

func (h *Handler) Login(c *gin.Context) {
	var req struct{ Email, Password string }
	if err := c.ShouldBindJSON(&req); err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	t, err := db.New(h.admin).GetTenantByEmail(c.Request.Context(), req.Email)
	if err != nil {
		render.Err(c, http.StatusUnauthorized, "unauthorized", "invalid email or password")
		return
	}
	ok := false
	if t.PasswordHash != nil {
		ok, _ = auth.VerifyPassword(*t.PasswordHash, req.Password)
	}
	if !ok {
		render.Err(c, http.StatusUnauthorized, "unauthorized", "invalid email or password")
		return
	}
	sid, _ := h.sessions.Create(c.Request.Context(), "tenant", t.ID.String(), sessionTTL)
	h.setSession(c, sid)
	render.JSON(c, http.StatusOK, dto.FromTenant(t))
}

func (h *Handler) Logout(c *gin.Context) {
	if sid, err := c.Cookie(middleware.CookieSession); err == nil {
		_ = h.sessions.Delete(c.Request.Context(), "tenant", sid)
	}
	c.SetCookie(middleware.CookieSession, "", -1, "/", "", h.cfg.AppEnv != "development", true)
	c.Status(http.StatusNoContent)
}

func (h *Handler) Me(c *gin.Context) {
	t, err := db.New(h.admin).GetTenantByID(c.Request.Context(), middleware.TenantID(c))
	if err != nil {
		render.Err(c, http.StatusUnauthorized, "unauthorized", "tenant not found")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromTenant(t))
}

// ----------------------------------------------------------------- plans

func (h *Handler) ListPlans(c *gin.Context) {
	includeDeleted := c.Query("include_deleted") == "true"
	var plans []db.Plan
	if err := h.tenantQ(c, func(q *db.Queries) (err error) {
		plans, err = q.ListPlans(c.Request.Context(), includeDeleted)
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not list plans")
		return
	}
	out := make([]dto.Plan, 0, len(plans))
	for _, p := range plans {
		out = append(out, dto.FromPlan(p))
	}
	render.JSON(c, http.StatusOK, out)
}

func (h *Handler) CreatePlan(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		AmountMinor int64  `json:"amount_minor"`
		Currency    string `json:"currency"`
		Interval    string `json:"interval"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Name == "" || req.AmountMinor < 0 || (req.Interval != "month" && req.Interval != "year") {
		render.ValidationErr(c, map[string]string{"interval": "name, non-negative amount_minor, interval in [month,year] required"})
		return
	}
	if req.Currency == "" {
		req.Currency = "NGN"
	}
	var plan db.Plan
	if err := h.tenantQ(c, func(q *db.Queries) (err error) {
		plan, err = q.CreatePlan(c.Request.Context(), db.CreatePlanParams{
			TenantID: middleware.TenantID(c), Name: req.Name, Amount: req.AmountMinor,
			Currency: req.Currency, Interval: req.Interval,
		})
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not create plan")
		return
	}
	render.JSON(c, http.StatusCreated, dto.FromPlan(plan))
}

func (h *Handler) GetPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var plan db.Plan
	err = h.tenantQ(c, func(q *db.Queries) (e error) {
		plan, e = q.GetPlan(c.Request.Context(), id)
		return
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Err(c, http.StatusNotFound, "not_found", "plan not found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not get plan")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromPlan(plan))
}

func (h *Handler) UpdatePlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var req struct {
		Name        *string `json:"name"`
		AmountMinor *int64  `json:"amount_minor"`
		Currency    *string `json:"currency"`
		Interval    *string `json:"interval"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	var plan db.Plan
	err = h.tenantQ(c, func(q *db.Queries) (e error) {
		plan, e = q.UpdatePlan(c.Request.Context(), db.UpdatePlanParams{
			ID: id, Name: req.Name, Amount: req.AmountMinor, Currency: req.Currency, Interval: req.Interval,
		})
		return
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Err(c, http.StatusNotFound, "not_found", "plan not found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not update plan")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromPlan(plan))
}

func (h *Handler) DeletePlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var rows int64
	if err := h.tenantQ(c, func(q *db.Queries) (e error) {
		rows, e = q.SoftDeletePlan(c.Request.Context(), id)
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not delete plan")
		return
	}
	if rows == 0 {
		render.Err(c, http.StatusNotFound, "not_found", "plan not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// ----------------------------------------------------------------- customers

func (h *Handler) ListCustomers(c *gin.Context) {
	var q *string
	if s := c.Query("q"); s != "" {
		q = &s
	}
	var customers []db.Customer
	if err := h.tenantQ(c, func(qq *db.Queries) (err error) {
		customers, err = qq.ListCustomers(c.Request.Context(), db.ListCustomersParams{Q: q, Lim: 100})
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not list customers")
		return
	}
	data := make([]dto.Customer, 0, len(customers))
	for _, cu := range customers {
		data = append(data, dto.FromCustomer(cu))
	}
	render.JSON(c, http.StatusOK, gin.H{"data": data, "next_cursor": nil})
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	var req struct {
		Name  *string `json:"name"`
		Email string  `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		render.ValidationErr(c, map[string]string{"email": "email is required"})
		return
	}
	var cust db.Customer
	err := h.tenantQ(c, func(q *db.Queries) (e error) {
		cust, e = q.CreateCustomer(c.Request.Context(), db.CreateCustomerParams{
			TenantID: middleware.TenantID(c), Name: req.Name, Email: req.Email,
		})
		return
	})
	if isUniqueViolation(err) {
		render.Err(c, http.StatusConflict, "conflict", "a customer with that email already exists")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not create customer")
		return
	}
	// Synchronously provision the Nomba Virtual Account.
	// As per the Slack advisory, we pass the sub-account ID into the payload
	// and tag it with our custom `accountRef` ({tenantID}:{customerID}).
	tID := middleware.TenantID(c)
	accountRef := tID.String() + ":" + cust.ID.String()
	name := "Subba Customer"
	if req.Name != nil && *req.Name != "" {
		name = *req.Name
	}
	
	// Skip virtual-account provisioning when no Nomba client is configured (e.g. tests).
	if h.nomba == nil {
		render.JSON(c, http.StatusCreated, dto.FromCustomer(cust))
		return
	}
	// Create the virtual account on Nomba
	vaRes, err := h.nomba.CreateVirtualAccount(c.Request.Context(), h.cfg.NombaSubAccountID, nomba.CreateVirtualAccountRequest{
		AccountRef:  accountRef,
		AccountName: name,
	})
	if err != nil {
		h.log.Error().Err(err).Msg("failed to provision nomba virtual account")
		// We still return 201 Created because the customer exists, but the VA is missing.
		// A background job could retry this in a real system.
	} else if vaRes != nil {
		// Update the customer with the provisioned NUBAN.
		_ = h.tenantQ(c, func(q *db.Queries) error {
			cust, _ = q.UpdateCustomer(c.Request.Context(), db.UpdateCustomerParams{
				ID: cust.ID, Name: req.Name, Email: &req.Email,
			})
			// Since our SQLC query doesn't currently allow directly updating `nomba_virtual_account`,
			// we must execute a raw update or add a query. Let's do a direct pool query.
			_, e := h.pool.Exec(c.Request.Context(), "UPDATE customers SET nomba_virtual_account = $1 WHERE id = $2 AND tenant_id = $3", vaRes.Data.AccountNumber, cust.ID, tID)
			if e == nil {
				cust.NombaVirtualAccount = &vaRes.Data.AccountNumber
			}
			return nil
		})
	}

	render.JSON(c, http.StatusCreated, dto.FromCustomer(cust))
}

func (h *Handler) GetCustomer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var cust db.Customer
	err = h.tenantQ(c, func(q *db.Queries) (e error) {
		cust, e = q.GetCustomer(c.Request.Context(), id)
		return
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Err(c, http.StatusNotFound, "not_found", "customer not found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not get customer")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromCustomer(cust))
}

func (h *Handler) UpdateCustomer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var req struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	var cust db.Customer
	err = h.tenantQ(c, func(q *db.Queries) (e error) {
		cust, e = q.UpdateCustomer(c.Request.Context(), db.UpdateCustomerParams{ID: id, Name: req.Name, Email: req.Email})
		return
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Err(c, http.StatusNotFound, "not_found", "customer not found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not update customer")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromCustomer(cust))
}

func (h *Handler) ListCustomerInvoices(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var invs []db.Invoice
	if err := h.tenantQ(c, func(q *db.Queries) (e error) {
		invs, e = q.ListInvoicesByCustomer(c.Request.Context(), id)
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not list invoices")
		return
	}
	out := make([]dto.Invoice, 0, len(invs))
	for _, i := range invs {
		out = append(out, dto.FromInvoice(i))
	}
	render.JSON(c, http.StatusOK, out)
}

// ----------------------------------------------------------------- subscriptions

func (h *Handler) CreateSubscription(c *gin.Context) {
	var req struct {
		CustomerID uuid.UUID `json:"customer_id"`
		PlanID     uuid.UUID `json:"plan_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CustomerID == uuid.Nil || req.PlanID == uuid.Nil {
		render.ValidationErr(c, map[string]string{"customer_id": "customer_id and plan_id are required"})
		return
	}
	var sub db.Subscription
	var plan db.Plan
	err := h.tenantQ(c, func(q *db.Queries) error {
		var e error
		sub, e = q.CreateSubscription(c.Request.Context(), db.CreateSubscriptionParams{
			TenantID: middleware.TenantID(c), CustomerID: req.CustomerID, PlanID: req.PlanID,
		})
		if e != nil {
			return e
		}
		plan, e = q.GetPlan(c.Request.Context(), req.PlanID)
		return e
	})
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not create subscription")
		return
	}
	render.JSON(c, http.StatusCreated, dto.FromSubscription(sub, &plan))
}

func (h *Handler) GetSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var sub db.Subscription
	var plan db.Plan
	err = h.tenantQ(c, func(q *db.Queries) error {
		var e error
		sub, e = q.GetSubscription(c.Request.Context(), id)
		if e != nil {
			return e
		}
		plan, e = q.GetPlan(c.Request.Context(), sub.PlanID)
		return e
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Err(c, http.StatusNotFound, "not_found", "subscription not found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not get subscription")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromSubscription(sub, &plan))
}

func (h *Handler) CancelSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	req := struct {
		AtPeriodEnd *bool `json:"at_period_end"`
	}{}
	_ = c.ShouldBindJSON(&req)
	atEnd := true
	if req.AtPeriodEnd != nil {
		atEnd = *req.AtPeriodEnd
	}
	var sub db.Subscription
	err = h.tenantQ(c, func(q *db.Queries) (e error) {
		sub, e = q.CancelSubscription(c.Request.Context(), db.CancelSubscriptionParams{ID: id, AtPeriodEnd: atEnd})
		return
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Err(c, http.StatusNotFound, "not_found", "subscription not found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not cancel subscription")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromSubscription(sub, nil))
}

// ----------------------------------------------------------------- api keys

func (h *Handler) ListAPIKeys(c *gin.Context) {
	var keys []db.ApiKey
	if err := h.tenantQ(c, func(q *db.Queries) (e error) {
		keys, e = q.ListAPIKeys(c.Request.Context())
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not list keys")
		return
	}
	out := make([]dto.APIKey, 0, len(keys))
	for _, k := range keys {
		out = append(out, dto.FromAPIKey(k))
	}
	render.JSON(c, http.StatusOK, out)
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	var req struct {
		Name *string `json:"name"`
	}
	_ = c.ShouldBindJSON(&req)

	env := "test"
	if h.cfg.AppEnv != "development" {
		env = "live"
	}
	secret := "sk_" + env + "_" + auth.RandomToken()
	prefix := secret[:15]
	var key db.ApiKey
	if err := h.tenantQ(c, func(q *db.Queries) (e error) {
		key, e = q.CreateAPIKey(c.Request.Context(), db.CreateAPIKeyParams{
			TenantID: middleware.TenantID(c), Name: req.Name, KeyHash: auth.HashToken(secret), KeyPrefix: prefix,
		})
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not create key")
		return
	}
	render.JSON(c, http.StatusCreated, struct {
		dto.APIKey
		Secret string `json:"secret"`
	}{APIKey: dto.FromAPIKey(key), Secret: secret})
}

func (h *Handler) RevokeAPIKey(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var rows int64
	if err := h.tenantQ(c, func(q *db.Queries) (e error) {
		rows, e = q.RevokeAPIKey(c.Request.Context(), id)
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not revoke key")
		return
	}
	if rows == 0 {
		render.Err(c, http.StatusNotFound, "not_found", "key not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// ----------------------------------------------------------------- analytics & settings

func (h *Handler) AnalyticsOverview(c *gin.Context) {
	var mrr, active, payments, failed int64
	var series []db.RevenueSeriesRow
	if err := h.tenantQ(c, func(q *db.Queries) error {
		var e error
		if mrr, e = q.SumMRR(c.Request.Context()); e != nil {
			return e
		}
		if active, e = q.CountActiveSubscriptions(c.Request.Context()); e != nil {
			return e
		}
		if payments, e = q.CountPaymentsToday(c.Request.Context()); e != nil {
			return e
		}
		if failed, e = q.CountFailedInvoices(c.Request.Context()); e != nil {
			return e
		}
		series, e = q.RevenueSeries(c.Request.Context())
		return e
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not compute analytics")
		return
	}
	type point struct {
		Date        string `json:"date"`
		AmountMinor int64  `json:"amount_minor"`
	}
	pts := make([]point, 0, len(series))
	for _, s := range series {
		pts = append(pts, point{Date: s.Day.Time.Format("2006-01-02"), AmountMinor: s.Amount})
	}
	render.JSON(c, http.StatusOK, gin.H{
		"mrr":                  dto.Money{AmountMinor: mrr, Currency: "NGN"},
		"active_subscriptions": active,
		"payments_today":       payments,
		"failed_payments":      failed,
		"dlq_depth":            0, // wired to the engine's metrics in Phase 5
		"revenue_series":       pts,
	})
}

func (h *Handler) GetSettings(c *gin.Context) {
	t, err := db.New(h.admin).GetTenantByID(c.Request.Context(), middleware.TenantID(c))
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not load settings")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromSettings(t))
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var req struct {
		NombaAccountID    *string `json:"nomba_account_id"`
		NombaSubaccountID *string `json:"nomba_subaccount_id"`
		NombaClientID     *string `json:"nomba_client_id"`
		NombaClientSecret *string `json:"nomba_client_secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.NombaClientSecret != nil {
		enc, err := crypto.EncryptSecret(h.cfg.MasterEncryptionKey, *req.NombaClientSecret)
		if err != nil {
			render.Err(c, http.StatusInternalServerError, "internal", "could not secure settings")
			return
		}
		req.NombaClientSecret = &enc
	}
	var t db.Tenant
	if err := h.tenantQ(c, func(q *db.Queries) (e error) {
		t, e = q.UpdateTenantSettings(c.Request.Context(), db.UpdateTenantSettingsParams{
			ID:                middleware.TenantID(c),
			NombaAccountID:    req.NombaAccountID,
			NombaSubaccountID: req.NombaSubaccountID,
			NombaClientID:     req.NombaClientID,
			NombaClientSecret: req.NombaClientSecret,
		})
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not update settings")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromSettings(t))
}

// ----------------------------------------------------------------- portal link

func (h *Handler) CreatePortalLink(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	token := auth.RandomToken()
	expires := time.Now().Add(15 * time.Minute)
	err = h.tenantQ(c, func(q *db.Queries) error {
		if _, e := q.GetCustomer(c.Request.Context(), id); e != nil {
			return e
		}
		_, e := q.CreatePortalToken(c.Request.Context(), db.CreatePortalTokenParams{
			TenantID: middleware.TenantID(c), CustomerID: id, TokenHash: auth.HashToken(token), ExpiresAt: expires,
		})
		return e
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Err(c, http.StatusNotFound, "not_found", "customer not found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not create portal link")
		return
	}
	render.JSON(c, http.StatusCreated, gin.H{
		"url":        h.cfg.PortalBaseURL + "/access?token=" + token,
		"token":      token,
		"expires_at": expires,
	})
}
