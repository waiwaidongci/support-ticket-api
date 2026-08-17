package service

import (
	"fmt"
	"strings"

	"support-ticket-api/internal/model"
)

func validateCreateTicket(input *CreateTicketInput) error {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if !model.ValidTicketType(input.Type) {
		return fmt.Errorf("%w: unsupported type", ErrInvalidInput)
	}
	if input.Priority == "" {
		input.Priority = model.PriorityNormal
	}
	if !model.ValidPriority(input.Priority) {
		return fmt.Errorf("%w: unsupported priority", ErrInvalidInput)
	}
	return nil
}
