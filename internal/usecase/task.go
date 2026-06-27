package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nathanieldk/task-manager/internal/dto"
	"github.com/nathanieldk/task-manager/internal/entity"
	"github.com/nathanieldk/task-manager/internal/pkg/apperror"
	"github.com/nathanieldk/task-manager/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TaskUsecase interface {
	Create(ctx context.Context, userID string, req dto.CreateTaskRequest, idempotencyKey string) (*entity.Task, bool, error)
	GetByID(ctx context.Context, userID string, taskID string) (*entity.Task, error)
	List(ctx context.Context, userID string, query dto.TaskListQuery) ([]*entity.Task, int64, error)
	Update(ctx context.Context, userID string, taskID string, req dto.UpdateTaskRequest) (*entity.Task, error)
	Delete(ctx context.Context, userID string, taskID string) error
	Assign(ctx context.Context, userID string, taskID string, req dto.AssignTaskRequest) error
}

type taskUsecase struct {
	taskRepo    repository.TaskRepository
	taskLogRepo repository.TaskLogRepository
	userRepo    repository.UserRepository
	txManager   repository.TransactionManager
	idempotency IdempotencyUsecase
	logger      *zap.Logger
}

func NewTaskUsecase(
	taskRepo repository.TaskRepository,
	taskLogRepo repository.TaskLogRepository,
	userRepo repository.UserRepository,
	txManager repository.TransactionManager,
	idempotency IdempotencyUsecase,
	logger *zap.Logger,
) TaskUsecase {
	return &taskUsecase{
		taskRepo:    taskRepo,
		taskLogRepo: taskLogRepo,
		userRepo:    userRepo,
		txManager:   txManager,
		idempotency: idempotency,
		logger:      logger,
	}
}

func (uc *taskUsecase) Create(ctx context.Context, userID string, req dto.CreateTaskRequest, idempotencyKey string) (*entity.Task, bool, error) {
	// Check idempotency — if we have a cached response, return it
	if idempotencyKey != "" {
		cached, err := uc.idempotency.Get(ctx, idempotencyKey)
		if err != nil {
			uc.logger.Warn("Failed to check idempotency cache", zap.Error(err))
		}
		if cached != nil {
			return cached, true, nil
		}
	}

	now := time.Now().UTC()
	task := &entity.Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      entity.TaskStatusTodo,
		CreatorID:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Attempt to acquire the idempotency lock before creating
	if idempotencyKey != "" {
		acquired, err := uc.idempotency.Acquire(ctx, idempotencyKey)
		if err != nil {
			uc.logger.Warn("Failed to acquire idempotency lock", zap.Error(err))
		}
		if !acquired {
			for range 10 {
				time.Sleep(50 * time.Millisecond)
				cached, err := uc.idempotency.Get(ctx, idempotencyKey)
				if err == nil && cached != nil {
					return cached, true, nil
				}
			}
			return nil, false, apperror.ErrConflict("Duplicate request is being processed")
		}
	}

	if err := uc.taskRepo.Create(ctx, task); err != nil {
		return nil, false, apperror.ErrInternal(err)
	}

	// Store in idempotency cache
	if idempotencyKey != "" {
		if err := uc.idempotency.Store(ctx, idempotencyKey, task); err != nil {
			uc.logger.Warn("Failed to store idempotency result", zap.Error(err))
		}
	}

	return task, false, nil
}

func (uc *taskUsecase) GetByID(ctx context.Context, userID string, taskID string) (*entity.Task, error) {
	if _, err := uuid.Parse(taskID); err != nil {
		return nil, apperror.ErrNotFound("Task")
	}

	task, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, apperror.ErrInternal(err)
	}
	if task == nil {
		return nil, apperror.ErrNotFound("Task")
	}

	// Verify the user is the creator or assignee
	if task.CreatorID != userID && (task.AssigneeID == nil || *task.AssigneeID != userID) {
		return nil, apperror.ErrForbidden("You do not have access to this task")
	}

	return task, nil
}

func (uc *taskUsecase) List(ctx context.Context, userID string, query dto.TaskListQuery) ([]*entity.Task, int64, error) {
	query.Normalize()

	// Validate status filter if provided
	if query.Status != "" {
		if !entity.TaskStatus(query.Status).IsValid() {
			return nil, 0, apperror.ErrBadRequest("INVALID_STATUS", fmt.Sprintf("Invalid status: %s. Valid values: todo, in_progress, done", query.Status))
		}
	}

	tasks, total, err := uc.taskRepo.FindAll(ctx, userID, query.Status, query.Title, query.Limit, query.Offset())
	if err != nil {
		return nil, 0, apperror.ErrInternal(err)
	}

	return tasks, total, nil
}

