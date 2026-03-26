package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// listColumns are the columns returned in list queries (summary — no detail fields).
const listColumns = `id, slug, name, logo_url, description, size, headquarters,
	tech_stack, domains, hiring_status, compensation_tier,
	has_rsu, has_rsu_refresher, office_modes, updated_at`

// detailColumns are the columns returned in detail queries (full company profile).
const detailColumns = `id, slug, name, logo_url, description, size, headquarters,
	founded_year, careers_page_url, glassdoor_url, ambitionbox_url, linkedin_url,
	tech_stack, domains, hiring_status, interview_patterns, compensation_tier,
	has_rsu, has_rsu_refresher, office_modes, compensation_bands,
	last_verified_at, created_at, updated_at`

// CompanyRepo implements domain.CompanyRepository using pgx.
type CompanyRepo struct {
	pool *pgxpool.Pool
}

// NewCompanyRepo creates a new CompanyRepo.
func NewCompanyRepo(pool *pgxpool.Pool) *CompanyRepo {
	return &CompanyRepo{pool: pool}
}

// List returns companies matching the filter, ordered by the specified sort.
func (r *CompanyRepo) List(ctx context.Context, filter domain.CompanyFilter) ([]domain.Company, string, error) {
	q := getDBTX(ctx, r.pool)

	limit := normaliseLimit(filter.Limit)
	sort, order := normaliseSort(filter.Sort, filter.Order)

	// Build WHERE clauses
	args := []any{}
	clauses := []string{}
	argIdx := 1

	argIdx = appendFilterClauses(&clauses, &args, argIdx, filter)

	// Cursor pagination
	if filter.Cursor != "" {
		cursorClause, cursorArgs, _, err := decodeCursor(filter.Cursor, sort, order, argIdx)
		if err == nil {
			clauses = append(clauses, cursorClause)
			args = append(args, cursorArgs...)
		}
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	orderClause := buildOrderClause(sort, order)

	query := fmt.Sprintf(`SELECT %s FROM companies %s %s LIMIT %d`,
		listColumns, where, orderClause, limit+1)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, "", domain.InternalError(err)
	}
	defer rows.Close()

	companies, err := scanCompanyListRows(rows)
	if err != nil {
		return nil, "", err
	}

	// Determine next cursor
	var nextCursor string
	if len(companies) > limit {
		companies = companies[:limit]
		last := companies[len(companies)-1]
		nextCursor = encodeCursor(last, sort)
	}

	return companies, nextCursor, nil
}

// Search returns companies matching a search query plus filters.
// It uses prefix-enabled full-text search (word:*) so partial words like
// "goog" match "Google". For very short queries (single token < 3 chars)
// it falls back to ILIKE on the name column.
func (r *CompanyRepo) Search(ctx context.Context, query string, filter domain.CompanyFilter) ([]domain.Company, string, error) {
	q := getDBTX(ctx, r.pool)

	limit := normaliseLimit(filter.Limit)
	sort, order := normaliseSort(filter.Sort, filter.Order)

	args := []any{}
	clauses := []string{}
	argIdx := 1

	// Determine search strategy: prefix tsquery or ILIKE fallback.
	ftsArgIdx := 0 // 0 = not using FTS ranking
	prefixTsquery := buildPrefixTsquery(query)

	if prefixTsquery != "" {
		// Prefix full-text search:  "goog" → to_tsquery('english','goog:*')
		args = append(args, prefixTsquery)
		clauses = append(clauses, fmt.Sprintf("search_vector @@ to_tsquery('english', $%d)", argIdx))
		ftsArgIdx = argIdx
		argIdx++
	} else {
		// Very short / unparseable query → ILIKE fallback on name
		args = append(args, "%"+query+"%")
		clauses = append(clauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		argIdx++
	}

	argIdx = appendFilterClauses(&clauses, &args, argIdx, filter)

	// Cursor pagination
	if filter.Cursor != "" {
		cursorClause, cursorArgs, _, err := decodeCursor(filter.Cursor, sort, order, argIdx)
		if err == nil {
			clauses = append(clauses, cursorClause)
			args = append(args, cursorArgs...)
		}
	}

	where := "WHERE " + strings.Join(clauses, " AND ")

	// Default sort for search is relevance (only available with FTS)
	var orderClause string
	if ftsArgIdx > 0 && (sort == "" || sort == "relevance") {
		orderClause = fmt.Sprintf("ORDER BY ts_rank(search_vector, to_tsquery('english', $%d)) DESC, id ASC", ftsArgIdx)
	} else {
		orderClause = buildOrderClause(sort, order)
	}

	sql := fmt.Sprintf(`SELECT %s FROM companies %s %s LIMIT %d`,
		listColumns, where, orderClause, limit+1)

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, "", domain.InternalError(err)
	}
	defer rows.Close()

	companies, err := scanCompanyListRows(rows)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(companies) > limit {
		companies = companies[:limit]
		last := companies[len(companies)-1]
		nextCursor = encodeCursor(last, sort)
	}

	return companies, nextCursor, nil
}

