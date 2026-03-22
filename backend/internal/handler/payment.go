package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// PaymentHandler handles payment and credit HTTP endpoints.
type PaymentHandler struct {
	payment *service.PaymentService
	credit  *service.CreditService

	razorpayKeyID string
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(payment *service.PaymentService, credit *service.CreditService, razorpayKeyID string) *PaymentHandler {
	return &PaymentHandler{
		payment:       payment,
		credit:        credit,
		razorpayKeyID: razorpayKeyID,
	}
}

// --- Request DTOs ---

type createOrderRequest struct {
	ProductType string `json:"product_type"`
}

// --- Response DTOs ---

type paymentResponse struct {
	ID                string  `json:"id"`
	RazorpayOrderID   string  `json:"razorpay_order_id"`
	RazorpayPaymentID *string `json:"razorpay_payment_id,omitempty"`
	AmountPaise       int     `json:"amount_paise"`
	Currency          string  `json:"currency"`
	ProductType       string  `json:"product_type"`
	Status            string  `json:"status"`
	CreatedAt         string  `json:"created_at"`
}

type creditBalancesResponse struct {
	ResumeUpload int `json:"resume_upload"`
	ATSCheck     int `json:"ats_check"`
	CuratedList  int `json:"curated_list"`
	CVGeneration int `json:"cv_generation"`
}

type creditTransactionResponse struct {
	ID           string  `json:"id"`
	CreditType   string  `json:"credit_type"`
	Amount       int     `json:"amount"`
	BalanceAfter int     `json:"balance_after"`
	Reason       string  `json:"reason"`
	ReferenceID  *string `json:"reference_id,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// --- Handlers ---

// CreateOrder handles POST /api/payments/orders.
func (h *PaymentHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	if req.ProductType == "" {
		respondError(w, r, domain.ValidationError("product_type is required", map[string]any{
			"field": "product_type",
		}))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())

	result, err := h.payment.CreateOrder(r.Context(), service.CreateOrderInput{
		UserID:      userID,
		ProductType: domain.ProductType(req.ProductType),
	}, h.razorpayKeyID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, DataResponse{Data: result})
}

// HandleWebhook handles POST /api/webhooks/razorpay.
// Signature verification is done by the middleware — this handler receives
// a verified payload.
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body := middleware.WebhookBodyFromContext(r.Context())
	if body == nil {
		respondError(w, r, domain.ValidationError("missing webhook body", nil))
		return
	}

	if err := h.payment.HandleWebhook(r.Context(), body); err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{
		Data: map[string]string{"status": "ok"},
	})
}

// ListPayments handles GET /api/payments.
func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	payments, err := h.payment.ListPayments(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	resp := make([]paymentResponse, len(payments))
	for i, p := range payments {
		resp[i] = toPaymentResponse(&p)
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: resp})
}

// GetCredits handles GET /api/credits.
func (h *PaymentHandler) GetCredits(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	balances, err := h.credit.GetAllBalances(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	resp := creditBalancesResponse{
		ResumeUpload: balances[domain.CreditResumeUpload],
		ATSCheck:     balances[domain.CreditATSCheck],
		CuratedList:  balances[domain.CreditCuratedList],
		CVGeneration: balances[domain.CreditCVGeneration],
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: resp})
}

// GetCreditTransactions handles GET /api/credits/transactions.
func (h *PaymentHandler) GetCreditTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	// Parse optional credit_type filter
	var creditType *domain.CreditType
	if ct := r.URL.Query().Get("credit_type"); ct != "" {
		parsed, err := service.ValidateCreditType(ct)
		if err != nil {
			respondError(w, r, err)
			return
		}
		creditType = &parsed
	}

	// Parse limit
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	txns, err := h.credit.ListTransactions(r.Context(), userID, creditType, limit)
	if err != nil {
		respondError(w, r, err)
		return
	}

	resp := make([]creditTransactionResponse, len(txns))
	for i, t := range txns {
		resp[i] = toCreditTransactionResponse(&t)
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: resp})
}

// --- Converters ---

func toPaymentResponse(p *domain.Payment) paymentResponse {
	resp := paymentResponse{
		ID:              p.ID.String(),
		RazorpayOrderID: p.RazorpayOrderID,
		AmountPaise:     p.AmountPaise,
		Currency:        p.Currency,
		ProductType:     string(p.ProductType),
		Status:          string(p.Status),
		CreatedAt:       p.CreatedAt.Format(time.RFC3339),
	}
	if p.RazorpayPaymentID != nil {
		resp.RazorpayPaymentID = p.RazorpayPaymentID
	}
	return resp
}

func toCreditTransactionResponse(t *domain.CreditTransaction) creditTransactionResponse {
	resp := creditTransactionResponse{
		ID:           t.ID.String(),
		CreditType:   string(t.CreditType),
		Amount:       t.Amount,
		BalanceAfter: t.BalanceAfter,
		Reason:       t.Reason,
		CreatedAt:    t.CreatedAt.Format(time.RFC3339),
	}
	if t.ReferenceID != nil {
		s := t.ReferenceID.String()
		resp.ReferenceID = &s
	}
	return resp
}
