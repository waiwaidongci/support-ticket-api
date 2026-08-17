package repository

import "errors"

var (
	ErrNotFound        = errors.New("record not found")
	ErrAlreadyAssigned = errors.New("ticket is already assigned to another user")
	ErrConflict        = errors.New("ticket state changed, please retry")
)
