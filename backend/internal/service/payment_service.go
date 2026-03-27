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

const maxCartQuantityPerProduct = 5

// activeProductCatalog defines the currently purchasable products.
var activeProductCatalog = map[domain.ProductType]productConfig{
	domain.ProductStarterPack: {
		AmountPaise: 44900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 10,
			domain.CreditATSCheck:     50,
			domain.CreditCuratedList:  10,
			domain.CreditCVGeneration: 50,
		},
		SetsPremium: true,
	},
	domain.ProductStarterRefill: {
		AmountPaise: 39900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 10,
			domain.CreditATSCheck:     50,
			domain.CreditCuratedList:  10,
			domain.CreditCVGeneration: 50,
		},
	},
	domain.ProductResumeBundle: {
		AmountPaise: 8900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 10,
		},
	},
	domain.ProductATSBundle: {
		AmountPaise: 22900,
		Credits: map[domain.CreditType]int{
			domain.CreditATSCheck: 50,
		},
	},
	domain.ProductCuratedListBundle: {
		AmountPaise: 5900,
		Credits: map[domain.CreditType]int{
			domain.CreditCuratedList: 5,
		},
	},
}

type legacyProductKey struct {
	ProductType domain.ProductType
	AmountPaise int
}

// legacyProductCatalog remains available so previously-created orders can still
// be confirmed after the active catalogue changes.
var legacyProductCatalog = map[legacyProductKey]productConfig{
	{
		ProductType: domain.ProductStarterPack,
		AmountPaise: 79900,
	}: {
		AmountPaise: 79900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 10,
			domain.CreditATSCheck:     50,
			domain.CreditCuratedList:  10,
			domain.CreditCVGeneration: 50,
		},
		SetsPremium: true,
	},
	{
		ProductType: domain.ProductStarterRefill,
		AmountPaise: 79900,
	}: {
		AmountPaise: 79900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 10,
			domain.CreditATSCheck:     50,
			domain.CreditCuratedList:  10,
			domain.CreditCVGeneration: 50,
		},
	},
	{
		ProductType: domain.ProductResumeBundle,
		AmountPaise: 19900,
	}: {
		AmountPaise: 19900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 10,
		},
	},
	{
		ProductType: domain.ProductATSBundle,
		AmountPaise: 24900,
	}: {
		AmountPaise: 24900,
		Credits: map[domain.CreditType]int{
			domain.CreditATSCheck: 50,
		},
	},
	{
		ProductType: domain.ProductCuratedListBundle,
		AmountPaise: 14900,
	}: {
		AmountPaise: 14900,
		Credits: map[domain.CreditType]int{
			domain.CreditCuratedList: 5,
		},
	},
	{
		ProductType: domain.ProductCVBundle,
		AmountPaise: 24900,
	}: {
		AmountPaise: 24900,
		Credits: map[domain.CreditType]int{
			domain.CreditCVGeneration: 50,
		},
	},
	{
		ProductType: domain.ProductStarterPack,
		AmountPaise: 39900,
	}: {
		AmountPaise: 39900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 9,
			domain.CreditATSCheck:     20,
			domain.CreditCuratedList:  3,
		},
		SetsPremium: true,
	},
	{
		ProductType: domain.ProductATSBundle,
		AmountPaise: 9900,
	}: {
		AmountPaise: 9900,
		Credits: map[domain.CreditType]int{
			domain.CreditATSCheck: 10,
		},
	},
	{
		ProductType: domain.ProductType("resume_upload"),
		AmountPaise: 4900,
	}: {
		AmountPaise: 4900,
		Credits: map[domain.CreditType]int{
			domain.CreditResumeUpload: 1,
		},
	},
	{
		ProductType: domain.ProductType("rebuy_pack"),
		AmountPaise: 39900,
	}: {
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

type cartSnapshotItem struct {
	ProductType domain.ProductType        `json:"product_type"`
	Quantity    int                       `json:"quantity"`
	AmountPaise int                       `json:"amount_paise"`
	Credits     map[domain.CreditType]int `json:"credits"`
	SetsPremium bool                      `json:"sets_premium,omitempty"`
}

func lookupProductConfigForPayment(payment *domain.Payment) (productConfig, bool) {
	if legacy, ok := legacyProductCatalog[legacyProductKey{
		ProductType: payment.ProductType,
		AmountPaise: payment.AmountPaise,
	}]; ok {
		return legacy, true
	}

	product, ok := activeProductCatalog[payment.ProductType]
	if !ok {
		return productConfig{}, false
	}

	if product.AmountPaise != payment.AmountPaise {
		return productConfig{}, false
	}

	return product, true
}

func lookupCartConfigForPayment(payment *domain.Payment) (productConfig, bool) {
	if payment.ProductType != domain.ProductCartBundle || len(payment.CartSnapshot) == 0 {
		return productConfig{}, false
	}

	var items []cartSnapshotItem
	if err := json.Unmarshal(payment.CartSnapshot, &items); err != nil || len(items) == 0 {
		return productConfig{}, false
	}

	product := productConfig{
		AmountPaise: payment.AmountPaise,
		Credits:     make(map[domain.CreditType]int),
	}
	totalAmount := 0

	for _, item := range items {
		if item.Quantity < 1 || item.AmountPaise <= 0 {
			return productConfig{}, false
		}

		totalAmount += item.AmountPaise * item.Quantity
		if item.SetsPremium {
			product.SetsPremium = true
		}

		for creditType, amount := range item.Credits {
			product.Credits[creditType] += amount * item.Quantity
		}
	}

	if totalAmount != payment.AmountPaise {
		return productConfig{}, false
	}

	return product, true
}

func allocationForPayment(payment *domain.Payment) (productConfig, bool) {
	if product, ok := lookupCartConfigForPayment(payment); ok {
		return product, true
	}
	return lookupProductConfigForPayment(payment)
}

func cloneCredits(credits map[domain.CreditType]int) map[domain.CreditType]int {
	cloned := make(map[domain.CreditType]int, len(credits))
	for creditType, amount := range credits {
		cloned[creditType] = amount
	}
	return cloned
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

// CreateOrderItemInput represents one product line in a checkout request.
type CreateOrderItemInput struct {
	ProductType domain.ProductType
	Quantity    int
}

// CreateOrderInput holds input for creating a payment order.
type CreateOrderInput struct {
	UserID      uuid.UUID
	ProductType domain.ProductType
	CartItems   []CreateOrderItemInput
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

func validateProductForPurchase(productType domain.ProductType) (productConfig, error) {
	if productType == domain.ProductCVBundle {
		return productConfig{}, domain.ValidationError("cover letter bundle is coming soon", map[string]any{
			"field":  "product_type",
			"reason": "coming_soon",
		})
	}

	product, ok := activeProductCatalog[productType]
	if !ok {
		return productConfig{}, domain.ValidationError("invalid product_type", map[string]any{
			"field":  "product_type",
			"reason": "unknown_product_type",
		})
	}

	return product, nil
}

func validateUserCanPurchaseProduct(user *domain.User, productType domain.ProductType) error {
	if productType == domain.ProductStarterPack && user.IsPremium() {
		return domain.ValidationError("starter_pack is only available for non-premium users", map[string]any{
			"field":  "product_type",
			"reason": "already_premium",
		})
	}

	if productType != domain.ProductStarterPack && !user.IsPremium() {
		return domain.ValidationError(fmt.Sprintf("%s is only available for premium users", productType), map[string]any{
			"field":  "product_type",
			"reason": "not_premium",
		})
	}

	return nil
}

// CreateOrder validates the product, creates a Razorpay order, and records
// the payment in the database.
func (s *PaymentService) CreateOrder(ctx context.Context, input CreateOrderInput, razorpayKeyID string) (*CreateOrderResult, error) {
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	orderProductType := input.ProductType
	orderAmountPaise := 0
	var cartSnapshot json.RawMessage

	if len(input.CartItems) > 0 {
		if !user.IsPremium() {
			return nil, domain.ValidationError("cart checkout is only available for premium users", map[string]any{
				"field":  "items",
				"reason": "not_premium",
			})
		}

		items := make([]cartSnapshotItem, 0, len(input.CartItems))
		for _, item := range input.CartItems {
			if item.ProductType == domain.ProductStarterPack {
				return nil, domain.ValidationError("starter_pack cannot be purchased through the credit shop cart", map[string]any{
					"field":  "items",
					"reason": "starter_pack_not_allowed",
				})
			}
			if item.Quantity < 1 || item.Quantity > maxCartQuantityPerProduct {
				return nil, domain.ValidationError("invalid cart quantity", map[string]any{
					"field":    "items.quantity",
					"min":      1,
					"max":      maxCartQuantityPerProduct,
					"quantity": item.Quantity,
				})
			}

			product, err := validateProductForPurchase(item.ProductType)
			if err != nil {
				return nil, err
			}

			orderAmountPaise += product.AmountPaise * item.Quantity
			items = append(items, cartSnapshotItem{
				ProductType: item.ProductType,
				Quantity:    item.Quantity,
				AmountPaise: product.AmountPaise,
				Credits:     cloneCredits(product.Credits),
				SetsPremium: product.SetsPremium,
			})
		}

		if len(items) == 0 {
			return nil, domain.ValidationError("items are required", map[string]any{
				"field": "items",
			})
		}

		cartSnapshot, err = json.Marshal(items)
		if err != nil {
			return nil, domain.InternalError(fmt.Errorf("marshal cart snapshot: %w", err))
		}
		orderProductType = domain.ProductCartBundle
	} else {
		if err := validateUserCanPurchaseProduct(user, input.ProductType); err != nil {
			return nil, err
		}

		product, err := validateProductForPurchase(input.ProductType)
		if err != nil {
			return nil, err
		}
		orderAmountPaise = product.AmountPaise
	}

	// Create Razorpay order
	order, err := s.gateway.CreateOrder(ctx, &domain.CreateOrderRequest{
		UserID:      input.UserID,
		AmountPaise: orderAmountPaise,
		ProductType: orderProductType,
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
		AmountPaise:     orderAmountPaise,
		Currency:        order.Currency,
		ProductType:     orderProductType,
		CartSnapshot:    cartSnapshot,
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
		AmountPaise:     orderAmountPaise,
		Currency:        order.Currency,
		ProductType:     orderProductType,
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
	product, ok := allocationForPayment(payment)
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
	product, ok := allocationForPayment(payment)
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
