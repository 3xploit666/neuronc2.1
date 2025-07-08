package server

import (
	"database/sql"
	"neuronc2/config"
	"neuronc2/internal/agent"
	"neuronc2/internal/auth"
	"neuronc2/internal/database"
	"sync"
)

type C2Server struct {
	agents       map[string]*agent.Agent
	mutex        sync.RWMutex
	db           *sql.DB
	config       *config.Config
	queries      *database.Queries
	tokenManager *auth.TokenManager
	agentHandler *agent.Handler
}

func New(db *sql.DB, config *config.Config) *C2Server {
	queries := database.NewQueries(db)
	tokenManager := auth.NewTokenManager(queries)
	agentHandler := agent.NewHandler(queries)

	return &C2Server{
		agents:       make(map[string]*agent.Agent),
		db:           db,
		config:       config,
		queries:      queries,
		tokenManager: tokenManager,
		agentHandler: agentHandler,
	}
}

// Implement interface methods
func (c2 *C2Server) GetQueries() *database.Queries {
	return c2.queries
}

func (c2 *C2Server) GetTokenManager() *auth.TokenManager {
	return c2.tokenManager
}

func (c2 *C2Server) GetConfig() *config.Config {
	return c2.config
}

func (c2 *C2Server) GetAgentHandler() *agent.Handler {
	return c2.agentHandler
}

func (c2 *C2Server) AddAgent(agent *agent.Agent) {
	c2.mutex.Lock()
	defer c2.mutex.Unlock()
	c2.agents[agent.ID] = agent
}

func (c2 *C2Server) RemoveAgent(agentID string) {
	c2.mutex.Lock()
	defer c2.mutex.Unlock()
	delete(c2.agents, agentID)
}

func (c2 *C2Server) GetAgent(agentID string) (*agent.Agent, bool) {
	c2.mutex.RLock()
	defer c2.mutex.RUnlock()
	agent, exists := c2.agents[agentID]
	return agent, exists
}

func (c2 *C2Server) GetAllAgents() map[string]*agent.Agent {
	c2.mutex.RLock()
	defer c2.mutex.RUnlock()

	// Return a copy of the map
	agents := make(map[string]*agent.Agent)
	for k, v := range c2.agents {
		agents[k] = v
	}
	return agents
}
