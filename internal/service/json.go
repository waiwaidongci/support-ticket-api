package service

import "support-ticket-api/internal/model"

func emptyTicketList() []model.TicketListItem {
	return make([]model.TicketListItem, 0)
}

func emptyHistoryList() []model.TicketHistory {
	return make([]model.TicketHistory, 0)
}

func normalizeStatistics(stats model.Statistics) model.Statistics {
	if stats.ByStatus == nil {
		stats.ByStatus = make(map[string]int64)
	}
	if stats.ByPriority == nil {
		stats.ByPriority = make(map[string]int64)
	}
	if stats.ByAssignee == nil {
		stats.ByAssignee = make(map[string]int64)
	}
	return stats
}
