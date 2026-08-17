package service

import "fmt"

func claimTransitionError(status string) error {
	return fmt.Errorf("%w: ticket cannot be claimed from status %s", ErrInvalidTransition, status)
}

func invalidTransitionError(from, to string) error {
	return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
}

func resolveWithoutResultError() error {
	return fmt.Errorf("%w: result is required when resolving a ticket", ErrInvalidInput)
}
