package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// ApplicationService handles business logic for job applications (company-level).
type ApplicationService struct {
	apps domain.ApplicationRepository
	tx   domain.Transactor
}

// NewApplicationService creates a new ApplicationService.
func NewApplicationService(
	apps domain.ApplicationRepository,
	tx domain.Transactor,
) *ApplicationService {
	return &ApplicationService{apps: apps, tx: tx}
}

// --- Application CRUD ---

// CreateApplicationInput holds input for creating an application.
type CreateApplicationInput struct {
	UserID      uuid.UUID
	CompanyID   uuid.UUID
	RoleTitle   *string
	Status      domain.ApplicationStatus
	DateApplied *time.Time
	Notes       *string
}

// CreateApplication creates a new application for a company.
func (s *ApplicationService) CreateApplication(ctx context.Context, input CreateApplicationInput) (*domain.Application, error) {
	if input.Status == "" {
		input.Status = domain.StatusNotApplied
	}

	now := time.Now().UTC()
	app := &domain.Application{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      input.UserID,
		CompanyID:   input.CompanyID,
		RoleTitle:   input.RoleTitle,
		Status:      input.Status,
		DateApplied: input.DateApplied,
		Notes:       input.Notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.apps.Create(txCtx, app); err != nil {
			return err
		}

		// Record initial status in history
		history := &domain.StatusHistory{
			ID:            uuid.Must(uuid.NewV7()),
			ApplicationID: app.ID,
			FromStatus:    nil,
			ToStatus:      app.Status,
			ChangedAt:     now,
		}
		return s.apps.CreateStatusHistory(txCtx, history)
	})
	if err != nil {
		return nil, err
	}

	return app, nil
}

// ListApplications returns all applications for a user, optionally filtered by status.
func (s *ApplicationService) ListApplications(ctx context.Context, userID uuid.UUID, statusFilter *domain.ApplicationStatus, excludeNotApplied bool) ([]domain.ApplicationWithCompany, error) {
	return s.apps.ListByUser(ctx, userID, statusFilter, excludeNotApplied)
}

// ListByCompany returns all applications for a user+company.
func (s *ApplicationService) ListByCompany(ctx context.Context, userID, companyID uuid.UUID) ([]domain.Application, error) {
	return s.apps.ListByCompany(ctx, userID, companyID)
}

// UpdateApplicationInput holds input for updating an application.
type UpdateApplicationInput struct {
	ApplicationID uuid.UUID
	UserID        uuid.UUID
	RoleTitle     *string
	Status        *domain.ApplicationStatus
	DateApplied   *time.Time
	Notes         *string
}

// UpdateApplication updates an application, tracking status changes.
func (s *ApplicationService) UpdateApplication(ctx context.Context, input UpdateApplicationInput) (*domain.Application, error) {
	app, err := s.apps.GetByID(ctx, input.ApplicationID)
	if err != nil {
		return nil, err
	}
	if app.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this application")
	}

	oldStatus := app.Status

	if input.RoleTitle != nil {
		app.RoleTitle = input.RoleTitle
	}
	if input.Status != nil {
		app.Status = *input.Status
	}
	if input.DateApplied != nil {
		app.DateApplied = input.DateApplied
	}
	if input.Notes != nil {
		app.Notes = input.Notes
	}

	err = s.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.apps.Update(txCtx, app); err != nil {
			return err
		}

		// Track status change
		if input.Status != nil && oldStatus != *input.Status {
			history := &domain.StatusHistory{
				ID:            uuid.Must(uuid.NewV7()),
				ApplicationID: app.ID,
				FromStatus:    &oldStatus,
				ToStatus:      *input.Status,
				ChangedAt:     time.Now().UTC(),
			}
			return s.apps.CreateStatusHistory(txCtx, history)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return app, nil
}

