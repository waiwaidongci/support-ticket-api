package model

import "time"

const (
	TicketTypeBug       = "bug"
	TicketTypeQuestion  = "question"
	TicketTypeFeature   = "feature"
	TicketTypeComplaint = "complaint"
	TicketTypeOther     = "other"
)

const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusPending    = "pending"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"
)

var validTicketTypes = map[string]struct{}{
	TicketTypeBug:       {},
	TicketTypeQuestion:  {},
	TicketTypeFeature:   {},
	TicketTypeComplaint: {},
	TicketTypeOther:     {},
}

var validPriorities = map[string]struct{}{
	PriorityLow:    {},
	PriorityNormal: {},
	PriorityHigh:   {},
	PriorityUrgent: {},
}

var validStatuses = map[string]struct{}{
	StatusOpen:       {},
	StatusInProgress: {},
	StatusPending:    {},
	StatusResolved:   {},
	StatusClosed:     {},
}

func ValidTicketType(value string) bool {
	_, ok := validTicketTypes[value]
	return ok
}

func ValidPriority(value string) bool {
	_, ok := validPriorities[value]
	return ok
}

func ValidStatus(value string) bool {
	_, ok := validStatuses[value]
	return ok
}

type Ticket struct {
	ID          int64      `json:"id"`
	TicketNo    string     `json:"ticket_no"`
	CustomerID  int64      `json:"customer_id"`
	Type        string     `json:"type"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Result      string     `json:"result"`
	Note        string     `json:"note"`
	AssigneeID  *int64     `json:"assignee_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
	SLADueAt    time.Time  `json:"sla_due_at"`
}

type TicketListItem struct {
	Ticket
	SLAStatus   string `json:"sla_status"`
	SLABreached bool   `json:"sla_breached"`
}

type TicketHistory struct {
	ID         int64     `json:"id"`
	TicketID   int64     `json:"ticket_id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	OperatorID int64     `json:"operator_id"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

type TicketListFilter struct {
	Status     string
	Priority   string
	AssigneeID *int64
	CustomerID *int64
}

type CreateTicketParams struct {
	CustomerID  int64
	Type        string
	Priority    string
	Title       string
	Description string
}

type UpdateStatusParams struct {
	TicketID     int64
	ExpectedFrom string
	ToStatus     string
	OperatorID   int64
	Result       *string
	Note         *string
}

type UpdateResolutionParams struct {
	TicketID int64
	Result   string
	Note     string
}

type SetPriorityParams struct {
	TicketID   int64
	Priority   string
	OperatorID int64
}

type Statistics struct {
	Total       int64            `json:"total"`
	Open        int64            `json:"open"`
	InProgress  int64            `json:"in_progress"`
	Pending     int64            `json:"pending"`
	Resolved    int64            `json:"resolved"`
	Closed      int64            `json:"closed"`
	SLABreached int64            `json:"sla_breached"`
	ByStatus    map[string]int64 `json:"by_status"`
	ByPriority  map[string]int64 `json:"by_priority"`
	ByAssignee  map[string]int64 `json:"by_assignee"`
}
