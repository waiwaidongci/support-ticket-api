package service

import (
	"context"
	"testing"
	"time"

	"support-ticket-api/internal/model"
)

func TestListTicketsBuildsExactItems(t *testing.T) {
	now := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	store := &fakeStore{
		tickets: []model.Ticket{
			{ID: 1, Status: model.StatusOpen, Priority: model.PriorityHigh, SLADueAt: now.Add(time.Hour)},
			{ID: 2, Status: model.StatusOpen, Priority: model.PriorityNormal, SLADueAt: now.Add(2 * time.Hour)},
		},
	}
	svc := New(store)
	svc.now = func() time.Time { return now }

	items, err := svc.ListTickets(context.Background(), model.User{ID: 3, Role: model.RoleSupervisor}, model.TicketListFilter{})
	if err != nil {
		t.Fatalf("ListTickets returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d: %+v", len(items), items)
	}
	if items[0].ID != 1 || items[1].ID != 2 {
		t.Fatalf("unexpected item order or content: %+v", items)
	}
}
