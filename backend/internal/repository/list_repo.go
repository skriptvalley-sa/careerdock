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

// ListRepo implements domain.ListRepository using pgx.
type ListRepo struct {
	pool *pgxpool.Pool
}

// NewListRepo creates a new ListRepo.
func NewListRepo(pool *pgxpool.Pool) *ListRepo {
	return &ListRepo{pool: pool}
}

// --- List CRUD ---

// CreateList inserts a new user list.
func (r *ListRepo) CreateList(ctx context.Context, list *domain.UserList) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO user_lists (id, user_id, name, description, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		list.ID, list.UserID, list.Name, list.Description,
		list.Position, list.CreatedAt, list.UpdatedAt,
	).Scan(&list.ID)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.Conflict("list", "a list with that name already exists")
		}
		return domain.InternalError(err)
	}
	return nil
}

// GetListByID retrieves a list by ID.
func (r *ListRepo) GetListByID(ctx context.Context, id uuid.UUID) (*domain.UserList, error) {
	q := getDBTX(ctx, r.pool)

	list := &domain.UserList{}
	err := q.QueryRow(ctx, `
		SELECT id, user_id, name, description, position, created_at, updated_at
		FROM user_lists
		WHERE id = $1`, id,
	).Scan(&list.ID, &list.UserID, &list.Name, &list.Description,
		&list.Position, &list.CreatedAt, &list.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("list", id)
		}
		return nil, domain.InternalError(err)
	}
	return list, nil
}

// ListByUser returns all lists for a user, ordered by position.
func (r *ListRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserList, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT l.id, l.user_id, l.name, l.description, l.position, l.created_at, l.updated_at,
		       (SELECT COUNT(*) FROM list_entries WHERE list_id = l.id) AS entry_count
		FROM user_lists l
		WHERE l.user_id = $1
		ORDER BY l.position ASC`, userID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var lists []domain.UserList
	for rows.Next() {
		var list domain.UserList
		var entryCount int
		if err := rows.Scan(&list.ID, &list.UserID, &list.Name, &list.Description,
			&list.Position, &list.CreatedAt, &list.UpdatedAt, &entryCount); err != nil {
			return nil, domain.InternalError(err)
		}
		lists = append(lists, list)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}

	return lists, nil
}

// UpdateList updates an existing list's name, description, or position.
func (r *ListRepo) UpdateList(ctx context.Context, list *domain.UserList) error {
	q := getDBTX(ctx, r.pool)
	list.UpdatedAt = time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE user_lists SET name = $2, description = $3, position = $4, updated_at = $5
		WHERE id = $1
		RETURNING id`,
		list.ID, list.Name, list.Description, list.Position, list.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("list", list.ID)
		}
		if isUniqueViolation(err) {
			return domain.Conflict("list", "a list with that name already exists")
		}
		return domain.InternalError(err)
	}
	return nil
}

// DeleteList removes a list. Entries cascade via DB foreign key.
func (r *ListRepo) DeleteList(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `DELETE FROM user_lists WHERE id = $1 RETURNING id`, id).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("list", id)
		}
		return domain.InternalError(err)
	}
	return nil
}

// CountByUser returns the number of lists a user has.
func (r *ListRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	q := getDBTX(ctx, r.pool)

	var count int
	err := q.QueryRow(ctx, `SELECT COUNT(*) FROM user_lists WHERE user_id = $1`, userID).Scan(&count)
	if err != nil {
		return 0, domain.InternalError(err)
	}
	return count, nil
}

// --- Entry CRUD ---

// CreateEntry inserts a new list entry (company membership in a list).
func (r *ListRepo) CreateEntry(ctx context.Context, entry *domain.ListEntry) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO list_entries (id, list_id, company_id, company_status, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		entry.ID, entry.ListID, entry.CompanyID, string(entry.CompanyStatus),
		entry.Position, entry.CreatedAt, entry.UpdatedAt,
	).Scan(&entry.ID)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.Conflict("list entry", "this company is already in the list")
		}
		return domain.InternalError(err)
	}
	return nil
}

// GetEntryByID retrieves a list entry by ID.
func (r *ListRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.ListEntry, error) {
	q := getDBTX(ctx, r.pool)

	entry := &domain.ListEntry{}
	err := q.QueryRow(ctx, `
		SELECT id, list_id, company_id, company_status, position, created_at, updated_at
		FROM list_entries
		WHERE id = $1`, id,
	).Scan(&entry.ID, &entry.ListID, &entry.CompanyID, &entry.CompanyStatus,
		&entry.Position, &entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("list entry", id)
		}
		return nil, domain.InternalError(err)
	}
	return entry, nil
}

