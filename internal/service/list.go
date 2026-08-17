package service

import (
	"time"

	"support-ticket-api/internal/model"
	"support-ticket-api/internal/pkg/sla"
)

func buildTicketList(tickets []model.Ticket, now time.Time) []model.TicketListItem {
	items := make([]model.TicketListItem, 0, len(tickets))
	for _, ticket := range tickets {
		status, breached := sla.Status(now, ticket.SLADueAt)
		items = append(items, model.TicketListItem{
			Ticket:      ticket,
			SLAStatus:   status,
			SLABreached: breached,
		})
	}
	return items
}
