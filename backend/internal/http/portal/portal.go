// Package portal implements the end-user customer portal API: passwordless magic-link
// access, subscription view + cancel, invoices, and payment method (card + cardless).
package portal

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/shamigam1/subba/internal/auth"
	"github.com/shamigam1/subba/internal/config"
	"github.com/shamigam1/subba/internal/http/dto"
	"github.com/shamigam1/subba/internal/http/middleware"
	"github.com/shamigam1/subba/internal/http/render"
	"github.com/shamigam1/subba/internal/nomba"
	"github.com/shamigam1/subba/internal/notify"
	"github.com/shamigam1/subba/internal/store"
	"github.com/shamigam1/subba/internal/store/db"
)

type Handler struct {
	cfg      *config.Config
	log      zerolog.Logger
	pool     *pgxpool.Pool
	admin    *pgxpool.Pool
	sessions *auth.Sessions
	mailer   *notify.Mailer
	nomba    *nomba.Client
}

func New(cfg *config.Config, log zerolog.Logger, pool, admin *pgxpool.Pool, sessions *auth.Sessions, mailer *notify.Mailer, nc *nomba.Client) *Handler {
	return &Handler{cfg: cfg, log: log, pool: pool, admin: admin, sessions: sessions, mailer: mailer, nomba: nc}
}

func (h *Handler) tenantQ(c *gin.Context, tenantID uuid.UUID, fn func(*db.Queries) error) error {
	return store.WithTenant(c.Request.Context(), h.pool, tenantID.String(), func(tx pgx.Tx) error {
		return fn(db.New(tx))
	})
}

// Portal sessions are short-lived (consumer-facing, passwordless).
const portalSessionTTL = 30 * time.Minute

func (h *Handler) setPortalCookie(c *gin.Context, sid string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.CookiePortal, sid, int(portalSessionTTL.Seconds()), "/", "", h.cfg.AppEnv != "development", true)
}

