// Package dto defines the JSON shapes returned by the API and maps sqlc models onto
// them. Keeping this explicit (rather than returning db models directly) lets the wire
// format stay stable and lets us mask secrets and rename fields (e.g. amount_minor).
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shamigam1/subba/internal/store/db"
)

func tsPtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		v := t.Time
		return &v
	}
	return nil
}

func uuidPtr(u pgtype.UUID) *string {
	if u.Valid {
		s := uuid.UUID(u.Bytes).String()
		return &s
	}
	return nil
}

type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Token     string    `json:"token,omitempty"`
}

func FromTenant(t db.Tenant) Tenant {
	return Tenant{ID: t.ID, Name: t.Name, Email: t.Email, CreatedAt: t.CreatedAt}
}

type Plan struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	AmountMinor int64      `json:"amount_minor"`
	Currency    string     `json:"currency"`
	Interval    string     `json:"interval"`
	DeletedAt   *time.Time `json:"deleted_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func FromPlan(p db.Plan) Plan {
	return Plan{
		ID: p.ID, Name: p.Name, AmountMinor: p.Amount, Currency: p.Currency,
		Interval: p.Interval, DeletedAt: tsPtr(p.DeletedAt), CreatedAt: p.CreatedAt,
	}
}

type Customer struct {
	ID            uuid.UUID `json:"id"`
	Name          *string   `json:"name"`
	Email         string    `json:"email"`
	HasCardOnFile bool      `json:"has_card_on_file"`
	CreatedAt     time.Time `json:"created_at"`
}

func FromCustomer(c db.Customer) Customer {
	return Customer{
		ID: c.ID, Name: c.Name, Email: c.Email,
		HasCardOnFile: c.NombaTokenKey != nil && *c.NombaTokenKey != "",
		CreatedAt:     c.CreatedAt,
	}
}

type Subscription struct {
	ID                 uuid.UUID  `json:"id"`
	CustomerID         uuid.UUID  `json:"customer_id"`
	Plan               *Plan      `json:"plan,omitempty"`
	Status             string     `json:"status"`
	CurrentPeriodStart *time.Time `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`
	CanceledAt         *time.Time `json:"canceled_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

func FromSubscription(s db.Subscription, plan *db.Plan) Subscription {
	out := Subscription{
		ID: s.ID, CustomerID: s.CustomerID, Status: s.Status,
		CurrentPeriodStart: tsPtr(s.CurrentPeriodStart),
		CurrentPeriodEnd:   tsPtr(s.CurrentPeriodEnd),
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		CanceledAt:         tsPtr(s.CanceledAt),
		CreatedAt:          s.CreatedAt,
	}
	if plan != nil {
		p := FromPlan(*plan)
		out.Plan = &p
	}
	return out
}

type Invoice struct {
	ID             uuid.UUID  `json:"id"`
	SubscriptionID *string    `json:"subscription_id"`
	AmountMinor    int64      `json:"amount_minor"`
	Currency       string     `json:"currency"`
	Status         string     `json:"status"`
	PeriodStart    *time.Time `json:"period_start"`
	PeriodEnd      *time.Time `json:"period_end"`
	IssuedAt       time.Time  `json:"issued_at"`
}

func FromInvoice(i db.Invoice) Invoice {
	return Invoice{
		ID: i.ID, SubscriptionID: uuidPtr(i.SubscriptionID), AmountMinor: i.Amount,
		Currency: i.Currency, Status: i.Status, PeriodStart: tsPtr(i.PeriodStart),
		PeriodEnd: tsPtr(i.PeriodEnd), IssuedAt: i.IssuedAt,
	}
}

type InvoiceItem struct {
	ID          uuid.UUID  `json:"id"`
	Description string     `json:"description"`
	AmountMinor int64      `json:"amount_minor"`
	Quantity    int32      `json:"quantity"`
	PeriodStart *time.Time `json:"period_start"`
	PeriodEnd   *time.Time `json:"period_end"`
}

func FromInvoiceItem(it db.InvoiceItem) InvoiceItem {
	return InvoiceItem{
		ID: it.ID, Description: it.Description, AmountMinor: it.Amount,
		Quantity: it.Quantity, PeriodStart: tsPtr(it.PeriodStart), PeriodEnd: tsPtr(it.PeriodEnd),
	}
}

type InvoiceDetail struct {
	Invoice
	Items []InvoiceItem `json:"items"`
}

type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	Name       *string    `json:"name"`
	Masked     string     `json:"masked"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

func FromAPIKey(k db.ApiKey) APIKey {
	return APIKey{ID: k.ID, Name: k.Name, Masked: k.KeyPrefix + "…", CreatedAt: k.CreatedAt, LastUsedAt: tsPtr(k.LastUsedAt)}
}

type Money struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}

type Settings struct {
	BusinessName         string  `json:"business_name"`
	SupportEmail         *string `json:"support_email"`
	NombaAccountID       *string `json:"nomba_account_id"`
	NombaSubaccountID    *string `json:"nomba_subaccount_id"`
	NombaClientID        *string `json:"nomba_client_id"`
	NombaClientSecretSet bool    `json:"nomba_client_secret_set"`
	WebhookURL           *string `json:"webhook_url"`
	WebhookSecretSet     bool    `json:"webhook_secret_set"`
}

func FromSettings(t db.Tenant) Settings {
	return Settings{
		BusinessName:         t.Name,
		SupportEmail:         t.SupportEmail,
		NombaAccountID:       t.NombaAccountID,
		NombaSubaccountID:    t.NombaSubaccountID,
		NombaClientID:        t.NombaClientID,
		NombaClientSecretSet: t.NombaClientSecret != nil && *t.NombaClientSecret != "",
		WebhookURL:           t.WebhookUrl,
		WebhookSecretSet:     t.WebhookSecret != nil && *t.WebhookSecret != "",
	}
}

type PortalContext struct {
	Customer       Customer `json:"customer"`
	TenantBranding struct {
		TenantName string  `json:"tenant_name"`
		LogoURL    *string `json:"logo_url"`
	} `json:"tenant_branding"`
	Token string `json:"token,omitempty"`
}
