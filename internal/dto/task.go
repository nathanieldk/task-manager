package dto

import "github.com/nathanieldk/task-manager/internal/entity"

type CreateTaskRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=500"`
	Description string `json:"description" validate:"max=5000"`
}

type UpdateTaskRequest struct {
	Title       *string            `json:"title" validate:"omitempty,min=1,max=500"`
	Description *string            `json:"description" validate:"omitempty,max=5000"`
	Status      *entity.TaskStatus `json:"status" validate:"omitempty"`
}

type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id" validate:"required,uuid"`
}

type TaskResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Status      string  `json:"status"`
	CreatorID   string  `json:"creator_id"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type TaskListResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	Pagination PaginationMeta `json:"pagination"`
}

func TaskFromEntity(t *entity.Task) TaskResponse {
	return TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		CreatorID:   t.CreatorID,
		AssigneeID:  t.AssigneeID,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type TaskListQuery struct {
	Status string `query:"status"`
	Title  string `query:"title"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

func (q *TaskListQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 10
	}
}

func (q *TaskListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
