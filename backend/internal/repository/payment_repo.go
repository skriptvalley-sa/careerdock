package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// PaymentRepo implements domain.PaymentRepository using pgx.
type PaymentRepo struct {
	pool *pgxpool.Pool
}

// NewPaymentRepo creates a new PaymentRepo.
func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool}
}

// Create inserts a new payment record.
func (r *PaymentRepo) Create(ctx context.Context, payment *domain.Payment) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO payments (
			id, user_id, razorpay_order_id, razorpay_payment_id,
			amount_paise, currency, product_type, status,
			receipt_number, webhook_received_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10,
			$11, $12
		) RETURNING id`,
		payment.ID, payment.UserID, payment.RazorpayOrderID, payment.RazorpayPaymentID,
		payment.AmountPaise, payment.Currency, string(payment.ProductType), string(payment.Status),
		payment.ReceiptNumber, payment.WebhookReceivedAt,
		payment.CreatedAt, payment.UpdatedAt,
	).Scan(&payment.ID)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.Conflict("payment", "order already exists")
		}
		return domain.InternalError(err)
	}
	return nil
}

// GetByOrderID retrieves a payment by its Razorpay order ID.
func (r *PaymentRepo) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	q := getDBTX(ctx, r.pool)

	p := &domain.Payment{}
	err := q.QueryRow(ctx, `
		SELECT id, user_id, razorpay_order_id, razorpay_payment_id,
		       amount_paise, currency, product_type, status,
		       receipt_number, refund_reason, refunded_at, refunded_by,
		       webhook_received_at, created_at, updated_at
		FROM payments
		WHERE razorpay_order_id = $1`, orderID,
	).Scan(
		&p.ID, &p.UserID, &p.RazorpayOrderID, &p.RazorpayPaymentID,
		&p.AmountPaise, &p.Currency, &p.ProductType, &p.Status,
		&p.ReceiptNumber, &p.RefundReason, &p.RefundedAt, &p.RefundedBy,
		&p.WebhookReceivedAt, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("payment", orderID)
		}
		return nil, domain.InternalError(err)
	}
	return p, nil
}

// UpdateStatus updates a payment's status and sets updated_at.
func (r *PaymentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	q := getDBTX(ctx, r.pool)

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE payments SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id`, id, string(status),
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("payment", id)
		}
		return domain.InternalError(err)
	}
	return nil
}

// UpdateWebhookCapture atomically captures a payment: sets status, razorpay_payment_id, and webhook timestamp.
func (r *PaymentRepo) UpdateWebhookCapture(ctx context.Context, id uuid.UUID, razorpayPaymentID string, webhookReceivedAt time.Time) error {
	q := getDBTX(ctx, r.pool)

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE payments SET
			status = $2,
			razorpay_payment_id = $3,
			webhook_received_at = $4,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id`,
		id, string(domain.PaymentStatusCaptured), razorpayPaymentID, webhookReceivedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("payment", id)
		}
		return domain.InternalError(err)
	}
	return nil
}

// ListByUser retrieves all payments for a user, newest first.
func (r *PaymentRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Payment, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT id, user_id, razorpay_order_id, razorpay_payment_id,
		       amount_paise, currency, product_type, status,
		       receipt_number, refund_reason, refunded_at, refunded_by,
		       webhook_received_at, created_at, updated_at
		FROM payments
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.RazorpayOrderID, &p.RazorpayPaymentID,
			&p.AmountPaise, &p.Currency, &p.ProductType, &p.Status,
			&p.ReceiptNumber, &p.RefundReason, &p.RefundedAt, &p.RefundedBy,
			&p.WebhookReceivedAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, domain.InternalError(err)
		}
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return payments, nil
}