// DeleteApplication removes an application, verifying ownership.
func (s *ApplicationService) DeleteApplication(ctx context.Context, applicationID, userID uuid.UUID) error {
	app, err := s.apps.GetByID(ctx, applicationID)
	if err != nil {
		return err
	}
	if app.UserID != userID {
		return domain.Forbidden("you do not own this application")
	}
	return s.apps.Delete(ctx, applicationID)
}

// GetApplicationHistory returns status change history for an application.
func (s *ApplicationService) GetApplicationHistory(ctx context.Context, applicationID, userID uuid.UUID) ([]domain.StatusHistory, error) {
	app, err := s.apps.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, domain.Forbidden("you do not own this application")
	}
	return s.apps.ListStatusHistory(ctx, applicationID)
}

// --- Interview Rounds ---

// CreateRoundInput holds input for creating an interview round.
type CreateRoundInput struct {
	ApplicationID uuid.UUID
	UserID        uuid.UUID
	RoundNumber   int
	RoundType     string
	ScheduledDate *time.Time
	Outcome       domain.InterviewOutcome
	Notes         *string
}

// CreateRound adds an interview round, verifying ownership.
func (s *ApplicationService) CreateRound(ctx context.Context, input CreateRoundInput) (*domain.InterviewRound, error) {
	app, err := s.apps.GetByID(ctx, input.ApplicationID)
	if err != nil {
		return nil, err
	}
	if app.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this application")
	}

	if input.Outcome == "" {
		input.Outcome = domain.OutcomePending
	}

	now := time.Now().UTC()
	round := &domain.InterviewRound{
		ID:            uuid.Must(uuid.NewV7()),
		ApplicationID: input.ApplicationID,
		RoundNumber:   input.RoundNumber,
		RoundType:     input.RoundType,
		ScheduledDate: input.ScheduledDate,
		Outcome:       input.Outcome,
		Notes:         input.Notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.apps.CreateInterviewRound(ctx, round); err != nil {
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
func (s *ApplicationService) UpdateRound(ctx context.Context, input UpdateRoundInput) (*domain.InterviewRound, error) {
	round, err := s.apps.GetInterviewRoundByID(ctx, input.RoundID)
	if err != nil {
		return nil, err
	}

	app, err := s.apps.GetByID(ctx, round.ApplicationID)
	if err != nil {
		return nil, err
	}
	if app.UserID != input.UserID {
		return nil, domain.Forbidden("you do not own this application")
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

	if err := s.apps.UpdateInterviewRound(ctx, round); err != nil {
		return nil, err
	}
	return round, nil
}

// DeleteRound removes an interview round, verifying ownership.
func (s *ApplicationService) DeleteRound(ctx context.Context, roundID, userID uuid.UUID) error {
	round, err := s.apps.GetInterviewRoundByID(ctx, roundID)
	if err != nil {
		return err
	}

	app, err := s.apps.GetByID(ctx, round.ApplicationID)
	if err != nil {
		return err
	}
	if app.UserID != userID {
		return domain.Forbidden("you do not own this application")
	}

	return s.apps.DeleteInterviewRound(ctx, roundID)
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

// GetDashboardCounts returns aggregated status counts from the applications table.
func (s *ApplicationService) GetDashboardCounts(ctx context.Context, userID uuid.UUID) (*StatusCounts, error) {
	statusMap, err := s.apps.CountByStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	counts := &StatusCounts{}
	for status, count := range statusMap {
		counts.Total += count
		switch status {
		case domain.StatusNotApplied:
			counts.NotApplied = count
		case domain.StatusApplied:
			counts.Applied = count
		case domain.StatusPhoneScreen:
			counts.PhoneScreen = count
		case domain.StatusInterview:
			counts.Interview = count
		case domain.StatusOffer:
			counts.Offer = count
		case domain.StatusRejected:
			counts.Rejected = count
		case domain.StatusAccepted:
			counts.Accepted = count
		case domain.StatusWithdrawn:
			counts.Withdrawn = count
		}
	}

	return counts, nil
}
