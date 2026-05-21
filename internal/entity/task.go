package entity

import "time"

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID        string     `json:"id" db:"id"`
	Type      string     `json:"type" db:"type"`
	Status    TaskStatus `json:"status" db:"status"`
	MachineID string     `json:"machine_id" db:"machine_id"`
	Payload   string     `json:"payload" db:"payload"`
	Result    string     `json:"result" db:"result"`
	Error     string     `json:"error" db:"error"`
	Logs      string     `json:"logs" db:"logs"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}
