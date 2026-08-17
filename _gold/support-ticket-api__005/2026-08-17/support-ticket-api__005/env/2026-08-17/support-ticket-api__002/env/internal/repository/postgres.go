package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"support-ticket-api/internal/model"
	"support-ticket-api/internal/pkg/sla"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (s *PostgresStore) GetUserByID(ctx context.Context, id int64) (model.User, error) {
	row := s.pool.QueryRow(ctx, `SELECT id, name, role, created_at FROM users WHERE id=$1`, id)
	var user model.User
	if err := row.Scan(&user.ID, &user.Name, &user.Role, &user.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}
	return user, nil
}

func (s *PostgresStore) CreateTicket(ctx context.Context, params model.CreateTicketParams) (model.Ticket, error) {
	dueAt := sla.DueAt(time.Now(), params.Priority)
	row := s.pool.QueryRow(ctx, `
		INSERT INTO tickets (customer_id, type, priority, status, title, description, sla_due_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, ticket_no, customer_id, type, priority, status, title, description,
			result, note, assignee_id, created_at, updated_at, resolved_at, closed_at, sla_due_at`,
		params.CustomerID, params.Type, params.Priority, model.StatusOpen, params.Title, params.Description, dueAt,
	)
	return scanTicket(row)
}

func (s *PostgresStore) GetTicketByID(ctx context.Context, id int64) (model.Ticket, error) {
	row := s.pool.QueryRow(ctx, ticketSelect+" WHERE id=$1", id)
	return scanTicket(row)
}

func (s *PostgresStore) ListTickets(ctx context.Context, filter model.TicketListFilter) ([]model.Ticket, error) {
	where := []string{"1=1"}
	args := []any{}
	argIndex := 1
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status=$%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}
	if filter.Priority != "" {
		where = append(where, fmt.Sprintf("priority=$%d", argIndex))
		args = append(args, filter.Priority)
		argIndex++
	}
	if filter.AssigneeID != nil {
		where = append(where, fmt.Sprintf("assignee_id=$%d", argIndex))
		args = append(args, *filter.AssigneeID)
		argIndex++
	}
	if filter.CustomerID != nil {
		where = append(where, fmt.Sprintf("customer_id=$%d", argIndex))
		args = append(args, *filter.CustomerID)
		argIndex++
	}
	query := ticketSelect + " WHERE " + strings.Join(where, " AND ") + " ORDER BY created_at DESC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := make([]model.Ticket, 0)
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	return tickets, rows.Err()
}

func (s *PostgresStore) UpdateStatus(ctx context.Context, params model.UpdateStatusParams) (model.Ticket, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Ticket{}, err
	}
	defer tx.Rollback(ctx)

	var current string
	err = tx.QueryRow(ctx, `SELECT status FROM tickets WHERE id=$1 FOR UPDATE`, params.TicketID).Scan(&current)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.Ticket{}, ErrNotFound
		}
		return model.Ticket{}, err
	}
	if current != params.ExpectedFrom {
		return model.Ticket{}, ErrConflict
	}

	result := nullableString(params.Result)
	note := nullableString(params.Note)
	row := tx.QueryRow(ctx, `
		UPDATE tickets
		SET status=$2,
			updated_at=now(),
			result=COALESCE($3, result),
			note=COALESCE($4, note),
			resolved_at=CASE
				WHEN $2='resolved' THEN now()
				WHEN $2 IN ('open','in_progress','pending') THEN NULL
				ELSE resolved_at
			END,
			closed_at=CASE
				WHEN $2='closed' THEN now()
				WHEN $2 IN ('open','in_progress','pending') THEN NULL
				ELSE closed_at
			END
		WHERE id=$1
		RETURNING id, ticket_no, customer_id, type, priority, status, title, description,
			result, note, assignee_id, created_at, updated_at, resolved_at, closed_at, sla_due_at`,
		params.TicketID, params.ToStatus, result, note,
	)
	ticket, err := scanTicket(row)
	if err != nil {
		return model.Ticket{}, err
	}

	historyNote := ""
	if params.Note != nil {
		historyNote = *params.Note
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO ticket_status_history (ticket_id, from_status, to_status, operator_id, note)
		VALUES ($1, $2, $3, $4, $5)`,
		params.TicketID, current, params.ToStatus, params.OperatorID, historyNote,
	)
	if err != nil {
		return model.Ticket{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Ticket{}, err
	}
	return ticket, nil
}

func (s *PostgresStore) UpdateResolution(ctx context.Context, params model.UpdateResolutionParams) (model.Ticket, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE tickets
		SET result=$2, note=$3, updated_at=now()
		WHERE id=$1
		RETURNING id, ticket_no, customer_id, type, priority, status, title, description,
			result, note, assignee_id, created_at, updated_at, resolved_at, closed_at, sla_due_at`,
		params.TicketID, params.Result, params.Note,
	)
	ticket, err := scanTicket(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.Ticket{}, ErrNotFound
		}
		return model.Ticket{}, err
	}
	return ticket, nil
}

