package database

import "time"

type DeploymentToken struct {
	ID         int64     `json:"id"`
	Token      string    `json:"token"`
	ValidUntil time.Time `json:"valid_until"`
	MaxUses    int       `json:"max_uses"`
	UsedCount  int       `json:"used_count"`
	CreatedAt  time.Time `json:"created_at"`
	Notes      string    `json:"notes"`
	Status     string    `json:"status"`
	UsedBy     []string  `json:"used_by,omitempty"`
}

type AgentRecord struct {
	ID          int64     `json:"id"`
	AgentID     string    `json:"agent_id"`
	APIKey      string    `json:"api_key"`
	Hostname    string    `json:"hostname"`
	Username    string    `json:"username"`
	OS          string    `json:"os"`
	Arch        string    `json:"arch"`
	ActivatedAt time.Time `json:"activated_at"`
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"`
}

type CommandHistory struct {
	ID         int64     `json:"id"`
	AgentID    string    `json:"agent_id"`
	Command    string    `json:"command"`
	Response   string    `json:"response"`
	ExecutedAt time.Time `json:"executed_at"`
}
