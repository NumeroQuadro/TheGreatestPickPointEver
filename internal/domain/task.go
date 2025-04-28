package domain

import "time"

type TaskStatus string

var (
	Created        TaskStatus = "CREATED"
	Processing     TaskStatus = "PROCESSING"
	Failed         TaskStatus = "FAILED"
	NoAttemptsLeft TaskStatus = "NO_ATTEMPTS_LEFT"
)

type TaskType string

var (
	AuditLog       TaskType = "AUDIT_LOG"
	OrderStatusLog TaskType = "ORDER_STATUS_LOG"
)

type Task struct {
	TaskID        int64      `json:"task_id" db:"task_id"`
	TaskStatus    TaskStatus `json:"task_status" db:"task_status"`
	TaskType      TaskType   `json:"task_type" db:"task_type"`
	EntryID       int64      `json:"entry_id" db:"entry_id"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	AttemptsCount int        `json:"attempts_count" db:"attempts_count"`
	NextAttemptAt time.Time  `json:"next_attempt_at" db:"next_attempt_at"`
	FinishedAt    time.Time  `json:"finished_at" db:"finished_at"`
}
