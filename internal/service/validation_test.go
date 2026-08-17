package service

import (
	"context"
	"errors"
	"testing"

	"support-ticket-api/internal/model"
)

func TestCreateTicketValidationWrapsInvalidInput(t *testing.T) {
	svc := New(&fakeStore{})
	customer := model.User{ID: 1, Role: model.RoleCustomer}

	cases := []CreateTicketInput{
		{Type: model.TicketTypeBug, Priority: model.PriorityHigh, Title: "   "},
		{Type: "unknown", Priority: model.PriorityHigh, Title: "Cannot login"},
		{Type: model.TicketTypeBug, Priority: "critical", Title: "Cannot login"},
	}

	for _, input := range cases {
		_, err := svc.CreateTicket(context.Background(), customer, input)
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("CreateTicket(%+v) error = %v, want errors.Is(ErrInvalidInput)", input, err)
		}
	}
}