func (uc *taskUsecase) Update(ctx context.Context, userID string, taskID string, req dto.UpdateTaskRequest) (*entity.Task, error) {
	if _, err := uuid.Parse(taskID); err != nil {
		return nil, apperror.ErrNotFound("Task")
	}

	task, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, apperror.ErrInternal(err)
	}
	if task == nil {
		return nil, apperror.ErrNotFound("Task")
	}

	// Only the creator can update a task
	if task.CreatorID != userID {
		return nil, apperror.ErrForbidden("Only the task creator can update this task")
	}

	// Apply partial updates
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, apperror.ErrBadRequest("INVALID_STATUS", fmt.Sprintf("Invalid status: %s. Valid values: todo, in_progress, done", *req.Status))
		}
		task.Status = *req.Status
	}
	task.UpdatedAt = time.Now().UTC()

	if err := uc.taskRepo.Update(ctx, task); err != nil {
		return nil, apperror.ErrInternal(err)
	}

	return task, nil
}

func (uc *taskUsecase) Delete(ctx context.Context, userID string, taskID string) error {
	if _, err := uuid.Parse(taskID); err != nil {
		return apperror.ErrNotFound("Task")
	}

	task, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return apperror.ErrInternal(err)
	}
	if task == nil {
		return apperror.ErrNotFound("Task")
	}

	// Only the creator can delete a task
	if task.CreatorID != userID {
		return apperror.ErrForbidden("Only the task creator can delete this task")
	}

	if err := uc.taskRepo.Delete(ctx, taskID); err != nil {
		return apperror.ErrInternal(err)
	}

	return nil
}

func (uc *taskUsecase) Assign(ctx context.Context, userID string, taskID string, req dto.AssignTaskRequest) error {
	if _, err := uuid.Parse(taskID); err != nil {
		return apperror.ErrNotFound("Task")
	}

	// Verify the task exists and belongs to the requesting user
	task, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return apperror.ErrInternal(err)
	}
	if task == nil {
		return apperror.ErrNotFound("Task")
	}
	if task.CreatorID != userID {
		return apperror.ErrForbidden("Only the task creator can assign this task")
	}

	// Fetch the creator to get their team_id
	creator, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.ErrInternal(err)
	}

	// Verify the assignee exists
	assignee, err := uc.userRepo.FindByID(ctx, req.AssigneeID)
	if err != nil {
		return apperror.ErrInternal(err)
	}
	if assignee == nil {
		return apperror.ErrNotFound("Assignee user")
	}

	// Verify assignee is in the same team by comparing team_id in the users table
	if assignee.TeamID != creator.TeamID {
		return apperror.ErrBadRequest("DIFFERENT_TEAM", "Assignee must be in the same team")
	}

	// Run update assignee + create log + mock notification in transaction block
	err = uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Update assignee
		if err = uc.taskRepo.UpdateAssignee(txCtx, taskID, req.AssigneeID); err != nil {
			return err
		}

		// Create task log
		oldValue, _ := json.Marshal(map[string]any{
			"assignee_id": task.AssigneeID,
		})
		newValue, _ := json.Marshal(map[string]any{
			"assignee_id": req.AssigneeID,
		})

		taskLog := &entity.TaskLog{
			ID:        uuid.New().String(),
			TaskID:    taskID,
			ChangedBy: userID,
			Action:    "assigned",
			OldValue:  oldValue,
			NewValue:  newValue,
			CreatedAt: time.Now().UTC(),
		}

		if err = uc.taskLogRepo.Create(txCtx, taskLog); err != nil {
			return err
		}

		// Mock notification
		uc.logger.Info("Task assignment notification",
			zap.String("task_id", taskID),
			zap.String("assignee_id", req.AssigneeID),
			zap.String("assigned_by", userID),
			zap.String("notification", fmt.Sprintf("Task '%s' has been assigned to user %s", task.Title, assignee.Username)),
		)

		return nil
	})
	if err != nil {
		return apperror.ErrInternal(err)
	}

	return nil
}