// buildPrefixTsquery converts a user search string into a prefix tsquery.
// "goog"        → "goog:*"
// "amazon web"  → "amazon:* & web:*"
// Returns "" if no valid tokens can be built (caller should use ILIKE).
func buildPrefixTsquery(query string) string {
	words := strings.Fields(query)
	if len(words) == 0 {
		return ""
	}

	tokens := make([]string, 0, len(words))
	for _, w := range words {
		// Strip characters that are special in tsquery syntax
		cleaned := sanitiseTsqueryToken(w)
		if cleaned == "" {
			continue
		}
		tokens = append(tokens, cleaned+":*")
	}

	if len(tokens) == 0 {
		return ""
	}

	return strings.Join(tokens, " & ")
}

// sanitiseTsqueryToken removes characters that are special in PostgreSQL
// tsquery syntax so the token can be safely interpolated.
func sanitiseTsqueryToken(s string) string {
	var b strings.Builder
	for _, r := range s {
		// Allow only alphanumeric, underscore, hyphen
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// GetByID retrieves a company by UUID with full detail.
func (r *CompanyRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	q := getDBTX(ctx, r.pool)

	sql := fmt.Sprintf("SELECT %s FROM companies WHERE id = $1", detailColumns)
	return scanCompanyDetail(q.QueryRow(ctx, sql, id))
}

// GetBySlug retrieves a company by URL slug with full detail.
func (r *CompanyRepo) GetBySlug(ctx context.Context, slug string) (*domain.Company, error) {
	q := getDBTX(ctx, r.pool)

	sql := fmt.Sprintf("SELECT %s FROM companies WHERE slug = $1", detailColumns)
	return scanCompanyDetail(q.QueryRow(ctx, sql, slug))
}

// Create inserts a new company. The company.ID must be set by the caller.
func (r *CompanyRepo) Create(ctx context.Context, company *domain.Company) error {
	q := getDBTX(ctx, r.pool)

	now := time.Now().UTC()
	if company.CreatedAt.IsZero() {
		company.CreatedAt = now
	}
	company.UpdatedAt = now

	err := q.QueryRow(ctx, `
		INSERT INTO companies (
			id, slug, name, logo_url, description, size, headquarters,
			founded_year, careers_page_url, glassdoor_url, ambitionbox_url, linkedin_url,
			tech_stack, domains, hiring_status, interview_patterns, compensation_tier,
			has_rsu, has_rsu_refresher, office_modes, compensation_bands,
			last_verified_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18, $19, $20, $21,
			$22, $23, $24
		) RETURNING id`,
		company.ID, company.Slug, company.Name, company.LogoURL, company.Description,
		companySizeToString(company.Size), company.Headquarters,
		company.FoundedYear, company.CareersPageURL, company.GlassdoorURL,
		company.AmbitionboxURL, company.LinkedinURL,
		company.TechStack, company.Domains, string(company.HiringStatus),
		company.InterviewPatterns, company.CompensationTier,
		company.HasRSU, company.HasRSURefresher, company.OfficeModes, company.CompensationBands,
		company.LastVerifiedAt, company.CreatedAt, company.UpdatedAt,
	).Scan(&company.ID)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.Conflict("company", "slug already exists")
		}
		return domain.InternalError(err)
	}
	return nil
}

