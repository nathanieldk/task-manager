CREATE TABLE task_logs (
    id UUID UNIQUE PRIMARY KEY,
    task_id UUID NOT NULL,
    changed_by UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_value JSONB,
    new_value JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_logs_task_id ON task_logs(task_id);
CREATE INDEX idx_task_logs_changed_by ON task_logs(changed_by);
CREATE INDEX idx_task_logs_created_at ON task_logs(created_at DESC);
