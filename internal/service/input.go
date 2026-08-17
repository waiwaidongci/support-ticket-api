package service

import (
	"strings"

	"support-ticket-api/internal/model"
)

func normalizeCreateTicket(input *CreateTicketInput) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Priority == "" {
		input.Priority = model.PriorityNormal
	}
}

func normalizeUpdateStatusInput(input *UpdateStatusInput) {
	input.Result = strings.TrimSpace(input.Result)
	input.Note = strings.TrimSpace(input.Note)
}

func normalizeUpdateResolutionInput(input *UpdateResolutionInput) {
	input.Result = strings.TrimSpace(input.Result)
	input.Note = strings.TrimSpace(input.Note)
}

func normalizeSetPriorityInput(input *SetPriorityInput) {
	input.Priority = strings.TrimSpace(input.Priority)
}
