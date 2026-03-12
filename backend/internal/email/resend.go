// Package email provides transactional email delivery via the Resend API.
// In development, emails are logged rather than sent (use Mailhog for
// actual SMTP testing via Docker Compose).
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/skriptvalley/careerdock/internal/domain"
)

const resendAPIURL = "https://api.resend.com/emails"

// ResendSender sends transactional emails via the Resend API.
// Implements domain.EmailSender.
type ResendSender struct {
	apiKey string
	from   string
	client *http.Client
}

// NewResendSender creates a new ResendSender.
// If apiKey is empty, sends are logged but not delivered (dev mode).
func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send delivers an email via the Resend API.
func (s *ResendSender) Send(ctx context.Context, msg *domain.EmailMessage) error {
	if s.apiKey == "" {
		slog.Info("email send (dev mode — not delivered)",
			"to", msg.To,
			"subject", msg.Subject,
		)
		return nil
	}

	payload := map[string]string{
		"from":    s.from,
		"to":      msg.To,
		"subject": msg.Subject,
		"html":    msg.HTML,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create email request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send email request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	slog.Info("email sent successfully",
		"to", msg.To,
		"subject", msg.Subject,
	)

	return nil
}
