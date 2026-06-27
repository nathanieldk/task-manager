package repository

import (
	"context"

	"github.com/nathanieldk/task-manager/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	FindByID(ctx context.Context, id string) (*entity.Task, error)
	FindAll(ctx context.Context, creatorID string, status string, title string, limit, offset int) ([]*entity.Task, int64, error)
	Update(ctx context.Context, task *entity.Task) error
	Delete(ctx context.Context, id string) error
	UpdateAssignee(ctx context.Context, taskID string, assigneeID string) error
}

type TaskLogRepository interface {
	Create(ctx context.Context, log *entity.TaskLog) error
}

type TransactionManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
