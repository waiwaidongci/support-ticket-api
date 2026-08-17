package service

import "fmt"

func claimTransitionError(status string) error {
	return fmt.Errorf("%v: ticket cannot be claimed from status %s", ErrInvalidTransition, status)
}

func invalidTransitionError(from, to string) error {
	return fmt.Errorf("%v: %s -> %s", ErrInvalidTransition, from, to)
}

func resolveWithoutResultError() error {
	return fmt.Errorf("%v: result is required when resolving a ticket", ErrInvalidInput)
}
