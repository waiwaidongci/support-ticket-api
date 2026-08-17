package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"support-ticket-api/internal/model"
	"support-ticket-api/internal/repository"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from, to string
		want     bool
	}{
		{model.StatusOpen, model.StatusInProgress, true},
		{model.StatusOpen, model.StatusResolved, true},
		{model.StatusInProgress, model.StatusResolved, true},
		{model.StatusResolved, model.StatusClosed, true},
		{model.StatusClosed, model.StatusOpen, true},
		{model.StatusInProgress, model.StatusOpen, false},
		{model.StatusClosed, model.StatusInProgress, false},
	}
	for _, tt := range tests {
		if got := canTransition(tt.from, tt.to); got != tt.want {
			t.Fatalf("canTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestServiceCreateTicket(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)

	user := model.User{ID: 1, Role: model.RoleCustomer}
	ticket, err := svc.CreateTicket(context.Background(), user, CreateTicketInput{
		Type:        model.TicketTypeBug,
		Priority:    model.PriorityHigh,
		Title:       "Cannot login",
		Description: "The login page returns 500.",
	})
	if err != nil {
		t.Fatalf("CreateTicket returned error: %v", err)
	}
	if ticket.Status != model.StatusOpen {
		t.Fatalf("expected open status, got %s", ticket.Status)
	}
	if store.created.CustomerID != user.ID || store.created.Priority != model.PriorityHigh {
		t.Fatalf("unexpected create params: %+v", store.created)
	}

	_, err = svc.CreateTicket(context.Background(), model.User{ID: 2, Role: model.RoleAgent}, CreateTicketInput{
		Type:  model.TicketTypeBug,
		Title: "not allowed",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestServiceListTicketsMarksSLA(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{
		tickets: []model.Ticket{
			{ID: 1, Status: model.StatusOpen, Priority: model.PriorityNormal, SLADueAt: now.Add(-time.Minute), CreatedAt: now.Add(-time.Hour)},
			{ID: 2, Status: model.StatusOpen, Priority: model.PriorityNormal, SLADueAt: now.Add(time.Hour), CreatedAt: now},
		},
	}
	svc := New(store)
	svc.now = func() time.Time { return now }

	items, err := svc.ListTickets(context.Background(), model.User{ID: 3, Role: model.RoleSupervisor}, model.TicketListFilter{})
	if err != nil {
		t.Fatalf("ListTickets returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if !items[0].SLABreached {
		t.Fatal("expected first ticket to be SLA breached")
	}
	if items[1].SLABreached {
		t.Fatal("expected second ticket not to be SLA breached")
	}
}

func TestServiceAssignTicketRequiresAgent(t *testing.T) {
	store := &fakeStore{
		user: model.User{ID: 1, Role: model.RoleCustomer},
	}
	svc := New(store)

	_, err := svc.AssignTicket(context.Background(), model.User{ID: 3, Role: model.RoleSupervisor}, 1, 1)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceUpdateStatusValidatesTransitionAndResult(t *testing.T) {
	store := &fakeStore{
		ticket: model.Ticket{
			ID:         1,
			Status:     model.StatusInProgress,
			CustomerID: 1,
		},
	}
	svc := New(store)

	_, err := svc.UpdateStatus(context.Background(), model.User{ID: 3, Role: model.RoleSupervisor}, UpdateStatusInput{
		TicketID: 1,
		ToStatus: model.StatusOpen,
		Result:   "waiting for customer",
		Note:     "asked for logs",
	})
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}

	_, err = svc.UpdateStatus(context.Background(), model.User{ID: 3, Role: model.RoleSupervisor}, UpdateStatusInput{
		TicketID: 1,
		ToStatus: model.StatusResolved,
		Result:   "",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for missing result, got %v", err)
	}
}

type fakeStore struct {
	created model.CreateTicketParams
	tickets []model.Ticket
	ticket  model.Ticket
	user    model.User
}

func (f *fakeStore) GetUserByID(_ context.Context, id int64) (model.User, error) {
	if f.user.ID == id {
		return f.user, nil
	}
	return model.User{ID: id, Name: "fake", Role: model.RoleAgent}, nil
}

func (f *fakeStore) CreateTicket(_ context.Context, params model.CreateTicketParams) (model.Ticket, error) {
	f.created = params
	return model.Ticket{
		ID:         1,
		CustomerID: params.CustomerID,
		Type:       params.Type,
		Priority:   params.Priority,
		Status:     model.StatusOpen,
		Title:      params.Title,
		SLADueAt:   time.Now().Add(24 * time.Hour),
	}, nil
}

func (f *fakeStore) GetTicketByID(_ context.Context, id int64) (model.Ticket, error) {
	if f.ticket.ID == id {
		return f.ticket, nil
	}
	for _, ticket := range f.tickets {
		if ticket.ID == id {
			return ticket, nil
		}
	}
	return model.Ticket{}, repository.ErrNotFound
}

func (f *fakeStore) ListTickets(_ context.Context, _ model.TicketListFilter) ([]model.Ticket, error) {
	return append([]model.Ticket(nil), f.tickets...), nil
}

func (f *fakeStore) UpdateStatus(_ context.Context, params model.UpdateStatusParams) (model.Ticket, error) {
	ticket := f.ticket
	ticket.Status = params.ToStatus
	if params.Result != nil {
		ticket.Result = *params.Result
	}
	if params.Note != nil {
		ticket.Note = *params.Note
	}
	return ticket, nil
}

func (f *fakeStore) UpdateResolution(_ context.Context, params model.UpdateResolutionParams) (model.Ticket, error) {
	ticket := f.ticket
	ticket.Result = params.Result
	ticket.Note = params.Note
	return ticket, nil
}

func (f *fakeStore) SetPriority(_ context.Context, params model.SetPriorityParams) (model.Ticket, error) {
	ticket := f.ticket
	ticket.Priority = params.Priority
	return ticket, nil
}

func (f *fakeStore) AssignTicket(_ context.Context, ticketID, assigneeID int64) (model.Ticket, error) {
	ticket := f.ticket
	ticket.ID = ticketID
	ticket.AssigneeID = &assigneeID
	return ticket, nil
}

func (f *fakeStore) ClaimTicket(_ context.Context, ticketID, assigneeID int64) (model.Ticket, error) {
	ticket := f.ticket
	ticket.ID = ticketID
	ticket.AssigneeID = &assigneeID
	ticket.Status = model.StatusInProgress
	return ticket, nil
}

func (f *fakeStore) ListHistory(_ context.Context, _ int64) ([]model.TicketHistory, error) {
	return []model.TicketHistory{}, nil
}

func (f *fakeStore) Statistics(_ context.Context, _ time.Time) (model.Statistics, error) {
	return model.Statistics{ByStatus: map[string]int64{}, ByPriority: map[string]int64{}, ByAssignee: map[string]int64{}}, nil
}
