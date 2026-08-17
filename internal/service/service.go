package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"support-ticket-api/internal/model"
	"support-ticket-api/internal/pkg/sla"
	"support-ticket-api/internal/repository"
)

type Service struct {
	repo repository.Store
	now  func() time.Time
}

func New(repo repository.Store) *Service {
	return &Service{repo: repo, now: time.Now}
}

func NewWithClock(repo repository.Store, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

func (s *Service) CreateTicket(ctx context.Context, user model.User, input CreateTicketInput) (model.Ticket, error) {
	if user.Role != model.RoleCustomer {
		return model.Ticket{}, ErrForbidden
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return model.Ticket{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if !model.ValidTicketType(input.Type) {
		return model.Ticket{}, fmt.Errorf("%w: unsupported type", ErrInvalidInput)
	}
	if input.Priority == "" {
		input.Priority = model.PriorityNormal
	}
	if !model.ValidPriority(input.Priority) {
		return model.Ticket{}, fmt.Errorf("%w: unsupported priority", ErrInvalidInput)
	}
	return s.repo.CreateTicket(ctx, model.CreateTicketParams{
		CustomerID:  user.ID,
		Type:        input.Type,
		Priority:    input.Priority,
		Title:       input.Title,
		Description: input.Description,
	})
}

func (s *Service) ListTickets(ctx context.Context, user model.User, filter model.TicketListFilter) ([]model.TicketListItem, error) {
	if user.Role == model.RoleCustomer {
		filter.CustomerID = &user.ID
	}
	tickets, err := s.repo.ListTickets(ctx, filter)
	if err != nil {
		return nil, err
	}
	now := s.now()
	items := make([]model.TicketListItem, 0, len(tickets))
	for _, ticket := range tickets {
		status, breached := sla.Status(now, ticket.SLADueAt)
		items = append(items, model.TicketListItem{
			Ticket:      ticket,
			SLAStatus:   status,
			SLABreached: breached,
		})
	}
	return items, nil
}

func (s *Service) GetTicket(ctx context.Context, user model.User, ticketID int64) (model.Ticket, error) {
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return model.Ticket{}, err
	}
	if err := s.ensureCanView(user, ticket); err != nil {
		return model.Ticket{}, err
	}
	return ticket, nil
}

func (s *Service) AssignTicket(ctx context.Context, operator model.User, ticketID, assigneeID int64) (model.Ticket, error) {
	if operator.Role != model.RoleSupervisor {
		return model.Ticket{}, ErrForbidden
	}
	assignee, err := s.repo.GetUserByID(ctx, assigneeID)
	if err != nil {
		return model.Ticket{}, err
	}
	if assignee.Role != model.RoleAgent {
		return model.Ticket{}, fmt.Errorf("%w: assignee must be an agent", ErrInvalidInput)
	}
	if _, err := s.repo.GetTicketByID(ctx, ticketID); err != nil {
		return model.Ticket{}, err
	}
	return s.repo.AssignTicket(ctx, ticketID, assigneeID)
}

func (s *Service) ClaimTicket(ctx context.Context, agent model.User, ticketID int64) (model.Ticket, error) {
	if agent.Role != model.RoleAgent {
		return model.Ticket{}, ErrForbidden
	}
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return model.Ticket{}, err
	}
	if !isClaimableStatus(ticket.Status) {
		return model.Ticket{}, claimTransitionError(ticket.Status)
	}
	return s.repo.ClaimTicket(ctx, ticketID, agent.ID)
}

func (s *Service) UpdateStatus(ctx context.Context, operator model.User, input UpdateStatusInput) (model.Ticket, error) {
	ticket, err := s.repo.GetTicketByID(ctx, input.TicketID)
	if err != nil {
		return model.Ticket{}, err
	}
	if err := s.ensureCanOperate(operator, ticket); err != nil {
		return model.Ticket{}, err
	}
	if !model.ValidStatus(input.ToStatus) {
		return model.Ticket{}, fmt.Errorf("%w: unsupported status", ErrInvalidInput)
	}
	if !canTransition(ticket.Status, input.ToStatus) {
		return model.Ticket{}, fmt.Errorf("%v: %s -> %s", ErrInvalidTransition, ticket.Status, input.ToStatus)
	}
	if input.ToStatus == model.StatusResolved && strings.TrimSpace(input.Result) == "" {
		return model.Ticket{}, resolveWithoutResultError()
	}
	params := model.UpdateStatusParams{
		TicketID:     input.TicketID,
		ExpectedFrom: ticket.Status,
		ToStatus:     input.ToStatus,
		OperatorID:   operator.ID,
		Result:       stringPointerOrNil(input.Result),
		Note:         stringPointerOrNil(input.Note),
	}
	return s.repo.UpdateStatus(ctx, params)
}

func (s *Service) UpdateResolution(ctx context.Context, operator model.User, input UpdateResolutionInput) (model.Ticket, error) {
	ticket, err := s.repo.GetTicketByID(ctx, input.TicketID)
	if err != nil {
		return model.Ticket{}, err
	}
	if err := s.ensureCanOperate(operator, ticket); err != nil {
		return model.Ticket{}, err
	}
	input.Result = strings.TrimSpace(input.Result)
	input.Note = strings.TrimSpace(input.Note)
	if input.Result == "" {
		return model.Ticket{}, fmt.Errorf("%w: result is required", ErrInvalidInput)
	}
	return s.repo.UpdateResolution(ctx, model.UpdateResolutionParams{
		TicketID: input.TicketID,
		Result:   input.Result,
		Note:     input.Note,
	})
}

func (s *Service) SetPriority(ctx context.Context, supervisor model.User, input SetPriorityInput) (model.Ticket, error) {
	if supervisor.Role != model.RoleSupervisor {
		return model.Ticket{}, ErrForbidden
	}
	if !model.ValidPriority(input.Priority) {
		return model.Ticket{}, fmt.Errorf("%w: unsupported priority", ErrInvalidInput)
	}
	if _, err := s.repo.GetTicketByID(ctx, input.TicketID); err != nil {
		return model.Ticket{}, err
	}
	return s.repo.SetPriority(ctx, model.SetPriorityParams{
		TicketID:   input.TicketID,
		Priority:   input.Priority,
		OperatorID: supervisor.ID,
	})
}

func (s *Service) History(ctx context.Context, user model.User, ticketID int64) ([]model.TicketHistory, error) {
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCanView(user, ticket); err != nil {
		return nil, err
	}
	return s.repo.ListHistory(ctx, ticketID)
}

func (s *Service) Statistics(ctx context.Context, user model.User) (model.Statistics, error) {
	if user.Role != model.RoleSupervisor {
		return model.Statistics{}, ErrForbidden
	}
	return s.repo.Statistics(ctx, s.now())
}

func (s *Service) ensureCanView(user model.User, ticket model.Ticket) error {
	if user.Role == model.RoleCustomer && ticket.CustomerID != user.ID {
		return ErrForbidden
	}
	return nil
}

func (s *Service) ensureCanOperate(user model.User, ticket model.Ticket) error {
	switch user.Role {
	case model.RoleSupervisor:
		return nil
	case model.RoleAgent:
		if ticket.AssigneeID == nil || *ticket.AssigneeID != user.ID {
			return ErrForbidden
		}
		return nil
	default:
		return ErrForbidden
	}
}

var allowedTransitions = map[string]map[string]struct{}{
	model.StatusOpen: {
		model.StatusInProgress: {},
		model.StatusPending:    {},
		model.StatusResolved:   {},
		model.StatusClosed:     {},
	},
	model.StatusInProgress: {
		model.StatusPending:  {},
		model.StatusResolved: {},
		model.StatusClosed:   {},
	},
	model.StatusPending: {
		model.StatusInProgress: {},
		model.StatusResolved:   {},
		model.StatusClosed:     {},
	},
	model.StatusResolved: {
		model.StatusClosed: {},
		model.StatusOpen:   {},
	},
	model.StatusClosed: {
		model.StatusOpen: {},
	},
}

func canTransition(from, to string) bool {
	targets, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	_, ok = targets[to]
	return ok
}

func isClaimableStatus(status string) bool {
	return status == model.StatusOpen || status == model.StatusPending
}

func stringPointerOrNil(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	value = strings.TrimSpace(value)
	return &value
}

type CreateTicketInput struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateStatusInput struct {
	TicketID int64  `json:"-"`
	ToStatus string `json:"status"`
	Result   string `json:"result"`
	Note     string `json:"note"`
}

type UpdateResolutionInput struct {
	TicketID int64  `json:"-"`
	Result   string `json:"result"`
	Note     string `json:"note"`
}

type SetPriorityInput struct {
	TicketID int64  `json:"-"`
	Priority string `json:"priority"`
}
