package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// EmailSendHandler sends transactional emails via the configured email sender.
type EmailSendHandler struct {
	sender domain.EmailSender
}

// NewEmailSendHandler creates a handler for email:send tasks.
func NewEmailSendHandler(sender domain.EmailSender) *EmailSendHandler {
	return &EmailSendHandler{sender: sender}
}

// EmailPayload is the JSON payload for the email:send task.
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// Handle processes an email:send task.
func (h *EmailSendHandler) Handle(ctx context.Context, t *asynq.Task) error {
	var payload EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal email payload: %w", err)
	}

	slog.Info("sending email", "to", payload.To, "subject", payload.Subject)

	msg := &domain.EmailMessage{
		To:      payload.To,
		Subject: payload.Subject,
		HTML:    payload.HTML,
	}

	if err := h.sender.Send(ctx, msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	slog.Info("email sent successfully", "to", payload.To)
	return nil
}

// EnqueueEmail creates an Asynq task to send an email.
func EnqueueEmail(client *asynq.Client, to, subject, html string) error {
	payload, err := json.Marshal(EmailPayload{
		To:      to,
		Subject: subject,
		HTML:    html,
	})
	if err != nil {
		return fmt.Errorf("marshal email payload: %w", err)
	}

	task := asynq.NewTask("email:send", payload)
	_, err = client.Enqueue(task,
		asynq.Queue("critical"),
		asynq.MaxRetry(5),
	)
	return err
}
