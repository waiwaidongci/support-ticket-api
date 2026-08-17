package model

import "time"

const (
	RoleCustomer   = "customer"
	RoleAgent      = "agent"
	RoleSupervisor = "supervisor"
)

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func ValidRole(role string) bool {
	switch role {
	case RoleCustomer, RoleAgent, RoleSupervisor:
		return true
	default:
		return false
	}
}
