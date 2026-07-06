// Package middleware holds the gin middleware chain: correlation ids, recovery, CORS,
// Prometheus metrics, rate limiting, and the auth guards for dashboard and portal.
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/shamigam1/subba/internal/auth"
	"github.com/shamigam1/subba/internal/http/render"
	"github.com/shamigam1/subba/internal/store/db"
)

// Context keys for values set by the auth guards.
const (
	CtxTenantID   = "tenant_id"
	CtxCustomerID = "customer_id"

	CookieSession = "subba_session"
	CookiePortal  = "subba_portal"
)

var httpDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency by route.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "route", "status"},
)

type Middleware struct {
	sessions       *auth.Sessions
	admin          *pgxpool.Pool
	rdb            *redis.Client
	log            zerolog.Logger
	allowedOrigins map[string]bool
	secureCookies  bool
}

func New(sessions *auth.Sessions, admin *pgxpool.Pool, rdb *redis.Client, log zerolog.Logger, origins []string, secureCookies bool) *Middleware {
	set := make(map[string]bool, len(origins))
	for _, o := range origins {
		set[o] = true
	}
	return &Middleware{sessions: sessions, admin: admin, rdb: rdb, log: log, allowedOrigins: set, secureCookies: secureCookies}
}

// MetricsHandler exposes the Prometheus scrape endpoint.
func MetricsHandler() gin.HandlerFunc { return gin.WrapH(promhttp.Handler()) }

// RequestID assigns/propagates a correlation id on every request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(render.RequestIDKey, rid)
		c.Header("X-Request-Id", rid)
		c.Next()
	}
}

// Metrics records request latency labelled by route template (low cardinality).
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		httpDuration.WithLabelValues(c.Request.Method, route, http.StatusText(c.Writer.Status())).
			Observe(time.Since(start).Seconds())
	}
}

// Recovery converts panics into a clean 500 error envelope.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, _ any) {
		render.Err(c, http.StatusInternalServerError, "internal", "internal server error")
	})
}

// CORS allows the configured dashboard/portal origins with credentials (cookies).
func (m *Middleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && m.allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Request-Id")
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RateLimit caps requests per window keyed by keyFn (e.g. email or client IP).
func (m *Middleware) RateLimit(max int, window time.Duration, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		k := "rl:" + keyFn(c)
		n, err := m.rdb.Incr(c, k).Result()
		if err == nil && n == 1 {
			m.rdb.Expire(c, k, window)
		}
		if n > int64(max) {
			render.Err(c, http.StatusTooManyRequests, "rate_limited", "too many requests, try again later")
			return
		}
		c.Next()
	}
}

// RequireTenant authenticates a dashboard request via a Bearer API key or the session
// cookie, and sets the tenant id in context.
func (m *Middleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			// Try a session token first (cheap Redis lookup). Session ids are
			// 43-char base64url strings from auth.RandomToken — NOT UUIDs, so
			// don't gate this on uuid.Parse(token); only the stored subject
			// (the tenant id) is a UUID.
			if sub, err := m.sessions.Get(c, "tenant", token); err == nil {
				if tid, err := uuid.Parse(sub); err == nil {
					c.Set(CtxTenantID, tid)
					c.Next()
					return
				}
			}
			if k, err := db.New(m.admin).GetAPIKeyByHash(c, auth.HashToken(token)); err == nil {
				_ = db.New(m.admin).TouchAPIKey(c, k.ID)
				c.Set(CtxTenantID, k.TenantID)
				c.Next()
				return
			}
		}
		if sid, err := c.Cookie(CookieSession); err == nil {
			if sub, err := m.sessions.Get(c, "tenant", sid); err == nil {
				if tid, err := uuid.Parse(sub); err == nil {
					c.Set(CtxTenantID, tid)
					c.Next()
					return
				}
			}
		}
		render.Err(c, http.StatusUnauthorized, "unauthorized", "authentication required")
	}
}

// RequirePortal authenticates a customer-portal request via the portal cookie or Bearer token, and sets
// both tenant id and customer id in context.
func (m *Middleware) RequirePortal() gin.HandlerFunc {
	return func(c *gin.Context) {
		var sid string
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			sid = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		} else if cookie, err := c.Cookie(CookiePortal); err == nil {
			sid = cookie
		}

		if sid == "" {
			render.Err(c, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}
		sub, err := m.sessions.Get(c, "portal", sid)
		if err != nil {
			render.Err(c, http.StatusUnauthorized, "unauthorized", "session expired")
			return
		}
		parts := strings.SplitN(sub, ":", 2)
		if len(parts) != 2 {
			render.Err(c, http.StatusUnauthorized, "unauthorized", "invalid session")
			return
		}
		tid, err1 := uuid.Parse(parts[0])
		cid, err2 := uuid.Parse(parts[1])
		if err1 != nil || err2 != nil {
			render.Err(c, http.StatusUnauthorized, "unauthorized", "invalid session")
			return
		}
		c.Set(CtxTenantID, tid)
		c.Set(CtxCustomerID, cid)
		c.Next()
	}
}

// TenantID / CustomerID read the ids set by the guards.
func TenantID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(CtxTenantID)
	id, _ := v.(uuid.UUID)
	return id
}
func CustomerID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(CtxCustomerID)
	id, _ := v.(uuid.UUID)
	return id
}