// AccessRequest emails a magic link. Enumeration-safe: always 202 regardless of whether
// the customer exists.
func (h *Handler) AccessRequest(c *gin.Context) {
	var req struct {
		TenantID uuid.UUID `json:"tenant_id"`
		Email    string    `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TenantID == uuid.Nil || req.Email == "" {
		render.Err(c, http.StatusBadRequest, "bad_request", "tenant_id and email are required")
		return
	}
	const accepted = "if an account exists, a secure link has been sent"

	var cust db.Customer
	err := h.tenantQ(c, req.TenantID, func(q *db.Queries) (e error) {
		cust, e = q.GetCustomerByEmail(c.Request.Context(), req.Email)
		return
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.JSON(c, http.StatusAccepted, gin.H{"message": accepted}) // no such customer; reveal nothing
		return
	}
	if err != nil {
		render.JSON(c, http.StatusAccepted, gin.H{"message": accepted})
		return
	}

	token := auth.RandomToken()
	if e := h.tenantQ(c, req.TenantID, func(q *db.Queries) error {
		_, err := q.CreatePortalToken(c.Request.Context(), db.CreatePortalTokenParams{
			TenantID: req.TenantID, CustomerID: cust.ID, TokenHash: auth.HashToken(token), ExpiresAt: time.Now().Add(15 * time.Minute),
		})
		return err
	}); e != nil {
		render.JSON(c, http.StatusAccepted, gin.H{"message": accepted})
		return
	}

	tenantName := "your provider"
	if t, e := db.New(h.admin).GetTenantByID(c.Request.Context(), req.TenantID); e == nil {
		tenantName = t.Name
	}
	link := h.cfg.PortalBaseURL + "/access?token=" + token
	if e := h.mailer.SendMagicLink(c.Request.Context(), req.Email, tenantName, link); e != nil {
		h.log.Error().Err(e).Msg("failed to send magic link")
	}
	render.JSON(c, http.StatusAccepted, gin.H{"message": accepted})
}

// Session exchanges a magic-link token for a portal session cookie.
func (h *Handler) Session(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" {
		render.Err(c, http.StatusBadRequest, "bad_request", "token is required")
		return
	}
	hash := auth.HashToken(req.Token)
	// Single-use + unexpired check is atomic in ConsumePortalToken.
	rows, err := db.New(h.admin).ConsumePortalToken(c.Request.Context(), hash)
	if err != nil || rows == 0 {
		render.Err(c, http.StatusUnauthorized, "unauthorized", "link is invalid or expired")
		return
	}
	tok, err := db.New(h.admin).GetPortalTokenByHash(c.Request.Context(), hash)
	if err != nil {
		render.Err(c, http.StatusUnauthorized, "unauthorized", "link is invalid")
		return
	}
	sid, err := h.sessions.Create(c.Request.Context(), "portal", tok.TenantID.String()+":"+tok.CustomerID.String(), portalSessionTTL)
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not start session")
		return
	}
	h.setPortalCookie(c, sid)
	out := h.context(c, tok.TenantID, tok.CustomerID)
	out.Token = sid
	render.JSON(c, http.StatusOK, out)
}

func (h *Handler) Me(c *gin.Context) {
	render.JSON(c, http.StatusOK, h.context(c, middleware.TenantID(c), middleware.CustomerID(c)))
}

// context assembles the portal shell payload (customer + tenant branding).
func (h *Handler) context(c *gin.Context, tenantID, customerID uuid.UUID) dto.PortalContext {
	var ctx dto.PortalContext
	if cust, err := db.New(h.admin).GetCustomer(c.Request.Context(), customerID); err == nil {
		ctx.Customer = dto.FromCustomer(cust, nil, nil)
	}
	if t, err := db.New(h.admin).GetTenantByID(c.Request.Context(), tenantID); err == nil {
		ctx.TenantBranding.TenantName = t.Name
	}
	return ctx
}

func (h *Handler) Logout(c *gin.Context) {
	if sid, err := c.Cookie(middleware.CookiePortal); err == nil {
		_ = h.sessions.Delete(c.Request.Context(), "portal", sid)
	}
	c.SetCookie(middleware.CookiePortal, "", -1, "/", "", h.cfg.AppEnv != "development", true)
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetSubscription(c *gin.Context) {
	customerID := middleware.CustomerID(c)
	var sub db.Subscription
	var plan db.Plan
	err := h.tenantQ(c, middleware.TenantID(c), func(q *db.Queries) error {
		var e error
		sub, e = q.GetSubscriptionByCustomer(c.Request.Context(), customerID)
		if e != nil {
			return e
		}
		plan, e = q.GetPlan(c.Request.Context(), sub.PlanID)
		return e
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Err(c, http.StatusNotFound, "not_found", "no subscription found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not load subscription")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromSubscription(sub, &plan))
}

func (h *Handler) CancelSubscription(c *gin.Context) {
	req := struct {
		AtPeriodEnd *bool `json:"at_period_end"`
	}{}
	_ = c.ShouldBindJSON(&req)
	atEnd := true
	if req.AtPeriodEnd != nil {
		atEnd = *req.AtPeriodEnd
	}
	customerID := middleware.CustomerID(c)
	var sub db.Subscription
	err := h.tenantQ(c, middleware.TenantID(c), func(q *db.Queries) error {
		current, e := q.GetSubscriptionByCustomer(c.Request.Context(), customerID)
		if e != nil {
			return e
		}
		sub, e = q.CancelSubscription(c.Request.Context(), db.CancelSubscriptionParams{ID: current.ID, AtPeriodEnd: atEnd})
		return e
	})
	if errors.Is(err, pgx.ErrNoRows) {
		render.Err(c, http.StatusNotFound, "not_found", "no subscription found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not cancel subscription")
		return
	}
	render.JSON(c, http.StatusOK, dto.FromSubscription(sub, nil))
}

func (h *Handler) ListInvoices(c *gin.Context) {
	customerID := middleware.CustomerID(c)
	var invs []db.Invoice
	if err := h.tenantQ(c, middleware.TenantID(c), func(q *db.Queries) (e error) {
		invs, e = q.ListInvoicesByCustomer(c.Request.Context(), customerID)
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

func (h *Handler) GetInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	customerID := middleware.CustomerID(c)
	var inv db.Invoice
	var items []db.InvoiceItem
	err = h.tenantQ(c, middleware.TenantID(c), func(q *db.Queries) error {
		var e error
		inv, e = q.GetInvoice(c.Request.Context(), id)
		if e != nil {
			return e
		}
		items, e = q.ListInvoiceItems(c.Request.Context(), id)
		return e
	})
	// Scope to the signed-in customer: don't leak another customer's invoice.
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && inv.CustomerID != customerID) {
		render.Err(c, http.StatusNotFound, "not_found", "invoice not found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not load invoice")
		return
	}
	detail := dto.InvoiceDetail{Invoice: dto.FromInvoice(inv)}
	for _, it := range items {
		detail.Items = append(detail.Items, dto.FromInvoiceItem(it))
	}
	render.JSON(c, http.StatusOK, detail)
}

func (h *Handler) CreateCheckoutLink(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		render.Err(c, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	customerID := middleware.CustomerID(c)
	tenantID := middleware.TenantID(c)

	var inv db.Invoice
	var cust db.Customer
	var tenant db.Tenant
	err = h.tenantQ(c, tenantID, func(q *db.Queries) error {
		var e error
		inv, e = q.GetInvoice(c.Request.Context(), id)
		if e != nil {
			return e
		}
		cust, e = q.GetCustomer(c.Request.Context(), customerID)
		if e != nil {
			return e
		}
		tenant, e = q.GetTenantByID(c.Request.Context(), tenantID)
		return e
	})
	
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && inv.CustomerID != customerID) {
		render.Err(c, http.StatusNotFound, "not_found", "invoice not found")
		return
	}
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not load invoice")
		return
	}

	if inv.Status == "paid" {
		render.Err(c, http.StatusBadRequest, "bad_request", "invoice is already paid")
		return
	}

	var tenantAccountID string
	if tenant.NombaAccountID != nil {
		tenantAccountID = *tenant.NombaAccountID
	}

	req := nomba.CreateCheckoutOrderRequest{
		Order: nomba.CheckoutOrder{
			OrderReference: "inv:" + inv.ID.String(),
			CustomerID:     customerID.String(),
			CallbackURL:    h.cfg.PortalBaseURL + "/invoices/" + inv.ID.String(),
			CustomerEmail:  cust.Email,
			// Integer math only — never float on money (kobo → "naira.kobo").
			Amount:         fmt.Sprintf("%d.%02d", inv.Amount/100, inv.Amount%100),
			Currency:       "NGN",
			AccountID:      tenantAccountID,
		},
		TokenizeCard: true,
	}

	resp, err := h.nomba.CreateCheckoutOrder(c.Request.Context(), req)
	if err != nil {
		h.log.Error().Err(err).Msg("nomba checkout order failed")
		render.Err(c, http.StatusInternalServerError, "nomba_error", "failed to create checkout link")
		return
	}

	render.JSON(c, http.StatusOK, gin.H{
		"checkoutLink": resp.Data.CheckoutLink,
	})
}

// PaymentMethod returns the card on file plus the cardless virtual account. The Nomba
// card metadata and live virtual account are filled in once Track A's Nomba client lands.
func (h *Handler) PaymentMethod(c *gin.Context) {
	cust, err := h.getCustomer(c)
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not load payment method")
		return
	}
	render.JSON(c, http.StatusOK, gin.H{
		"card":            cardOnFile(cust),
		"virtual_account": virtualAccount(cust),
	})
}

func (h *Handler) SaveCard(c *gin.Context) {
	var req struct {
		NombaTokenKey string `json:"nomba_token_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.NombaTokenKey == "" {
		render.Err(c, http.StatusBadRequest, "bad_request", "nomba_token_key is required")
		return
	}
	customerID := middleware.CustomerID(c)
	var cust db.Customer
	if err := h.tenantQ(c, middleware.TenantID(c), func(q *db.Queries) (e error) {
		cust, e = q.SetCustomerCardToken(c.Request.Context(), db.SetCustomerCardTokenParams{ID: customerID, NombaTokenKey: &req.NombaTokenKey})
		return
	}); err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not save card")
		return
	}
	render.JSON(c, http.StatusOK, cardOnFile(cust))
}

