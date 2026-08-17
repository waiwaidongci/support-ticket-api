package service

import (
	"time"

	"support-ticket-api/internal/model"
)

func buildTicketList(tickets []model.Ticket, now time.Time) []model.TicketListItem {
	items := make([]model.TicketListItem, 0, len(tickets))
	for _, ticket := range copyTicketList(tickets) {
		items = append(items, projectTicketListItem(ticket, now))
	}
	return items
}
