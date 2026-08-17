package service

import (
	"fmt"

	"support-ticket-api/internal/model"
)

func validateCreateTicket(input *CreateTicketInput) error {
	normalizeCreateTicket(input)
	if input.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if !model.ValidTicketType(input.Type) {
		return fmt.Errorf("%w: unsupported type", ErrInvalidInput)
	}
	if !model.ValidPriority(input.Priority) {
		return fmt.Errorf("%w: unsupported priority", ErrInvalidInput)
	}
	return nil
}

func validateUpdateStatusInput(input *UpdateStatusInput) error {
	normalizeUpdateStatusInput(input)
	if !model.ValidStatus(input.ToStatus) {
		return fmt.Errorf("%w: unsupported status", ErrInvalidInput)
	}
	if input.ToStatus == model.StatusResolved && input.Result == "" {
		return fmt.Errorf("%w: result is required when resolving a ticket", ErrInvalidInput)
	}
	return nil
}

func validateUpdateResolutionInput(input *UpdateResolutionInput) error {
	normalizeUpdateResolutionInput(input)
	if input.Result == "" {
		return fmt.Errorf("%w: result is required", ErrInvalidInput)
	}
	return nil
}

func validateSetPriorityInput(input *SetPriorityInput) error {
	normalizeSetPriorityInput(input)
	if !model.ValidPriority(input.Priority) {
		return fmt.Errorf("%w: unsupported priority", ErrInvalidInput)
	}
	return nil
}