func (s *PostgresStore) SetPriority(ctx context.Context, params model.SetPriorityParams) (model.Ticket, error) {
	ticket, err := s.GetTicketByID(ctx, params.TicketID)
	if err != nil {
		return model.Ticket{}, err
	}
	dueAt := sla.DueAt(ticket.CreatedAt, params.Priority)
	row := s.pool.QueryRow(ctx, `
		UPDATE tickets
		SET priority=$2, sla_due_at=$3, updated_at=now()
		WHERE id=$1
		RETURNING id, ticket_no, customer_id, type, priority, status, title, description,
			result, note, assignee_id, created_at, updated_at, resolved_at, closed_at, sla_due_at`,
		params.TicketID, params.Priority, dueAt,
	)
	return scanTicket(row)
}

func (s *PostgresStore) AssignTicket(ctx context.Context, ticketID, assigneeID int64) (model.Ticket, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Ticket{}, err
	}
	defer tx.Rollback(ctx)

	var currentAssignee *int64
	err = tx.QueryRow(ctx, `SELECT assignee_id FROM tickets WHERE id=$1 FOR UPDATE`, ticketID).Scan(&currentAssignee)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.Ticket{}, ErrNotFound
		}
		return model.Ticket{}, err
	}
	if currentAssignee != nil && *currentAssignee != assigneeID {
		return model.Ticket{}, ErrAlreadyAssigned
	}

	row := tx.QueryRow(ctx, `
		UPDATE tickets SET assignee_id=$2, updated_at=now()
		WHERE id=$1
		RETURNING id, ticket_no, customer_id, type, priority, status, title, description,
			result, note, assignee_id, created_at, updated_at, resolved_at, closed_at, sla_due_at`,
		ticketID, assigneeID,
	)
	ticket, err := scanTicket(row)
	if err != nil {
		return model.Ticket{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Ticket{}, err
	}
	return ticket, nil
}

func (s *PostgresStore) ClaimTicket(ctx context.Context, ticketID, assigneeID int64) (model.Ticket, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Ticket{}, err
	}
	defer tx.Rollback(ctx)

	var currentAssignee *int64
	var currentStatus string
	err = tx.QueryRow(ctx, `SELECT assignee_id, status FROM tickets WHERE id=$1 FOR UPDATE`, ticketID).Scan(&currentAssignee, &currentStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.Ticket{}, ErrNotFound
		}
		return model.Ticket{}, err
	}
	if currentAssignee != nil && *currentAssignee != assigneeID {
		return model.Ticket{}, ErrAlreadyAssigned
	}

	row := tx.QueryRow(ctx, `
		UPDATE tickets
		SET assignee_id=$2, status=$3, updated_at=now()
		WHERE id=$1
		RETURNING id, ticket_no, customer_id, type, priority, status, title, description,
			result, note, assignee_id, created_at, updated_at, resolved_at, closed_at, sla_due_at`,
		ticketID, assigneeID, model.StatusInProgress,
	)
	ticket, err := scanTicket(row)
	if err != nil {
		return model.Ticket{}, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO ticket_status_history (ticket_id, from_status, to_status, operator_id, note)
		VALUES ($1, $2, $3, $4, '')`,
		ticketID, currentStatus, model.StatusInProgress, assigneeID,
	)
	if err != nil {
		return model.Ticket{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Ticket{}, err
	}
	return ticket, nil
}

func (s *PostgresStore) ListHistory(ctx context.Context, ticketID int64) ([]model.TicketHistory, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, ticket_id, from_status, to_status, operator_id, note, created_at
		FROM ticket_status_history
		WHERE ticket_id=$1
		ORDER BY created_at ASC, id ASC`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]model.TicketHistory, 0)
	for rows.Next() {
		var item model.TicketHistory
		if err := rows.Scan(&item.ID, &item.TicketID, &item.FromStatus, &item.ToStatus, &item.OperatorID, &item.Note, &item.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, item)
	}
	return history, rows.Err()
}

