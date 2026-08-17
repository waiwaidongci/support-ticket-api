package service

import "errors"

var (
	ErrForbidden         = errors.New("permission denied")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidTransition = errors.New("invalid status transition")
)
