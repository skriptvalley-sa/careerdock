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

// CreateEntryInput holds input for creating a list entry.
type CreateEntryInput struct {
	ListID        uuid.UUID
	UserID        uuid.UUID
	CompanyID     uuid.UUID
	CompanyStatus domain.CompanyTrackingStatus
	RoleTitle     *string
	Status        domain.ApplicationStatus
	DateApplied   *time.Time
	Notes         *string
}

// CreateEntry adds a company+role to a list, verifying ownership.
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
	if input.Status == "" {
		input.Status = domain.StatusNotApplied
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
		RoleTitle:     input.RoleTitle,
		Status:        input.Status,
		DateApplied:   input.DateApplied,
		Notes:         input.Notes,
		Position:      len(entries),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err = s.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.lists.CreateEntry(txCtx, entry); err != nil {
			return err
		}

		// Record initial status in history
		history := &domain.StatusHistory{
			ID:          uuid.Must(uuid.NewV7()),
			ListEntryID: entry.ID,
			FromStatus:  nil,
			ToStatus:    entry.Status,
			ChangedAt:   now,
		}
		return s.lists.CreateStatusHistory(txCtx, history)
	})
	if err != nil {
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
				Status:        domain.StatusNotApplied,
				Position:      position,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			position++

			if err := s.lists.CreateEntry(txCtx, entry); err != nil {
				return err
			}

			// Record initial status
			history := &domain.StatusHistory{
				ID:          uuid.Must(uuid.NewV7()),
				ListEntryID: entry.ID,
				FromStatus:  nil,
				ToStatus:    entry.Status,
				ChangedAt:   now,
			}
			if err := s.lists.CreateStatusHistory(txCtx, history); err != nil {
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

// UpdateEntryInput holds input for updating a list entry.
type UpdateEntryInput struct {
	EntryID       uuid.UUID
	UserID        uuid.UUID
	CompanyStatus *domain.CompanyTrackingStatus
	Status        *domain.ApplicationStatus
	RoleTitle     *string
	Notes         *string
	DateApplied   *time.Time
	Position      *int
}

// UpdateEntry updates a list entry, tracking status changes.
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

	oldStatus := entry.Status

	if input.CompanyStatus != nil {
		entry.CompanyStatus = *input.CompanyStatus
	}
	if input.Status != nil {
		entry.Status = *input.Status
	}
	if input.RoleTitle != nil {
		entry.RoleTitle = input.RoleTitle
	}
	if input.Notes != nil {
		entry.Notes = input.Notes
	}
	if input.DateApplied != nil {
		entry.DateApplied = input.DateApplied
	}
	if input.Position != nil {
		entry.Position = *input.Position
	}

	err = s.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.lists.UpdateEntry(txCtx, entry); err != nil {
			return err
		}

		// Track status change
		if input.Status != nil && oldStatus != *input.Status {
			history := &domain.StatusHistory{
				ID:          uuid.Must(uuid.NewV7()),
				ListEntryID: entry.ID,
				FromStatus:  &oldStatus,
				ToStatus:    *input.Status,
				ChangedAt:   time.Now().UTC(),
			}
			return s.lists.CreateStatusHistory(txCtx, history)
		}
		return nil
	})
	if err != nil {
		return nil, err
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

// GetEntryHistory returns status change history for an entry.
func (s *ListService) GetEntryHistory(ctx context.Context, entryID, userID uuid.UUID) ([]domain.StatusHistory, error) {
	entry, err := s.lists.GetEntryByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	list, err := s.lists.GetListByID(ctx, entry.ListID)
	if err != nil {
		return nil, err
	}
	if list.UserID != userID {
		return nil, domain.Forbidden("you do not own this list")
	}

	return s.lists.ListStatusHistory(ctx, entryID)
}

// --- Interview Round operations ---

// CreateRoundInput holds input for creating an interview round.
type CreateRoundInput struct {
	EntryID       uuid.UUID
	UserID        uuid.UUID
	RoundNumber   int
	RoundType     string
	ScheduledDate *time.Time
	Outcome       domain.InterviewOutcome
	Notes         *string
}

// CreateRound adds an interview round, verifying ownership.
func (s *ListService) CreateRound(ctx context.Context, input CreateRoundInput) (*domain.InterviewRound, error) {
	entry, err := s.lists.GetEntryByID(ctx, input.EntryID)
	if err != nil {
		return nil, err
	}

	list, err := s.lists.GetListByID(ctx, entry.ListID)
	if err != nil {
		return nil, err
	}
	if list.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this list")
	}

	if input.Outcome == "" {
		input.Outcome = domain.OutcomePending
	}

	now := time.Now().UTC()
	round := &domain.InterviewRound{
		ID:            uuid.Must(uuid.NewV7()),
		ListEntryID:   input.EntryID,
		RoundNumber:   input.RoundNumber,
		RoundType:     input.RoundType,
		ScheduledDate: input.ScheduledDate,
		Outcome:       input.Outcome,
		Notes:         input.Notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.lists.CreateInterviewRound(ctx, round); err != nil {
		return nil, err
	}
	return round, nil
}

// UpdateRoundInput holds input for updating an interview round.
type UpdateRoundInput struct {
	RoundID       uuid.UUID
	UserID        uuid.UUID
	Outcome       *domain.InterviewOutcome
	Notes         *string
	ScheduledDate *time.Time
}

// UpdateRound updates an interview round, verifying ownership.
func (s *ListService) UpdateRound(ctx context.Context, input UpdateRoundInput) (*domain.InterviewRound, error) {
	round, err := s.lists.GetInterviewRoundByID(ctx, input.RoundID)
	if err != nil {
		return nil, err
	}

	entry, err := s.lists.GetEntryByID(ctx, round.ListEntryID)
	if err != nil {
		return nil, err
	}

	list, err := s.lists.GetListByID(ctx, entry.ListID)
	if err != nil {
		return nil, err
	}
	if list.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this list")
	}

	if input.Outcome != nil {
		round.Outcome = *input.Outcome
	}
	if input.Notes != nil {
		round.Notes = input.Notes
	}
	if input.ScheduledDate != nil {
		round.ScheduledDate = input.ScheduledDate
	}

	if err := s.lists.UpdateInterviewRound(ctx, round); err != nil {
		return nil, err
	}
	return round, nil
}

