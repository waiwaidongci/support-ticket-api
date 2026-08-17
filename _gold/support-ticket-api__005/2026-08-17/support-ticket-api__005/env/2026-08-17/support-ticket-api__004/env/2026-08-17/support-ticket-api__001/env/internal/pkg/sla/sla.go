package sla

import (
	"time"

	"support-ticket-api/internal/model"
)

func DueAt(createdAt time.Time, priority string) time.Time {
	switch priority {
	case model.PriorityUrgent:
		return createdAt.Add(4 * time.Hour)
	case model.PriorityHigh:
		return createdAt.Add(8 * time.Hour)
	case model.PriorityLow:
		return createdAt.Add(48 * time.Hour)
	default:
		return createdAt.Add(24 * time.Hour)
	}
}

func Status(now, dueAt time.Time) (string, bool) {
	if now.After(dueAt) {
		return "breached", true
	}
	return "ok", false
}