func (h *Handler) VirtualAccount(c *gin.Context) {
	cust, err := h.getCustomer(c)
	if err != nil {
		render.Err(c, http.StatusInternalServerError, "internal", "could not load virtual account")
		return
	}
	render.JSON(c, http.StatusOK, virtualAccount(cust))
}

func (h *Handler) getCustomer(c *gin.Context) (db.Customer, error) {
	customerID := middleware.CustomerID(c)
	var cust db.Customer
	err := h.tenantQ(c, middleware.TenantID(c), func(q *db.Queries) (e error) {
		cust, e = q.GetCustomer(c.Request.Context(), customerID)
		return
	})
	return cust, err
}

// cardOnFile/virtualAccount are stubs until Track A's Nomba client provides real
// card metadata and provisions virtual accounts.
func cardOnFile(cust db.Customer) any {
	if cust.NombaTokenKey == nil || *cust.NombaTokenKey == "" {
		return nil
	}
	return gin.H{"brand": "card", "last4": "____", "exp_month": 0, "exp_year": 0, "expired": false}
}

func virtualAccount(cust db.Customer) gin.H {
	acct := ""
	bankName := "Nomba (pending provisioning)"
	if cust.NombaVirtualAccount != nil && *cust.NombaVirtualAccount != "" {
		acct = *cust.NombaVirtualAccount
		bankName = "Nomba"
	}
	name := cust.Email
	if cust.Name != nil && *cust.Name != "" {
		name = *cust.Name
	}
	return gin.H{
		"bank_name":      bankName,
		"account_number": acct,
		"account_name":   name,
		"amount_due":     gin.H{"amount_minor": 0, "currency": "NGN"},
	}
}
