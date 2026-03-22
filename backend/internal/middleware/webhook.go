package middleware

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
)

// webhookBodyKey stores the raw request body for later use by handlers.
type webhookBodyKey struct{}

// VerifyRazorpayWebhook is middleware that verifies the HMAC-SHA256 signature
// on incoming Razorpay webhook requests. Invalid signatures are rejected with
// 400 Bad Request. The raw request body is stored in context for downstream
// handlers to use.
func VerifyRazorpayWebhook(verifyFn func(payload []byte, signature string) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			signature := r.Header.Get("X-Razorpay-Signature")
			if signature == "" {
				slog.Warn("webhook missing signature header",
					"remote_addr", r.RemoteAddr,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":{"code":"INVALID_SIGNATURE","message":"missing X-Razorpay-Signature header"}}`))
				return
			}

			// Read body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("webhook: failed to read body", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			// Verify HMAC-SHA256 signature
			if !verifyFn(body, signature) {
				slog.Warn("webhook signature verification failed",
					"remote_addr", r.RemoteAddr,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":{"code":"INVALID_SIGNATURE","message":"webhook signature verification failed"}}`))
				return
			}

			// Store body in context for handler use, restore body for downstream
			ctx := context.WithValue(r.Context(), webhookBodyKey{}, body)
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WebhookBodyFromContext retrieves the raw webhook body from context.
func WebhookBodyFromContext(ctx context.Context) []byte {
	if body, ok := ctx.Value(webhookBodyKey{}).([]byte); ok {
		return body
	}
	return nil
}
