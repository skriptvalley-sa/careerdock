package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// List limits per tier.
const (
	FreeListLimit    = 3
	PremiumListLimit = 5
)

// ListService handles business logic for user lists and entries.
type ListService struct {
	lists domain.ListRepository
	users domain.UserRepository
	tx    domain.Transactor
}

// NewListService creates a new ListService.
func NewListService(
	lists domain.ListRepository,
	users domain.UserRepository,
	tx domain.Transactor,
) *ListService {
	return &ListService{lists: lists, users: users, tx: tx}
}

// --- List operations ---

// CreateListInput holds input for creating a list.
type CreateListInput struct {
	UserID      uuid.UUID
	Name        string
	Description string
}

// CreateList creates a new user list, enforcing the tier limit.
func (s *ListService) CreateList(ctx context.Context, input CreateListInput) (*domain.UserList, error) {
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	limit := FreeListLimit
	if user.IsPremium() {
		limit = PremiumListLimit
	}

	count, err := s.lists.CountByUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	if count >= limit {
		return nil, domain.ValidationError("list limit reached", map[string]any{
			"max":     limit,
			"current": count,
		})
	}

	now := time.Now().UTC()
	var desc *string
	if input.Description != "" {
		desc = &input.Description
	}

	list := &domain.UserList{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      input.UserID,
		Name:        input.Name,
		Description: desc,
		Position:    count,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.lists.CreateList(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

// ListsByUser returns all lists for a user.
func (s *ListService) ListsByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserList, error) {
	return s.lists.ListByUser(ctx, userID)
}

// GetList returns a list by ID, verifying ownership.
func (s *ListService) GetList(ctx context.Context, listID, userID uuid.UUID) (*domain.UserList, error) {
	list, err := s.lists.GetListByID(ctx, listID)
	if err != nil {
		return nil, err
	}
	if list.UserID != userID {
		return nil, domain.Forbidden("you do not own this list")
	}
	return list, nil
}

// UpdateListInput holds input for updating a list.
type UpdateListInput struct {
	ListID      uuid.UUID
	UserID      uuid.UUID
	Name        *string
	Description *string
	Position    *int
}

// UpdateList updates a list, verifying ownership.
func (s *ListService) UpdateList(ctx context.Context, input UpdateListInput) (*domain.UserList, error) {
	list, err := s.lists.GetListByID(ctx, input.ListID)
	if err != nil {
		return nil, err
	}
	if list.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this list")
	}

	if input.Name != nil {
		list.Name = *input.Name
	}
	if input.Description != nil {
		list.Description = input.Description
	}
	if input.Position != nil {
		list.Position = *input.Position
	}

	if err := s.lists.UpdateList(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

// DeleteList deletes a list, verifying ownership.
func (s *ListService) DeleteList(ctx context.Context, listID, userID uuid.UUID) error {
	list, err := s.lists.GetListByID(ctx, listID)
	if err != nil {
		return err
	}
	if list.UserID != userID {
		return domain.Forbidden("you do not own this list")
	}
	return s.lists.DeleteList(ctx, listID)
}

// --- Entry operations ---

// CreateEntryInput holds input for adding a company to a list.
type CreateEntryInput struct {
	ListID        uuid.UUID
	UserID        uuid.UUID
	CompanyID     uuid.UUID
	CompanyStatus domain.CompanyTrackingStatus
}

// CreateEntry adds a company to a list, verifying ownership.
func (s *ListService) CreateEntry(ctx context.Context, input CreateEntryInput) (*domain.ListEntry, error) {
	// Verify list ownership
	list, err := s.lists.GetListByID(ctx, input.ListID)
	if err != nil {
		return nil, err
	}
	if list.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this list")
	}

	if input.CompanyStatus == "" {
		input.CompanyStatus = domain.CompanyStatusMarked
	}

	// Get current entry count for position
	entries, err := s.lists.ListEntries(ctx, input.ListID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	entry := &domain.ListEntry{
		ID:            uuid.Must(uuid.NewV7()),
		ListID:        input.ListID,
		CompanyID:     input.CompanyID,
		CompanyStatus: input.CompanyStatus,
		Position:      len(entries),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.lists.CreateEntry(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// BatchCreateEntriesInput holds input for batch-adding companies to a list.
type BatchCreateEntriesInput struct {
	ListID     uuid.UUID
	UserID     uuid.UUID
	CompanyIDs []uuid.UUID
}

// BatchCreateEntries adds multiple companies to a list at once with default
// company_status "marked". Skips companies already in the list.
func (s *ListService) BatchCreateEntries(ctx context.Context, input BatchCreateEntriesInput) ([]*domain.ListEntry, error) {
	// Verify list ownership
	list, err := s.lists.GetListByID(ctx, input.ListID)
	if err != nil {
		return nil, err
	}
	if list.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this list")
	}

	// Get existing entries to determine positions and skip duplicates
	existing, err := s.lists.ListEntries(ctx, input.ListID)
	if err != nil {
		return nil, err
	}

	existingCompanies := make(map[uuid.UUID]bool, len(existing))
	for _, e := range existing {
		existingCompanies[e.CompanyID] = true
	}

	now := time.Now().UTC()
	position := len(existing)
	var created []*domain.ListEntry

	err = s.tx.WithTx(ctx, func(txCtx context.Context) error {
		for _, companyID := range input.CompanyIDs {
			if existingCompanies[companyID] {
				continue // skip duplicates
			}

			entry := &domain.ListEntry{
				ID:            uuid.Must(uuid.NewV7()),
				ListID:        input.ListID,
				CompanyID:     companyID,
				CompanyStatus: domain.CompanyStatusMarked,
				Position:      position,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			position++

			if err := s.lists.CreateEntry(txCtx, entry); err != nil {
				return err
			}

			created = append(created, entry)
			existingCompanies[companyID] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

// ListEntries returns all entries for a list, verifying ownership.
func (s *ListService) ListEntries(ctx context.Context, listID, userID uuid.UUID) ([]domain.ListEntry, error) {
	list, err := s.lists.GetListByID(ctx, listID)
	if err != nil {
		return nil, err
	}
	if list.UserID != userID {
		return nil, domain.Forbidden("you do not own this list")
	}
	return s.lists.ListEntries(ctx, listID)
}

// UpdateEntryInput holds input for updating a list entry (company_status and position only).
type UpdateEntryInput struct {
	EntryID       uuid.UUID
	UserID        uuid.UUID
	CompanyStatus *domain.CompanyTrackingStatus
	Position      *int
}

// UpdateEntry updates a list entry's company status or position.
func (s *ListService) UpdateEntry(ctx context.Context, input UpdateEntryInput) (*domain.ListEntry, error) {
	entry, err := s.lists.GetEntryByID(ctx, input.EntryID)
	if err != nil {
		return nil, err
	}

	// Verify ownership via list
	list, err := s.lists.GetListByID(ctx, entry.ListID)
	if err != nil {
		return nil, err
	}
	if list.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this list")
	}

	if input.CompanyStatus != nil {
		// Sync status across ALL of the user's lists that contain this company
		// so the status remains consistent regardless of which list is viewed.
		if err := s.lists.UpdateCompanyStatusForUser(ctx, input.UserID, entry.CompanyID, *input.CompanyStatus); err != nil {
			return nil, err
		}
		entry.CompanyStatus = *input.CompanyStatus
	}

	if input.Position != nil {
		entry.Position = *input.Position
		if err := s.lists.UpdateEntry(ctx, entry); err != nil {
			return nil, err
		}
	}

	return entry, nil
}

// DeleteEntry removes an entry, verifying ownership.
func (s *ListService) DeleteEntry(ctx context.Context, entryID, userID uuid.UUID) error {
	entry, err := s.lists.GetEntryByID(ctx, entryID)
	if err != nil {
		return err
	}

	list, err := s.lists.GetListByID(ctx, entry.ListID)
	if err != nil {
		return err
	}
	if list.UserID != userID {
		return domain.Forbidden("you do not own this list")
	}

	return s.lists.DeleteEntry(ctx, entryID)
}

// ListEntriesByCompanyID returns entries for a specific company across all
// of a user's lists. Used to show which lists contain a given company.
func (s *ListService) ListEntriesByCompanyID(ctx context.Context, userID, companyID uuid.UUID) ([]domain.ListEntryWithList, error) {
	return s.lists.ListEntriesByCompanyID(ctx, userID, companyID)
}

// SyncListEntriesInput holds input for syncing a list's company set.
type SyncListEntriesInput struct {
	ListID     uuid.UUID
	UserID     uuid.UUID
	CompanyIDs []uuid.UUID
}

// SyncListEntries reconciles the desired set of company IDs for a list.
// Companies in the array but not in the list are added. Companies in the list
// but not in the array are removed. Existing entries are left untouched.
func (s *ListService) SyncListEntries(ctx context.Context, input SyncListEntriesInput) ([]*domain.ListEntry, error) {
	// Verify list ownership
	list, err := s.lists.GetListByID(ctx, input.ListID)
	if err != nil {
		return nil, err
	}
	if list.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this list")
	}

	// Get existing entries
	existing, err := s.lists.ListEntries(ctx, input.ListID)
	if err != nil {
		return nil, err
	}

	existingByCompany := make(map[uuid.UUID]domain.ListEntry, len(existing))
	for _, e := range existing {
		existingByCompany[e.CompanyID] = e
	}

	desiredSet := make(map[uuid.UUID]bool, len(input.CompanyIDs))
	for _, cid := range input.CompanyIDs {
		desiredSet[cid] = true
	}

	// Compute diffs
	var toAdd []uuid.UUID
	for _, cid := range input.CompanyIDs {
		if _, exists := existingByCompany[cid]; !exists {
			toAdd = append(toAdd, cid)
		}
	}

	var toRemove []uuid.UUID
	for cid := range existingByCompany {
		if !desiredSet[cid] {
			toRemove = append(toRemove, cid)
		}
	}

	now := time.Now().UTC()
	position := len(existing)
	var result []*domain.ListEntry

	err = s.tx.WithTx(ctx, func(txCtx context.Context) error {
		// Remove entries no longer desired
		for _, cid := range toRemove {
			if err := s.lists.DeleteEntryByCompany(txCtx, input.ListID, cid); err != nil {
				return err
			}
		}

		// Add new entries
		for _, cid := range toAdd {
			entry := &domain.ListEntry{
				ID:            uuid.Must(uuid.NewV7()),
				ListID:        input.ListID,
				CompanyID:     cid,
				CompanyStatus: domain.CompanyStatusMarked,
				Position:      position,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			position++

			if err := s.lists.CreateEntry(txCtx, entry); err != nil {
				return err
			}

			result = append(result, entry)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListsWithCompanyFlag returns all lists for a user with a flag indicating
// whether each list contains the given company.
func (s *ListService) ListsWithCompanyFlag(ctx context.Context, userID, companyID uuid.UUID) ([]domain.ListCompanyFlag, error) {
	return s.lists.ListsWithCompanyFlag(ctx, userID, companyID)
}

// CompanyListCounts returns a map of company_id → list count for all companies
// across the user's lists. Used for the company card list indicator chip.
func (s *ListService) CompanyListCounts(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int, error) {
	return s.lists.CompanyListCounts(ctx, userID)
}

// DeleteEntryByCompany removes an entry by list ID and company ID, verifying ownership.
func (s *ListService) DeleteEntryByCompany(ctx context.Context, listID, companyID, userID uuid.UUID) error {
	list, err := s.lists.GetListByID(ctx, listID)
	if err != nil {
		return err
	}
	if list.UserID != userID {
		return domain.Forbidden("you do not own this list")
	}
	return s.lists.DeleteEntryByCompany(ctx, listID, companyID)
}
