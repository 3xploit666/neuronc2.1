package mcptools

import (
	"neuronc2/config"
	"neuronc2/internal/agent"
	"neuronc2/internal/auth"
	"neuronc2/internal/database"
)

// ServerInterface defines all methods that C2Server must implement
type ServerInterface interface {
	GetAllAgents() map[string]*agent.Agent
	GetAgent(string) (*agent.Agent, bool)
	GetQueries() *database.Queries
	GetConfig() *config.Config
	GetTokenManager() *auth.TokenManager
	GetAgentHandler() *agent.Handler
}
