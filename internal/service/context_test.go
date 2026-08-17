package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"support-ticket-api/internal/model"
)

type contextStore struct {
	*fakeStore
}

func (s *contextStore) ListTickets(ctx context.Context, _ model.TicketListFilter) ([]model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []model.Ticket{{ID: 1, SLADueAt: time.Now().Add(time.Hour)}}, nil
}

func TestListTicketsPropagatesCanceledContext(t *testing.T) {
	svc := New(&contextStore{&fakeStore{}})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.ListTickets(ctx, model.User{ID: 3, Role: model.RoleSupervisor}, model.TicketListFilter{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ListTickets error = %v, want context.Canceled", err)
	}
}
