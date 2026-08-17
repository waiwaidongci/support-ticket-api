package service

import (
	"context"
	"time"

	"support-ticket-api/internal/model"
	"support-ticket-api/internal/repository"
)

func loadTickets(ctx context.Context, repo repository.Store, filter model.TicketListFilter) ([]model.Ticket, error) {
	return repo.ListTickets(ctx, filter)
}

func loadTicket(ctx context.Context, repo repository.Store, id int64) (model.Ticket, error) {
	return repo.GetTicketByID(ctx, id)
}

func loadHistory(ctx context.Context, repo repository.Store, id int64) ([]model.TicketHistory, error) {
	return repo.ListHistory(ctx, id)
}

func loadStatistics(ctx context.Context, repo repository.Store, now time.Time) (model.Statistics, error) {
	return repo.Statistics(ctx, now)
}
