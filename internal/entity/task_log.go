package entity

import (
	"encoding/json"
	"time"
)

type TaskLog struct {
	ID        string          `json:"id"`
	TaskID    string          `json:"task_id"`
	ChangedBy string          `json:"changed_by"`
	Action    string          `json:"action"`
	OldValue  json.RawMessage `json:"old_value,omitempty"`
	NewValue  json.RawMessage `json:"new_value"`
	CreatedAt time.Time       `json:"created_at"`
}
