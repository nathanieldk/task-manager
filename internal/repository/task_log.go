package repository

import (
	"context"
	"database/sql"

	"github.com/nathanieldk/task-manager/internal/entity"
)

type taskLogRepository struct {
	db *sql.DB
}

func NewTaskLogRepository(db *sql.DB) TaskLogRepository {
	return &taskLogRepository{db: db}
}

func (r *taskLogRepository) Create(ctx context.Context, log *entity.TaskLog) error {
	query := `
		INSERT INTO task_logs (id, task_id, changed_by, action, old_value, new_value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	executor := GetExecutor(ctx, r.db)
	_, err := executor.ExecContext(ctx, query,
		log.ID, log.TaskID, log.ChangedBy, log.Action,
		log.OldValue, log.NewValue, log.CreatedAt,
	)
	return err
}
