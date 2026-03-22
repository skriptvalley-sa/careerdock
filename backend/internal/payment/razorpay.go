// Package payment implements the domain.PaymentGateway interface for Razorpay.
package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/skriptvalley/careerdock/internal/domain"
)

const razorpayBaseURL = "https://api.razorpay.com/v1"

// RazorpayGateway implements domain.PaymentGateway using the Razorpay API.
type RazorpayGateway struct {
	keyID         string
	keySecret     string
	webhookSecret string
	client        *http.Client
}

// NewRazorpayGateway creates a new RazorpayGateway.
func NewRazorpayGateway(keyID, keySecret, webhookSecret string) *RazorpayGateway {
	return &RazorpayGateway{
		keyID:         keyID,
		keySecret:     keySecret,
		webhookSecret: webhookSecret,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// KeyID returns the Razorpay key ID (needed by frontend for checkout).
func (g *RazorpayGateway) KeyID() string {
	return g.keyID
}

// razorpayOrderRequest is the Razorpay Create Order API request body.
type razorpayOrderRequest struct {
	Amount   int               `json:"amount"`
	Currency string            `json:"currency"`
	Receipt  string            `json:"receipt,omitempty"`
	Notes    map[string]string `json:"notes,omitempty"`
}

// razorpayOrderResponse is the Razorpay Create Order API response body.
type razorpayOrderResponse struct {
	ID       string `json:"id"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

// CreateOrder calls the Razorpay Create Order API.
func (g *RazorpayGateway) CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	body := razorpayOrderRequest{
		Amount:   req.AmountPaise,
		Currency: "INR",
		Notes: map[string]string{
			"user_id":      req.UserID.String(),
			"product_type": string(req.ProductType),
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("marshal order request: %w", err))
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, razorpayBaseURL+"/orders", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("create HTTP request: %w", err))
	}

	httpReq.SetBasicAuth(g.keyID, g.keySecret)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, &domain.AppError{
			Code:    domain.ErrCodePaymentFailed,
			Message: "Failed to connect to payment gateway",
			Err:     err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("read razorpay response: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &domain.AppError{
			Code:    domain.ErrCodePaymentFailed,
			Message: "Payment gateway returned an error",
			Err:     fmt.Errorf("razorpay status %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	var orderResp razorpayOrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, domain.InternalError(fmt.Errorf("unmarshal razorpay response: %w", err))
	}

	return &domain.Order{
		RazorpayOrderID: orderResp.ID,
		AmountPaise:     orderResp.Amount,
		Currency:        orderResp.Currency,
	}, nil
}

// VerifyPayment verifies a Razorpay payment signature.
// The signature is HMAC-SHA256(order_id|payment_id, key_secret).
func (g *RazorpayGateway) VerifyPayment(_ context.Context, req *domain.VerifyPaymentRequest) (*domain.PaymentVerification, error) {
	message := req.OrderID + "|" + req.PaymentID
	mac := hmac.New(sha256.New, []byte(g.keySecret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	verified := hmac.Equal([]byte(expected), []byte(req.Signature))
	return &domain.PaymentVerification{Verified: verified}, nil
}

// VerifyWebhookSignature verifies a Razorpay webhook signature.
// The signature is HMAC-SHA256(request_body, webhook_secret).
func (g *RazorpayGateway) VerifyWebhookSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(g.webhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
