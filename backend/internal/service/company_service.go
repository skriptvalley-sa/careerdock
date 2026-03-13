package service

import (
	"context"
	"strings"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// CompanyService handles company directory business logic.
type CompanyService struct {
	companies domain.CompanyRepository
}

// NewCompanyService creates a new CompanyService.
func NewCompanyService(companies domain.CompanyRepository) *CompanyService {
	return &CompanyService{companies: companies}
}

// List returns a filtered, sorted, paginated list of companies.
// If query is non-empty, full-text search is used.
func (s *CompanyService) List(ctx context.Context, query string, filter domain.CompanyFilter) ([]domain.Company, string, error) {
	query = strings.TrimSpace(query)
	if query != "" {
		return s.companies.Search(ctx, query, filter)
	}
	return s.companies.List(ctx, filter)
}

// GetBySlug returns a single company by its URL slug.
func (s *CompanyService) GetBySlug(ctx context.Context, slug string) (*domain.Company, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return nil, domain.ValidationError("slug is required", nil)
	}
	return s.companies.GetBySlug(ctx, slug)
}
