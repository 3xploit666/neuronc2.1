package database

import (
	"database/sql"
	"time"
)

type Queries struct {
	db *sql.DB
}

func NewQueries(db *sql.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) GetDeploymentToken(token string) (*DeploymentToken, error) {
	var dt DeploymentToken
	err := q.db.QueryRow(`
		SELECT id, token, valid_until, max_uses, used_count 
		FROM deployment_tokens 
		WHERE token = ?`, token).Scan(
		&dt.ID, &dt.Token, &dt.ValidUntil, &dt.MaxUses, &dt.UsedCount)
	
	if err != nil {
		return nil, err
	}
	
	return &dt, nil
}

func (q *Queries) IncrementTokenUsage(tokenID int64) error {
	_, err := q.db.Exec(`
		UPDATE deployment_tokens 
		SET used_count = used_count + 1 
		WHERE id = ?`, tokenID)
	return err
}

func (q *Queries) CreateAgent(agent *AgentRecord) error {
	_, err := q.db.Exec(`
		INSERT INTO agents (agent_id, api_key, hostname, username, os, arch, activated_at, last_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		agent.AgentID, agent.APIKey, agent.Hostname, agent.Username,
		agent.OS, agent.Arch, agent.ActivatedAt, agent.LastSeen)
	return err
}

func (q *Queries) RecordTokenUsage(tokenID int64, agentID string) error {
	_, err := q.db.Exec(`
		INSERT INTO token_agent_usage (token_id, agent_id, used_at)
		VALUES (?, ?, ?)`, tokenID, agentID, time.Now())
	return err
}

func (q *Queries) GetAgentByAPIKey(apiKey string) (*AgentRecord, error) {
	var agent AgentRecord
	err := q.db.QueryRow(`
		SELECT agent_id, hostname, username, os, arch 
		FROM agents 
		WHERE api_key = ?`, apiKey).Scan(
		&agent.AgentID, &agent.Hostname, &agent.Username,
		&agent.OS, &agent.Arch)
	
	if err != nil {
		return nil, err
	}
	
	return &agent, nil
}

func (q *Queries) UpdateAgentLastSeen(agentID string) error {
	_, err := q.db.Exec(`UPDATE agents SET last_seen = ? WHERE agent_id = ?`, time.Now(), agentID)
	return err
}

func (q *Queries) SaveCommandHistory(agentID, command, response string) error {
	_, err := q.db.Exec(`
		INSERT INTO command_history (agent_id, command, response, executed_at)
		VALUES (?, ?, ?, ?)`,
		agentID, command, response, time.Now())
	return err
}

func (q *Queries) GetAllAgents() ([]AgentRecord, error) {
	rows, err := q.db.Query(`
		SELECT agent_id, hostname, username, os, arch, activated_at, last_seen
		FROM agents ORDER BY last_seen DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []AgentRecord
	for rows.Next() {
		var agent AgentRecord
		if err := rows.Scan(&agent.AgentID, &agent.Hostname, &agent.Username,
			&agent.OS, &agent.Arch, &agent.ActivatedAt, &agent.LastSeen); err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

func (q *Queries) GetAgentInfo(agentID string) (*AgentRecord, error) {
	var agent AgentRecord
	err := q.db.QueryRow(`
		SELECT id, agent_id, api_key, hostname, username, os, arch, activated_at, last_seen
		FROM agents WHERE agent_id = ?`, agentID).Scan(
		&agent.ID, &agent.AgentID, &agent.APIKey, &agent.Hostname, &agent.Username,
		&agent.OS, &agent.Arch, &agent.ActivatedAt, &agent.LastSeen)

	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (q *Queries) GetCommandHistory(agentID string, limit int) ([]CommandHistory, error) {
	rows, err := q.db.Query(`
		SELECT command, response, executed_at
		FROM command_history
		WHERE agent_id = ?
		ORDER BY executed_at DESC
		LIMIT ?`, agentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []CommandHistory
	for rows.Next() {
		var entry CommandHistory
		var cmd sql.NullString
		var resp sql.NullString
		var execTime time.Time

		if err := rows.Scan(&cmd, &resp, &execTime); err != nil {
			return nil, err
		}

		entry.AgentID = agentID
		entry.ExecutedAt = execTime

		if cmd.Valid {
			entry.Command = cmd.String
		}
		if resp.Valid {
			entry.Response = resp.String
		}

		history = append(history, entry)
	}

	return history, nil
}

func (q *Queries) GetDatabaseStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var tokenCount int64
	q.db.QueryRow("SELECT COUNT(*) FROM deployment_tokens").Scan(&tokenCount)
	stats["total_tokens"] = tokenCount

	var activeTokenCount int64
	q.db.QueryRow(`
		SELECT COUNT(*) FROM deployment_tokens 
		WHERE valid_until > ? AND used_count < max_uses`, time.Now()).Scan(&activeTokenCount)
	stats["active_tokens"] = activeTokenCount

	var agentCount int64
	q.db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&agentCount)
	stats["total_agents"] = agentCount

	var historyCount int64
	q.db.QueryRow("SELECT COUNT(*) FROM command_history").Scan(&historyCount)
	stats["total_commands"] = historyCount

	return stats, nil
}

func (q *Queries) CreateDeploymentToken(token string, validUntil time.Time, maxUses int, notes string) error {
	_, err := q.db.Exec(`
		INSERT INTO deployment_tokens (token, valid_until, max_uses, created_at, notes)
		VALUES (?, ?, ?, ?, ?)`,
		token, validUntil, maxUses, time.Now(), notes)
	return err
}

func (q *Queries) ListTokens() ([]DeploymentToken, error) {
	rows, err := q.db.Query(`
		SELECT id, token, valid_until, max_uses, used_count, created_at, notes
		FROM deployment_tokens
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []DeploymentToken
	for rows.Next() {
		var token DeploymentToken
		var notes sql.NullString

		if err := rows.Scan(&token.ID, &token.Token, &token.ValidUntil, &token.MaxUses,
			&token.UsedCount, &token.CreatedAt, &notes); err != nil {
			return nil, err
		}

		if notes.Valid {
			token.Notes = notes.String
		}

		// Determine status
		if time.Now().After(token.ValidUntil) {
			token.Status = "expired"
		} else if token.UsedCount >= token.MaxUses {
			token.Status = "exhausted"
		} else {
			token.Status = "active"
		}

		// Get agents that used this token
		agentRows, err := q.db.Query(`
			SELECT a.agent_id 
			FROM token_agent_usage tau
			JOIN agents a ON tau.agent_id = a.agent_id
			WHERE tau.token_id = ?`, token.ID)
		if err == nil {
			defer agentRows.Close()
			for agentRows.Next() {
				var agentID string
				if err := agentRows.Scan(&agentID); err == nil {
					token.UsedBy = append(token.UsedBy, agentID)
				}
			}
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (q *Queries) RevokeToken(token string) error {
	result, err := q.db.Exec(`
		UPDATE deployment_tokens 
		SET valid_until = ? 
		WHERE token = ?`, time.Now(), token)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
