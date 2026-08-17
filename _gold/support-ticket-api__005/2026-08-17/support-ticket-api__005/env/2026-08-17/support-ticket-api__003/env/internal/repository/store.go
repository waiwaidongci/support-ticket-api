package repository

import (
	"context"
	"time"

	"support-ticket-api/internal/model"
)

type Store interface {
	GetUserByID(ctx context.Context, id int64) (model.User, error)
	CreateTicket(ctx context.Context, params model.CreateTicketParams) (model.Ticket, error)
	GetTicketByID(ctx context.Context, id int64) (model.Ticket, error)
	ListTickets(ctx context.Context, filter model.TicketListFilter) ([]model.Ticket, error)
	UpdateStatus(ctx context.Context, params model.UpdateStatusParams) (model.Ticket, error)
	UpdateResolution(ctx context.Context, params model.UpdateResolutionParams) (model.Ticket, error)
	SetPriority(ctx context.Context, params model.SetPriorityParams) (model.Ticket, error)
	AssignTicket(ctx context.Context, ticketID, assigneeID int64) (model.Ticket, error)
	ClaimTicket(ctx context.Context, ticketID, assigneeID int64) (model.Ticket, error)
	ListHistory(ctx context.Context, ticketID int64) ([]model.TicketHistory, error)
	Statistics(ctx context.Context, now time.Time) (model.Statistics, error)
}
