package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// Product catalog — prices in paise, credit allocations per product type.
var productCatalog = map[domain.ProductType]productConfig{
	domain.ProductStarterPack: {
		AmountPaise: 39900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 9,
			domain.CreditATSCheck:     20,
			domain.CreditCuratedList:  3,
		},
		SetsPremium: true,
	},
	domain.ProductResumeUpload: {
		AmountPaise: 4900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 1,
		},
	},
	domain.ProductATSBundle: {
		AmountPaise: 9900,
		Credits: map[domain.CreditType]int{
			domain.CreditATSCheck: 10,
		},
	},
	domain.ProductRebuyPack: {
		AmountPaise: 39900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 9,
			domain.CreditATSCheck:     20,
			domain.CreditCuratedList:  3,
		},
	},
}

type productConfig struct {
	AmountPaise int
	Credits     map[domain.CreditType]int
	SetsPremium bool
}

// PaymentService handles payment orchestration, Razorpay integration,
// webhook processing, and credit allocation.
type PaymentService struct {
	payments domain.PaymentRepository
	credits  domain.CreditRepository
	users    domain.UserRepository
	gateway  domain.PaymentGateway
	tx       domain.Transactor
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(
	payments domain.PaymentRepository,
	credits domain.CreditRepository,
	users domain.UserRepository,
	gateway domain.PaymentGateway,
	tx domain.Transactor,
) *PaymentService {
	return &PaymentService{
		payments: payments,
		credits:  credits,
		users:    users,
		gateway:  gateway,
		tx:       tx,
	}
}

// CreateOrderInput holds input for creating a payment order.
type CreateOrderInput struct {
	UserID      uuid.UUID
	ProductType domain.ProductType
}

// CreateOrderResult holds the data needed by the frontend to open Razorpay checkout.
type CreateOrderResult struct {
	PaymentID       uuid.UUID          `json:"payment_id"`
	RazorpayOrderID string             `json:"razorpay_order_id"`
	AmountPaise     int                `json:"amount_paise"`
	Currency        string             `json:"currency"`
	ProductType     domain.ProductType `json:"product_type"`
	RazorpayKeyID   string             `json:"razorpay_key_id"`
}

// CreateOrder validates the product, creates a Razorpay order, and records
// the payment in the database.
func (s *PaymentService) CreateOrder(ctx context.Context, input CreateOrderInput, razorpayKeyID string) (*CreateOrderResult, error) {
	// Validate product type
	product, ok := productCatalog[input.ProductType]
	if !ok {
		return nil, domain.ValidationError("invalid product_type", map[string]any{
			"field":  "product_type",
			"reason": "unknown product type",
		})
	}

	// Business rules
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	if input.ProductType == domain.ProductStarterPack && user.IsPremium() {
		return nil, domain.ValidationError("starter_pack is only available for non-premium users", map[string]any{
			"field":  "product_type",
			"reason": "already_premium",
		})
	}
	if input.ProductType == domain.ProductRebuyPack && !user.IsPremium() {
		return nil, domain.ValidationError("rebuy_pack is only available for premium users", map[string]any{
			"field":  "product_type",
			"reason": "not_premium",
		})
	}

	// Create Razorpay order
	order, err := s.gateway.CreateOrder(ctx, &domain.CreateOrderRequest{
		UserID:      input.UserID,
		AmountPaise: product.AmountPaise,
		ProductType: input.ProductType,
	})
	if err != nil {
		return nil, err
	}

	// Record payment in DB
	now := time.Now().UTC()
	payment := &domain.Payment{
		ID:              uuid.Must(uuid.NewV7()),
		UserID:          input.UserID,
		RazorpayOrderID: order.RazorpayOrderID,
		AmountPaise:     product.AmountPaise,
		Currency:        order.Currency,
		ProductType:     input.ProductType,
		Status:          domain.PaymentStatusCreated,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, err
	}

	return &CreateOrderResult{
		PaymentID:       payment.ID,
		RazorpayOrderID: order.RazorpayOrderID,
		AmountPaise:     product.AmountPaise,
		Currency:        order.Currency,
		ProductType:     input.ProductType,
		RazorpayKeyID:   razorpayKeyID,
	}, nil
}

// razorpayWebhookPayload represents the relevant parts of a Razorpay webhook event.
type razorpayWebhookPayload struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID      string `json:"id"`
				OrderID string `json:"order_id"`
				Status  string `json:"status"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

// HandleWebhook processes a Razorpay webhook event. The signature has already
// been verified by the middleware. This method is idempotent — duplicate
// webhooks for the same order are safely ignored.
func (s *PaymentService) HandleWebhook(ctx context.Context, body []byte) error {
	var payload razorpayWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return domain.ValidationError("invalid webhook payload", nil)
	}

	// Only process payment.captured events
	if payload.Event != "payment.captured" {
		slog.Info("ignoring webhook event", "event", payload.Event)
		return nil
	}

	orderID := payload.Payload.Payment.Entity.OrderID
	razorpayPaymentID := payload.Payload.Payment.Entity.ID

	if orderID == "" || razorpayPaymentID == "" {
		return domain.ValidationError("missing order_id or payment_id in webhook", nil)
	}

	// Look up payment record
	payment, err := s.payments.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	// Idempotency: already captured → return success
	if payment.Status == domain.PaymentStatusCaptured {
		slog.Info("duplicate webhook ignored",
			"razorpay_order_id", orderID,
			"payment_id", payment.ID,
		)
		return nil
	}

	// Atomic: capture payment + allocate credits
	product, ok := productCatalog[payment.ProductType]
	if !ok {
		return domain.InternalError(fmt.Errorf("unknown product type in payment: %s", payment.ProductType))
	}

	err = s.tx.WithTx(ctx, func(txCtx context.Context) error {
		// 1. Mark payment as captured
		if err := s.payments.UpdateWebhookCapture(txCtx, payment.ID, razorpayPaymentID, time.Now().UTC()); err != nil {
			return err
		}

		// 2. Allocate credits per product
		for creditType, amount := range product.Credits {
			if err := s.credits.Allocate(txCtx, payment.UserID, creditType, amount); err != nil {
				return err
			}

			// Get new balance for audit trail
			newBalance, err := s.credits.GetBalance(txCtx, payment.UserID, creditType)
			if err != nil {
				return err
			}

			txn := &domain.CreditTransaction{
				ID:           uuid.Must(uuid.NewV7()),
				UserID:       payment.UserID,
				CreditType:   creditType,
				Amount:       amount,
				BalanceAfter: newBalance,
				Reason:       fmt.Sprintf("purchase_%s", payment.ProductType),
				ReferenceID:  &payment.ID,
				CreatedAt:    time.Now().UTC(),
			}
			if err := s.credits.LogTransaction(txCtx, txn); err != nil {
				return err
			}
		}

		// 3. Set premium_since for starter_pack (if not already premium)
		if product.SetsPremium {
			user, err := s.users.GetByID(txCtx, payment.UserID)
			if err != nil {
				return err
			}
			if !user.IsPremium() {
				now := time.Now().UTC()
				user.PremiumSince = &now
				if err := s.users.Update(txCtx, user); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	slog.Info("payment_webhook_processed",
		"user_id", payment.UserID,
		"payment_id", payment.ID,
		"razorpay_order_id", orderID,
		"product_type", payment.ProductType,
		"amount_paise", payment.AmountPaise,
	)

	return nil
}

// ConfirmPaymentInput holds client-side payment confirmation data from Razorpay checkout.
type ConfirmPaymentInput struct {
	RazorpayOrderID   string
	RazorpayPaymentID string
	RazorpaySignature string
}

// ConfirmPayment verifies a client-side Razorpay payment signature and processes
// the payment if valid. It is idempotent — if the webhook already captured the
// payment, it returns success immediately. This is the primary path for granting
// credits; the webhook is a fallback for cases where the client cannot reach us.
func (s *PaymentService) ConfirmPayment(ctx context.Context, input ConfirmPaymentInput) error {
	// Look up payment record
	payment, err := s.payments.GetByOrderID(ctx, input.RazorpayOrderID)
	if err != nil {
		return err
	}

	// Idempotency: webhook already captured this payment
	if payment.Status == domain.PaymentStatusCaptured {
		slog.Info("confirm: payment already captured",
			"razorpay_order_id", input.RazorpayOrderID,
			"payment_id", payment.ID,
		)
		return nil
	}

	// Verify Razorpay client-side signature
	verification, err := s.gateway.VerifyPayment(ctx, &domain.VerifyPaymentRequest{
		OrderID:   input.RazorpayOrderID,
		PaymentID: input.RazorpayPaymentID,
		Signature: input.RazorpaySignature,
	})
	if err != nil {
		return err
	}
	if !verification.Verified {
		return &domain.AppError{
			Code:    domain.ErrCodeForbidden,
			Message: "payment signature verification failed",
		}
	}

	// Atomic: capture payment + allocate credits (same as webhook path)
	product, ok := productCatalog[payment.ProductType]
	if !ok {
		return domain.InternalError(fmt.Errorf("unknown product type in payment: %s", payment.ProductType))
	}

	err = s.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.payments.UpdateWebhookCapture(txCtx, payment.ID, input.RazorpayPaymentID, time.Now().UTC()); err != nil {
			return err
		}

		for creditType, amount := range product.Credits {
			if err := s.credits.Allocate(txCtx, payment.UserID, creditType, amount); err != nil {
				return err
			}

			newBalance, err := s.credits.GetBalance(txCtx, payment.UserID, creditType)
			if err != nil {
				return err
			}

			txn := &domain.CreditTransaction{
				ID:           uuid.Must(uuid.NewV7()),
				UserID:       payment.UserID,
				CreditType:   creditType,
				Amount:       amount,
				BalanceAfter: newBalance,
				Reason:       fmt.Sprintf("purchase_%s", payment.ProductType),
				ReferenceID:  &payment.ID,
				CreatedAt:    time.Now().UTC(),
			}
			if err := s.credits.LogTransaction(txCtx, txn); err != nil {
				return err
			}
		}

		if product.SetsPremium {
			user, err := s.users.GetByID(txCtx, payment.UserID)
			if err != nil {
				return err
			}
			if !user.IsPremium() {
				now := time.Now().UTC()
				user.PremiumSince = &now
				if err := s.users.Update(txCtx, user); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	slog.Info("payment_confirm_processed",
		"user_id", payment.UserID,
		"payment_id", payment.ID,
		"razorpay_order_id", input.RazorpayOrderID,
		"product_type", payment.ProductType,
		"amount_paise", payment.AmountPaise,
	)

	return nil
}

// ListPayments returns all payments for a user.
func (s *PaymentService) ListPayments(ctx context.Context, userID uuid.UUID) ([]domain.Payment, error) {
	return s.payments.ListByUser(ctx, userID)
}
