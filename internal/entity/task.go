package entity

import (
	"slices"
	"time"
)

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

var ValidTaskStatuses = []TaskStatus{
	TaskStatusTodo,
	TaskStatusInProgress,
	TaskStatusDone,
}

func (s TaskStatus) IsValid() bool {
	return slices.Contains(ValidTaskStatuses, s)
}

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	CreatorID   string     `json:"creator_id"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