func (s *PostgresStore) Statistics(ctx context.Context, now time.Time) (model.Statistics, error) {
	stats := model.Statistics{
		ByStatus:   map[string]int64{},
		ByPriority: map[string]int64{},
		ByAssignee: map[string]int64{},
	}
	err := s.pool.QueryRow(ctx, `
		SELECT count(*),
			count(*) FILTER (WHERE status='open'),
			count(*) FILTER (WHERE status='in_progress'),
			count(*) FILTER (WHERE status='pending'),
			count(*) FILTER (WHERE status='resolved'),
			count(*) FILTER (WHERE status='closed'),
			count(*) FILTER (WHERE sla_due_at < $1)
		FROM tickets`, now).Scan(
		&stats.Total, &stats.Open, &stats.InProgress, &stats.Pending, &stats.Resolved, &stats.Closed, &stats.SLABreached,
	)
	if err != nil {
		return model.Statistics{}, err
	}

	statusRows, err := s.pool.Query(ctx, `SELECT status, count(*) FROM tickets GROUP BY status`)
	if err != nil {
		return model.Statistics{}, err
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var status string
		var count int64
		if err := statusRows.Scan(&status, &count); err != nil {
			return model.Statistics{}, err
		}
		stats.ByStatus[status] = count
	}

	priorityRows, err := s.pool.Query(ctx, `SELECT priority, count(*) FROM tickets GROUP BY priority`)
	if err != nil {
		return model.Statistics{}, err
	}
	defer priorityRows.Close()
	for priorityRows.Next() {
		var priority string
		var count int64
		if err := priorityRows.Scan(&priority, &count); err != nil {
			return model.Statistics{}, err
		}
		stats.ByPriority[priority] = count
	}

	assigneeRows, err := s.pool.Query(ctx, `
		SELECT u.name, count(t.id)
		FROM tickets t
		JOIN users u ON u.id=t.assignee_id
		GROUP BY u.name
		ORDER BY u.name`)
	if err != nil {
		return model.Statistics{}, err
	}
	defer assigneeRows.Close()
	for assigneeRows.Next() {
		var name string
		var count int64
		if err := assigneeRows.Scan(&name, &count); err != nil {
			return model.Statistics{}, err
		}
		stats.ByAssignee[name] = count
	}

	return stats, statusRows.Err()
}

const ticketSelect = `
	SELECT id, ticket_no, customer_id, type, priority, status, title, description,
		result, note, assignee_id, created_at, updated_at, resolved_at, closed_at, sla_due_at
	FROM tickets`

func scanTicket(row rowScanner) (model.Ticket, error) {
	var ticket model.Ticket
	var resolvedAt pgtype.Timestamptz
	var closedAt pgtype.Timestamptz
	var assigneeID *int64
	if err := row.Scan(
		&ticket.ID,
		&ticket.TicketNo,
		&ticket.CustomerID,
		&ticket.Type,
		&ticket.Priority,
		&ticket.Status,
		&ticket.Title,
		&ticket.Description,
		&ticket.Result,
		&ticket.Note,
		&assigneeID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
		&resolvedAt,
		&closedAt,
		&ticket.SLADueAt,
	); err != nil {
		return model.Ticket{}, err
	}
	ticket.AssigneeID = assigneeID
	if resolvedAt.Valid {
		value := resolvedAt.Time
		ticket.ResolvedAt = &value
	}
	if closedAt.Valid {
		value := closedAt.Time
		ticket.ClosedAt = &value
	}
	return ticket, nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
