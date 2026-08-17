package service

import "support-ticket-api/internal/model"

func normalizeTicketListItems(items []model.TicketListItem) []model.TicketListItem {
	if items == nil {
		return emptyTicketList()
	}
	return items
}

func normalizeHistoryList(history []model.TicketHistory) []model.TicketHistory {
	if history == nil {
		return emptyHistoryList()
	}
	return history
}

func cloneTicketListItem(item model.TicketListItem) model.TicketListItem {
	item.Ticket = cloneTicket(item.Ticket)
	return item
}

func cloneTicket(ticket model.Ticket) model.Ticket {
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