// DeleteRound removes an interview round, verifying ownership.
func (s *ListService) DeleteRound(ctx context.Context, roundID, userID uuid.UUID) error {
	round, err := s.lists.GetInterviewRoundByID(ctx, roundID)
	if err != nil {
		return err
	}

	entry, err := s.lists.GetEntryByID(ctx, round.ListEntryID)
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

	return s.lists.DeleteInterviewRound(ctx, roundID)
}

// ListEntriesByCompanyID returns entries for a specific company across all
// of a user's lists. Used to show "your applications" on company pages.
func (s *ListService) ListEntriesByCompanyID(ctx context.Context, userID, companyID uuid.UUID) ([]domain.ListEntryWithList, error) {
	return s.lists.ListEntriesByCompanyID(ctx, userID, companyID)
}

// ListAllEntries returns all entries across all user lists, optionally filtered by status.
// When excludeNotApplied is true and no status filter is set, entries with status
// "not_applied" are excluded. Used for the "All Applications" page.
func (s *ListService) ListAllEntries(ctx context.Context, userID uuid.UUID, statusFilter *domain.ApplicationStatus, excludeNotApplied bool) ([]domain.ListEntryFull, error) {
	return s.lists.ListAllEntries(ctx, userID, statusFilter, excludeNotApplied)
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
				Status:        domain.StatusNotApplied,
				Position:      position,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			position++

			if err := s.lists.CreateEntry(txCtx, entry); err != nil {
				return err
			}

			// Record initial status
			history := &domain.StatusHistory{
				ID:          uuid.Must(uuid.NewV7()),
				ListEntryID: entry.ID,
				FromStatus:  nil,
				ToStatus:    entry.Status,
				ChangedAt:   now,
			}
			if err := s.lists.CreateStatusHistory(txCtx, history); err != nil {
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

// --- Dashboard helpers ---

// StatusCounts holds counts per application status for dashboard funnel.
type StatusCounts struct {
	NotApplied  int `json:"not_applied"`
	Applied     int `json:"applied"`
	PhoneScreen int `json:"phone_screen"`
	Interview   int `json:"interview"`
	Offer       int `json:"offer"`
	Rejected    int `json:"rejected"`
	Accepted    int `json:"accepted"`
	Withdrawn   int `json:"withdrawn"`
	Total       int `json:"total"`
}

// GetDashboardCounts returns aggregated status counts across all user lists.
func (s *ListService) GetDashboardCounts(ctx context.Context, userID uuid.UUID) (*StatusCounts, error) {
	lists, err := s.lists.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	counts := &StatusCounts{}
	for _, list := range lists {
		entries, err := s.lists.ListEntries(ctx, list.ID)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			counts.Total++
			switch e.Status {
			case domain.StatusNotApplied:
				counts.NotApplied++
			case domain.StatusApplied:
				counts.Applied++
			case domain.StatusPhoneScreen:
				counts.PhoneScreen++
			case domain.StatusInterview:
				counts.Interview++
			case domain.StatusOffer:
				counts.Offer++
			case domain.StatusRejected:
				counts.Rejected++
			case domain.StatusAccepted:
				counts.Accepted++
			case domain.StatusWithdrawn:
				counts.Withdrawn++
			}
		}
	}

	return counts, nil
}
