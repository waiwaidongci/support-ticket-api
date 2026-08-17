package service

import (
	"context"
	"errors"
	"testing"

	"support-ticket-api/internal/model"
)

func TestUpdateStatusTransitionErrorWrapsSentinel(t *testing.T) {
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
	})
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("UpdateStatus error = %v, want errors.Is(ErrInvalidTransition)", err)
	}
}
