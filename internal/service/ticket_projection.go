package service

import (
	"time"

	"support-ticket-api/internal/model"
	"support-ticket-api/internal/pkg/sla"
)

func projectTicketListItem(ticket model.Ticket, now time.Time) model.TicketListItem {
	status, breached := sla.Status(now, ticket.SLADueAt)
	return model.TicketListItem{
		Ticket:      copyTicket(ticket),
		SLAStatus:   status,
		SLABreached: breached,
	}
}
