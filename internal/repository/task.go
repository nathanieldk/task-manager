package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/nathanieldk/task-manager/internal/entity"
)

type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new PostgreSQL-backed task repository.
func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *entity.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, status, creator_id, assignee_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.Status,
		task.CreatorID, task.AssigneeID, task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*entity.Task, error) {
	query := `
		SELECT id, title, description, status, creator_id, assignee_id, created_at, updated_at
		FROM tasks WHERE id = $1`

	task := &entity.Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatorID, &task.AssigneeID, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *taskRepository) FindAll(ctx context.Context, creatorID string, status string, title string, limit, offset int) ([]*entity.Task, int64, error) {
	// Build dynamic query with filters
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("creator_id = $%d", argIdx))
	args = append(args, creatorID)
	argIdx++

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if title != "" {
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argIdx))
		args = append(args, "%"+title+"%")
		argIdx++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total matching rows
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	dataQuery := fmt.Sprintf(`
		SELECT id, title, description, status, creator_id, assignee_id, created_at, updated_at
		FROM tasks %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*entity.Task
	for rows.Next() {
		task := &entity.Task{}
		if err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status,
			&task.CreatorID, &task.AssigneeID, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *taskRepository) Update(ctx context.Context, task *entity.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, updated_at = $4
		WHERE id = $5`

	result, err := r.db.ExecContext(ctx, query,
		task.Title, task.Description, task.Status, task.UpdatedAt, task.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *taskRepository) UpdateAssignee(ctx context.Context, taskID string, assigneeID string) error {
	query := `
		UPDATE tasks
		SET assignee_id = $1, updated_at = NOW()
		WHERE id = $2`

	executor := GetExecutor(ctx, r.db)
	result, err := executor.ExecContext(ctx, query, assigneeID, taskID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
