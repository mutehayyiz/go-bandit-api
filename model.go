package main

import "time"

type Status string

const (
	StatusDone    Status = "done"
	StatusProcess Status = "process"
	StatusError   Status = "error"
)

type Scan struct {
	ID        string                 `json:"id"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	URL       string                 `json:"url"`
	Status    Status                 `json:"status"`
	Result    map[string]interface{} `json:"result,omitempty"`
	IsSecure  bool                   `json:"is_secure,omitempty"`
	Error     string                 `json:"error,omitempty"`
}
