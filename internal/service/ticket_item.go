package service

import "support-ticket-api/internal/model"

func copyTicket(ticket model.Ticket) model.Ticket {
	if ticket.AssigneeID != nil {
		assigneeID := *ticket.AssigneeID
		ticket.AssigneeID = &assigneeID
	}
	if ticket.ResolvedAt != nil {
		resolvedAt := *ticket.ResolvedAt
		ticket.ResolvedAt = &resolvedAt
	}
	if ticket.ClosedAt != nil {
		closedAt := *ticket.ClosedAt
		ticket.ClosedAt = &closedAt
	}
	return ticket
}

func copyTicketList(tickets []model.Ticket) []model.Ticket {
	if tickets == nil {
		return make([]model.Ticket, 0)
	}
	copied := make([]model.Ticket, len(tickets))
	for i, ticket := range tickets {
		copied[i] = copyTicket(ticket)
	}
	return copied
}

func copyTicketListItem(item model.TicketListItem) model.TicketListItem {
	item.Ticket = copyTicket(item.Ticket)
	return item
}
