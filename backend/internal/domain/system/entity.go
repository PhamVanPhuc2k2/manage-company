package system

import "time"

type HealthSnapshot struct {
	DatabaseTime time.Time
	RedisOk      bool
	JobQueued    bool
	RequestID    string
}

type Job struct {
	Name      string         `json:"name"`
	RequestID string         `json:"request_id"`
	Payload   map[string]any `json:"payload"`
}

const JobSystemPing = "system.ping"