// Update persists changes to an existing company.
func (r *CompanyRepo) Update(ctx context.Context, company *domain.Company) error {
	q := getDBTX(ctx, r.pool)
	company.UpdatedAt = time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE companies SET
			slug = $2, name = $3, logo_url = $4, description = $5,
			size = $6, headquarters = $7, founded_year = $8,
			careers_page_url = $9, glassdoor_url = $10,
			ambitionbox_url = $11, linkedin_url = $12,
			tech_stack = $13, domains = $14, hiring_status = $15,
			interview_patterns = $16, compensation_tier = $17,
			has_rsu = $18, has_rsu_refresher = $19, office_modes = $20,
			compensation_bands = $21, last_verified_at = $22, updated_at = $23
		WHERE id = $1
		RETURNING id`,
		company.ID, company.Slug, company.Name, company.LogoURL, company.Description,
		companySizeToString(company.Size), company.Headquarters, company.FoundedYear,
		company.CareersPageURL, company.GlassdoorURL,
		company.AmbitionboxURL, company.LinkedinURL,
		company.TechStack, company.Domains, string(company.HiringStatus),
		company.InterviewPatterns, company.CompensationTier,
		company.HasRSU, company.HasRSURefresher, company.OfficeModes,
		company.CompensationBands, company.LastVerifiedAt, company.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("company", company.ID)
		}
		if isUniqueViolation(err) {
			return domain.Conflict("company", "slug already exists")
		}
		return domain.InternalError(err)
	}
	return nil
}

// Upsert inserts or updates a company by slug.
// Used by the seed runner to idempotently load company data.
func (r *CompanyRepo) Upsert(ctx context.Context, company *domain.Company) error {
	q := getDBTX(ctx, r.pool)

	now := time.Now().UTC()
	if company.CreatedAt.IsZero() {
		company.CreatedAt = now
	}
	company.UpdatedAt = now

	err := q.QueryRow(ctx, `
		INSERT INTO companies (
			id, slug, name, logo_url, description, size, headquarters,
			founded_year, careers_page_url, glassdoor_url, ambitionbox_url, linkedin_url,
			tech_stack, domains, hiring_status, interview_patterns, compensation_tier,
			has_rsu, has_rsu_refresher, office_modes, compensation_bands,
			last_verified_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18, $19, $20, $21,
			$22, $23, $24
		)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			logo_url = EXCLUDED.logo_url,
			description = EXCLUDED.description,
			size = EXCLUDED.size,
			headquarters = EXCLUDED.headquarters,
			founded_year = EXCLUDED.founded_year,
			careers_page_url = EXCLUDED.careers_page_url,
			glassdoor_url = EXCLUDED.glassdoor_url,
			ambitionbox_url = EXCLUDED.ambitionbox_url,
			linkedin_url = EXCLUDED.linkedin_url,
			tech_stack = EXCLUDED.tech_stack,
			domains = EXCLUDED.domains,
			hiring_status = EXCLUDED.hiring_status,
			interview_patterns = EXCLUDED.interview_patterns,
			compensation_tier = EXCLUDED.compensation_tier,
			has_rsu = EXCLUDED.has_rsu,
			has_rsu_refresher = EXCLUDED.has_rsu_refresher,
			office_modes = EXCLUDED.office_modes,
			compensation_bands = EXCLUDED.compensation_bands,
			last_verified_at = EXCLUDED.last_verified_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id`,
		company.ID, company.Slug, company.Name, company.LogoURL, company.Description,
		companySizeToString(company.Size), company.Headquarters,
		company.FoundedYear, company.CareersPageURL, company.GlassdoorURL,
		company.AmbitionboxURL, company.LinkedinURL,
		company.TechStack, company.Domains, string(company.HiringStatus),
		company.InterviewPatterns, company.CompensationTier,
		company.HasRSU, company.HasRSURefresher, company.OfficeModes, company.CompensationBands,
		company.LastVerifiedAt, company.CreatedAt, company.UpdatedAt,
	).Scan(&company.ID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// --- Filter helpers ---

// appendFilterClauses builds WHERE clauses from a CompanyFilter.
func appendFilterClauses(clauses *[]string, args *[]any, argIdx int, filter domain.CompanyFilter) int {
	if len(filter.Sizes) > 0 {
		strs := make([]string, len(filter.Sizes))
		for i, s := range filter.Sizes {
			strs[i] = string(s)
		}
		*args = append(*args, strs)
		*clauses = append(*clauses, fmt.Sprintf("size = ANY($%d)", argIdx))
		argIdx++
	}

	if filter.HiringStatus != nil {
		*args = append(*args, string(*filter.HiringStatus))
		*clauses = append(*clauses, fmt.Sprintf("hiring_status = $%d", argIdx))
		argIdx++
	}

	if len(filter.TechStack) > 0 {
		*args = append(*args, filter.TechStack)
		*clauses = append(*clauses, fmt.Sprintf("tech_stack @> $%d", argIdx))
		argIdx++
	}

	if len(filter.Domains) > 0 {
		*args = append(*args, filter.Domains)
		*clauses = append(*clauses, fmt.Sprintf("domains && $%d", argIdx))
		argIdx++
	}

	if len(filter.CompensationTiers) > 0 {
		*args = append(*args, filter.CompensationTiers)
		*clauses = append(*clauses, fmt.Sprintf("compensation_tier = ANY($%d)", argIdx))
		argIdx++
	}

	if filter.HasRSU != nil {
		*args = append(*args, *filter.HasRSU)
		*clauses = append(*clauses, fmt.Sprintf("has_rsu = $%d", argIdx))
		argIdx++
	}

	if filter.Headquarters != "" {
		*args = append(*args, "%"+filter.Headquarters+"%")
		*clauses = append(*clauses, fmt.Sprintf("headquarters ILIKE $%d", argIdx))
		argIdx++
	}

	return argIdx
}

// --- Sort / Order helpers ---

// allowedSorts is the set of valid sort fields for company listing.
var allowedSorts = map[string]string{
	"name":              "name",
	"size":              "size",
	"compensation_tier": "compensation_tier",
	"updated_at":        "updated_at",
}

func normaliseSort(sort, order string) (string, string) {
	if _, ok := allowedSorts[sort]; !ok {
		sort = "name"
	}
	if order != "desc" {
		order = "asc"
	}
	return sort, order
}

func normaliseLimit(limit int) int {
	if limit < 1 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func buildOrderClause(sort, order string) string {
	col := allowedSorts[sort]
	if col == "" {
		col = "name"
	}
	dir := "ASC"
	if order == "desc" {
		dir = "DESC"
	}

	// Secondary sort by id for deterministic pagination
	idDir := "ASC"
	if dir == "DESC" {
		idDir = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s, id %s", col, dir, idDir)
}

// --- Cursor encoding/decoding ---

type cursorData struct {
	ID        string `json:"id"`
	SortValue string `json:"sv,omitempty"`
}

func encodeCursor(c domain.Company, sort string) string {
	cd := cursorData{ID: c.ID.String()}

	switch sort {
	case "name":
		cd.SortValue = c.Name
	case "updated_at":
		cd.SortValue = c.UpdatedAt.Format(time.RFC3339Nano)
	case "size":
		if c.Size != nil {
			cd.SortValue = string(*c.Size)
		}
	case "compensation_tier":
		if c.CompensationTier != nil {
			cd.SortValue = *c.CompensationTier
		}
	default:
		cd.SortValue = c.Name
	}

	data, _ := json.Marshal(cd)
	return base64.URLEncoding.EncodeToString(data)
}

func decodeCursor(cursor, sort, order string, argIdx int) (string, []any, int, error) {
	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", nil, argIdx, err
	}

	var cd cursorData
	if err := json.Unmarshal(data, &cd); err != nil {
		return "", nil, argIdx, err
	}

	cursorID, err := uuid.Parse(cd.ID)
	if err != nil {
		return "", nil, argIdx, err
	}

	col := allowedSorts[sort]
	if col == "" {
		col = "name"
	}

	// Keyset pagination: (sort_col, id) > (cursor_sort_val, cursor_id)
	op := ">"
	if order == "desc" {
		op = "<"
	}

	// For string/text columns use direct comparison
	clause := fmt.Sprintf("(%s, id) %s ($%d, $%d)", col, op, argIdx, argIdx+1)

	var sortArg any
	switch sort {
	case "updated_at":
		t, parseErr := time.Parse(time.RFC3339Nano, cd.SortValue)
		if parseErr != nil {
			return "", nil, argIdx, parseErr
		}
		sortArg = t
	default:
		sortArg = cd.SortValue
	}

	args := []any{sortArg, cursorID}
	return clause, args, argIdx + 2, nil
}

// --- Row scanners ---

// scanCompanyListRows scans multiple rows into list-summary Company structs.
func scanCompanyListRows(rows pgx.Rows) ([]domain.Company, error) {
	var companies []domain.Company
	for rows.Next() {
		c, err := scanCompanyListRow(rows)
		if err != nil {
			return nil, err
		}
		companies = append(companies, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return companies, nil
}

// scanCompanyListRow scans a single row of list columns.
func scanCompanyListRow(row pgx.Row) (*domain.Company, error) {
	c := &domain.Company{}
	var size *string
	var compTier *string

	err := row.Scan(
		&c.ID, &c.Slug, &c.Name, &c.LogoURL, &c.Description,
		&size, &c.Headquarters,
		&c.TechStack, &c.Domains, &c.HiringStatus, &compTier,
		&c.HasRSU, &c.HasRSURefresher, &c.OfficeModes, &c.UpdatedAt,
	)
	if err != nil {
		return nil, domain.InternalError(err)
	}

	if size != nil {
		s := domain.CompanySize(*size)
		c.Size = &s
	}
	c.CompensationTier = compTier

	if c.TechStack == nil {
		c.TechStack = []string{}
	}
	if c.Domains == nil {
		c.Domains = []string{}
	}
	if c.OfficeModes == nil {
		c.OfficeModes = []string{}
	}

	return c, nil
}

// scanCompanyDetail scans a single row of detail columns into a full Company.
func scanCompanyDetail(row pgx.Row) (*domain.Company, error) {
	c := &domain.Company{}
	var size *string
	var compTier *string

	err := row.Scan(
		&c.ID, &c.Slug, &c.Name, &c.LogoURL, &c.Description,
		&size, &c.Headquarters,
		&c.FoundedYear, &c.CareersPageURL, &c.GlassdoorURL,
		&c.AmbitionboxURL, &c.LinkedinURL,
		&c.TechStack, &c.Domains, &c.HiringStatus,
		&c.InterviewPatterns, &compTier,
		&c.HasRSU, &c.HasRSURefresher, &c.OfficeModes, &c.CompensationBands,
		&c.LastVerifiedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("company", nil)
		}
		return nil, domain.InternalError(err)
	}

	if size != nil {
		s := domain.CompanySize(*size)
		c.Size = &s
	}
	c.CompensationTier = compTier

	if c.TechStack == nil {
		c.TechStack = []string{}
	}
	if c.Domains == nil {
		c.Domains = []string{}
	}
	if c.OfficeModes == nil {
		c.OfficeModes = []string{}
	}

	return c, nil
}

// Delete hard-deletes a company by ID.
func (r *CompanyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)
	tag, err := q.Exec(ctx, `DELETE FROM companies WHERE id = $1`, id)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("company", id)
	}
	return nil
}

// ListAll returns all companies as compact list-column summaries, ordered by name.
// Results are capped at 500 rows — sufficient for the AI curation use-case.
func (r *CompanyRepo) ListAll(ctx context.Context) ([]domain.Company, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, fmt.Sprintf(
		`SELECT %s FROM companies ORDER BY name ASC LIMIT 500`, listColumns,
	))
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	return scanCompanyListRows(rows)
}

// GetNamesByIDs returns a map of company ID → name for the given IDs.
// This is a lightweight batch lookup used to enrich list entries with company names.
func (r *CompanyRepo) GetNamesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]string{}, nil
	}

	q := getDBTX(ctx, r.pool)
	rows, err := q.Query(ctx, `SELECT id, name FROM companies WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]string, len(ids))
	for rows.Next() {
		var id uuid.UUID
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, domain.InternalError(err)
		}
		result[id] = name
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return result, nil
}

// GetNameAndSlugsByIDs returns a map of company ID → {name, slug} for the given IDs.
func (r *CompanyRepo) GetNameAndSlugsByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.CompanyNameSlug, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]domain.CompanyNameSlug{}, nil
	}

	q := getDBTX(ctx, r.pool)
	rows, err := q.Query(ctx, `SELECT id, name, slug FROM companies WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]domain.CompanyNameSlug, len(ids))
	for rows.Next() {
		var id uuid.UUID
		var ns domain.CompanyNameSlug
		if err := rows.Scan(&id, &ns.Name, &ns.Slug); err != nil {
			return nil, domain.InternalError(err)
		}
		result[id] = ns
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return result, nil
}

// companySizeToString converts *CompanySize to *string for DB.
func companySizeToString(s *domain.CompanySize) *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}