// ListEntries returns all entries for a list, ordered by position.
func (r *ListRepo) ListEntries(ctx context.Context, listID uuid.UUID) ([]domain.ListEntry, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT id, list_id, company_id, company_status, position, created_at, updated_at
		FROM list_entries
		WHERE list_id = $1
		ORDER BY position ASC`, listID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var entries []domain.ListEntry
	for rows.Next() {
		var e domain.ListEntry
		if err := rows.Scan(&e.ID, &e.ListID, &e.CompanyID, &e.CompanyStatus,
			&e.Position, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, domain.InternalError(err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}

	return entries, nil
}

// ListEntriesByCompanyID returns all entries for a specific company across all
// of a user's lists. Useful for showing which lists contain a given company.
func (r *ListRepo) ListEntriesByCompanyID(ctx context.Context, userID, companyID uuid.UUID) ([]domain.ListEntryWithList, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT le.id, le.list_id, le.company_id, le.company_status,
		       le.position, le.created_at, le.updated_at,
		       ul.name AS list_name
		FROM list_entries le
		JOIN user_lists ul ON ul.id = le.list_id
		WHERE ul.user_id = $1 AND le.company_id = $2
		ORDER BY le.created_at DESC`, userID, companyID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var entries []domain.ListEntryWithList
	for rows.Next() {
		var e domain.ListEntryWithList
		if err := rows.Scan(&e.ID, &e.ListID, &e.CompanyID, &e.CompanyStatus,
			&e.Position, &e.CreatedAt, &e.UpdatedAt, &e.ListName); err != nil {
			return nil, domain.InternalError(err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	if entries == nil {
		entries = []domain.ListEntryWithList{}
	}
	return entries, nil
}

// ListsWithCompanyFlag returns all lists for a user, each annotated with
// whether it contains the given company. Used for the quick-add-to-list modal.
func (r *ListRepo) ListsWithCompanyFlag(ctx context.Context, userID, companyID uuid.UUID) ([]domain.ListCompanyFlag, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT ul.id, ul.name,
		       EXISTS(SELECT 1 FROM list_entries le WHERE le.list_id = ul.id AND le.company_id = $2) AS contains_company
		FROM user_lists ul
		WHERE ul.user_id = $1
		ORDER BY ul.position ASC`, userID, companyID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var flags []domain.ListCompanyFlag
	for rows.Next() {
		var f domain.ListCompanyFlag
		if err := rows.Scan(&f.ListID, &f.Name, &f.ContainsCompany); err != nil {
			return nil, domain.InternalError(err)
		}
		flags = append(flags, f)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	if flags == nil {
		flags = []domain.ListCompanyFlag{}
	}
	return flags, nil
}

// CompanyListCounts returns a map of company_id → number of lists containing it
// for all companies across the user's lists. Used for the company card list indicator.
func (r *ListRepo) CompanyListCounts(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT le.company_id, COUNT(DISTINCT le.list_id)::int
		FROM list_entries le
		JOIN user_lists ul ON ul.id = le.list_id
		WHERE ul.user_id = $1
		GROUP BY le.company_id`, userID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	counts := make(map[uuid.UUID]int)
	for rows.Next() {
		var companyID uuid.UUID
		var count int
		if err := rows.Scan(&companyID, &count); err != nil {
			return nil, domain.InternalError(err)
		}
		counts[companyID] = count
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return counts, nil
}

// DeleteEntryByCompany removes a list entry by list ID and company ID.
// Used by the quick-add-to-list modal's remove action.
func (r *ListRepo) DeleteEntryByCompany(ctx context.Context, listID, companyID uuid.UUID) error {
	q := getDBTX(ctx, r.pool)

	var returnedID uuid.UUID
	err := q.QueryRow(ctx,
		`DELETE FROM list_entries WHERE list_id = $1 AND company_id = $2 RETURNING id`,
		listID, companyID,
	).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("list entry", listID)
		}
		return domain.InternalError(err)
	}
	return nil
}

// UpdateEntry updates an existing list entry (company_status and position).
func (r *ListRepo) UpdateEntry(ctx context.Context, entry *domain.ListEntry) error {
	q := getDBTX(ctx, r.pool)
	entry.UpdatedAt = time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE list_entries SET
			company_status = $2, position = $3, updated_at = $4
		WHERE id = $1
		RETURNING id`,
		entry.ID, string(entry.CompanyStatus), entry.Position, entry.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("list entry", entry.ID)
		}
		return domain.InternalError(err)
	}
	return nil
}

// DeleteEntry removes a list entry. Status history and interview rounds cascade.
func (r *ListRepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `DELETE FROM list_entries WHERE id = $1 RETURNING id`, id).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("list entry", id)
		}
		return domain.InternalError(err)
	}
	return nil
}
