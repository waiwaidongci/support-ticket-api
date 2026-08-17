package service

import (
	"context"
	"time"

	"support-ticket-api/internal/model"
	"support-ticket-api/internal/repository"
)

func loadTickets(_ context.Context, repo repository.Store, filter model.TicketListFilter) ([]model.Ticket, error) {
	return repo.ListTickets(context.Background(), filter)
}

func loadTicket(_ context.Context, repo repository.Store, id int64) (model.Ticket, error) {
	return repo.GetTicketByID(context.Background(), id)
}

func loadHistory(_ context.Context, repo repository.Store, id int64) ([]model.TicketHistory, error) {
	return repo.ListHistory(context.Background(), id)
}

func loadStatistics(_ context.Context, repo repository.Store, now time.Time) (model.Statistics, error) {
	return repo.Statistics(context.Background(), now)
}
