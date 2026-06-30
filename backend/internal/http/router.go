// Package httpapi assembles the gin router: middleware chain, health/metrics probes,
// and the dashboard + portal route tables.
package httpapi

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/shamigam1/subba/internal/auth"
	"github.com/shamigam1/subba/internal/config"
	"github.com/shamigam1/subba/internal/http/dashboard"
	"github.com/shamigam1/subba/internal/http/middleware"
	"github.com/shamigam1/subba/internal/http/portal"
	"github.com/shamigam1/subba/internal/http/render"
	"github.com/shamigam1/subba/internal/notify"
	"github.com/shamigam1/subba/internal/observability"
	"github.com/shamigam1/subba/internal/platform"
)

// NewRouter builds the fully-wired HTTP handler.
func NewRouter(cfg *config.Config, log zerolog.Logger, plat *platform.Platform) http.Handler {
	if cfg.AppEnv != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Expose connection-pool saturation to Prometheus (ignore re-registration in tests).
	_ = prometheus.Register(observability.NewPoolCollector(map[string]*pgxpool.Pool{
		"tenant": plat.DB, "admin": plat.AdminDB,
	}))

	sessions := auth.NewSessions(plat.Redis)
	mailer := notify.NewMailer(cfg.ResendAPIKey, cfg.EmailFromName, cfg.EmailFromEmail, log)
	mw := middleware.New(sessions, plat.AdminDB, plat.Redis, log, allowedOrigins(cfg), cfg.AppEnv != "development")

	dash := dashboard.New(cfg, log, plat.DB, plat.AdminDB, sessions)
	prt := portal.New(cfg, log, plat.DB, plat.AdminDB, sessions, mailer)

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recovery(), middleware.Metrics(), mw.CORS())

	// Probes & metrics.
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := plat.Ready(ctx); err != nil {
			render.Err(c, http.StatusServiceUnavailable, "unavailable", err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/metrics", middleware.MetricsHandler())

	v1 := r.Group("/v1")

	// --- public ---
	clientIP := func(c *gin.Context) string { return c.ClientIP() }
	v1.POST("/auth/signup", mw.RateLimit(10, time.Minute, clientIP), dash.Signup)
	v1.POST("/auth/login", mw.RateLimit(10, time.Minute, clientIP), dash.Login)
	v1.POST("/auth/logout", dash.Logout)
	v1.POST("/portal/access-request", mw.RateLimit(5, time.Minute, clientIP), prt.AccessRequest)
	v1.POST("/portal/session", mw.RateLimit(20, time.Minute, clientIP), prt.Session)
	v1.POST("/portal/logout", prt.Logout)

	// --- dashboard (tenant auth) ---
	d := v1.Group("", mw.RequireTenant())
	d.GET("/me", dash.Me)
	d.GET("/plans", dash.ListPlans)
	d.POST("/plans", dash.CreatePlan)
	d.GET("/plans/:id", dash.GetPlan)
	d.PATCH("/plans/:id", dash.UpdatePlan)
	d.DELETE("/plans/:id", dash.DeletePlan)
	d.GET("/customers", dash.ListCustomers)
	d.POST("/customers", dash.CreateCustomer)
	d.GET("/customers/:id", dash.GetCustomer)
	d.PATCH("/customers/:id", dash.UpdateCustomer)
	d.GET("/customers/:id/invoices", dash.ListCustomerInvoices)
	d.POST("/customers/:id/portal-link", dash.CreatePortalLink)
	d.POST("/subscriptions", dash.CreateSubscription)
	d.GET("/subscriptions/:id", dash.GetSubscription)
	d.POST("/subscriptions/:id/cancel", dash.CancelSubscription)
	d.GET("/api-keys", dash.ListAPIKeys)
	d.POST("/api-keys", dash.CreateAPIKey)
	d.DELETE("/api-keys/:id", dash.RevokeAPIKey)
	d.GET("/analytics/overview", dash.AnalyticsOverview)
	d.GET("/settings", dash.GetSettings)
	d.PATCH("/settings", dash.UpdateSettings)

	// --- portal (customer auth) ---
	p := v1.Group("/portal", mw.RequirePortal())
	p.GET("/me", prt.Me)
	p.GET("/subscription", prt.GetSubscription)
	p.POST("/subscription/cancel", prt.CancelSubscription)
	p.GET("/invoices", prt.ListInvoices)
	p.GET("/invoices/:id", prt.GetInvoice)
	p.GET("/payment-method", prt.PaymentMethod)
	p.POST("/payment-method/card", prt.SaveCard)
	p.GET("/virtual-account", prt.VirtualAccount)

	return r
}

// allowedOrigins derives CORS origins from the configured portal URL plus the local
// dev frontend.
func allowedOrigins(cfg *config.Config) []string {
	origins := []string{"http://localhost:3000"}
	if u, err := url.Parse(cfg.PortalBaseURL); err == nil && u.Scheme != "" && u.Host != "" {
		origins = append(origins, u.Scheme+"://"+u.Host)
	}
	return origins
}
