package service

import (
	"encoding/json"
	"testing"

	"github.com/skriptvalley/careerdock/internal/domain"
)

func TestLookupProductConfigForPayment_LegacyStarterPack(t *testing.T) {
	t.Parallel()

	payment := &domain.Payment{
		ProductType: domain.ProductStarterPack,
		AmountPaise: 39900,
	}

	product, ok := lookupProductConfigForPayment(payment)
	if !ok {
		t.Fatalf("expected legacy starter pack config to resolve")
	}

	if product.AmountPaise != 39900 {
		t.Fatalf("expected amount 39900, got %d", product.AmountPaise)
	}
	if !product.SetsPremium {
		t.Fatalf("expected legacy starter pack to set premium")
	}
	if got := product.Credits[domain.CreditResumeUpload]; got != 9 {
		t.Fatalf("expected 9 resume credits, got %d", got)
	}
	if got := product.Credits[domain.CreditATSCheck]; got != 20 {
		t.Fatalf("expected 20 ATS credits, got %d", got)
	}
	if got := product.Credits[domain.CreditCuratedList]; got != 3 {
		t.Fatalf("expected 3 curated list credits, got %d", got)
	}
}

func TestLookupProductConfigForPayment_LegacyPreviousPriceResolves(t *testing.T) {
	t.Parallel()

	payment := &domain.Payment{
		ProductType: domain.ProductATSBundle,
		AmountPaise: 24900,
	}

	product, ok := lookupProductConfigForPayment(payment)
	if !ok {
		t.Fatalf("expected legacy ATS bundle config to resolve")
	}

	if got := product.Credits[domain.CreditATSCheck]; got != 50 {
		t.Fatalf("expected 50 ATS credits, got %d", got)
	}
}

func TestLookupProductConfigForPayment_ActiveProductRequiresMatchingAmount(t *testing.T) {
	t.Parallel()

	payment := &domain.Payment{
		ProductType: domain.ProductATSBundle,
		AmountPaise: 22900,
	}

	product, ok := lookupProductConfigForPayment(payment)
	if !ok {
		t.Fatalf("expected active ATS bundle config to resolve")
	}

	if got := product.Credits[domain.CreditATSCheck]; got != 50 {
		t.Fatalf("expected 50 ATS credits, got %d", got)
	}

	mismatched := &domain.Payment{
		ProductType: domain.ProductStarterRefill,
		AmountPaise: 40123,
	}

	if _, ok := lookupProductConfigForPayment(mismatched); ok {
		t.Fatalf("expected mismatched active amount to be rejected")
	}
}

func TestAllocationForPayment_CartSnapshotAggregatesCredits(t *testing.T) {
	t.Parallel()

	cartSnapshot, err := json.Marshal([]cartSnapshotItem{
		{
			ProductType: domain.ProductResumeBundle,
			Quantity:    2,
			AmountPaise: 8900,
			Credits: map[domain.CreditType]int{
				domain.CreditResumeUpload: 10,
			},
		},
		{
			ProductType: domain.ProductCuratedListBundle,
			Quantity:    3,
			AmountPaise: 5900,
			Credits: map[domain.CreditType]int{
				domain.CreditCuratedList: 5,
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal cart snapshot: %v", err)
	}

	payment := &domain.Payment{
		ProductType:  domain.ProductCartBundle,
		AmountPaise:  35500,
		CartSnapshot: cartSnapshot,
	}

	product, ok := allocationForPayment(payment)
	if !ok {
		t.Fatalf("expected cart bundle config to resolve")
	}

	if got := product.Credits[domain.CreditResumeUpload]; got != 20 {
		t.Fatalf("expected 20 resume credits, got %d", got)
	}
	if got := product.Credits[domain.CreditCuratedList]; got != 15 {
		t.Fatalf("expected 15 curated list credits, got %d", got)
	}
}
