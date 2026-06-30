// Package notify sends transactional email via Resend (magic links, and later
// receipts and dunning, shared with the engine workers).
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

const resendEndpoint = "https://api.resend.com/emails"

type Mailer struct {
	apiKey    string
	fromName  string
	fromEmail string
	http      *http.Client
	log       zerolog.Logger
}

func NewMailer(apiKey, fromName, fromEmail string, log zerolog.Logger) *Mailer {
	return &Mailer{
		apiKey:    apiKey,
		fromName:  fromName,
		fromEmail: fromEmail,
		http:      &http.Client{Timeout: 10 * time.Second},
		log:       log,
	}
}

// SendMagicLink emails a portal access link. With no API key configured (local dev),
// it logs the link instead of sending, so the flow is testable without Resend.
func (m *Mailer) SendMagicLink(ctx context.Context, toEmail, tenantName, link string) error {
	subject := fmt.Sprintf("Your secure link to manage your %s subscription", tenantName)
	html := fmt.Sprintf(
		`<p>Tap the button below to securely access your subscription. This link expires in 15 minutes and can be used once.</p>`+
			`<p><a href="%s" style="display:inline-block;padding:12px 20px;background:#5B47E0;color:#fff;border-radius:8px;text-decoration:none">Manage subscription</a></p>`+
			`<p>If you didn't request this, you can ignore this email.</p>`, link)
	return m.send(ctx, toEmail, tenantName, subject, html, link)
}

func (m *Mailer) send(ctx context.Context, to, tenantName, subject, html, link string) error {
	if m.apiKey == "" {
		// Dev fallback: no provider configured, surface the link in logs.
		m.log.Warn().Str("to", to).Str("link", link).Msg("RESEND_API_KEY unset; logging magic link instead of emailing")
		return nil
	}
	from := fmt.Sprintf("%s via %s <%s>", tenantName, m.fromName, m.fromEmail)
	body, _ := json.Marshal(map[string]any{
		"from": from, "to": []string{to}, "subject": subject, "html": html,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("resend returned %d", resp.StatusCode)
	}
	return nil
}
