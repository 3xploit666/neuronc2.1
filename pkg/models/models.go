package models

import "time"

// JSON response structures
type MCPResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
	Metadata struct {
		Timestamp  time.Time `json:"timestamp"`
		Command    string    `json:"command"`
		Duration   int64     `json:"duration_ms,omitempty"`
		TotalCount int       `json:"total_count,omitempty"`
	} `json:"metadata"`
}

// Agent structures for API responses
type AgentJSON struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	OS       string    `json:"os"`
	Arch     string    `json:"arch"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"last_seen"`
}

// Agent activation request
type ActivationRequest struct {
	Token    string            `json:"token"`
	Metadata map[string]string `json:"metadata"`
}

// Agent activation response
type ActivationResponse struct {
	AgentID string `json:"agent_id"`
	APIKey  string `json:"api_key"`
	Status  string `json:"status"`
}

// Agent response format
type AgentResponse struct {
	AgentID string `json:"agent_id"`
	Output  string `json:"output"`
}
