package auth

import (
	"fmt"
	"neuronc2/internal/database"
	"neuronc2/internal/utils"
	"time"
)

type TokenManager struct {
	queries *database.Queries
}

func NewTokenManager(queries *database.Queries) *TokenManager {
	return &TokenManager{
		queries: queries,
	}
}

func (tm *TokenManager) GenerateDeploymentToken(notes string, maxUses int, duration time.Duration) (string, error) {
	token := fmt.Sprintf("DEPLOY-%s", utils.GenerateRandomString(12))
	validUntil := time.Now().Add(duration)

	err := tm.queries.CreateDeploymentToken(token, validUntil, maxUses, notes)
	return token, err
}

func (tm *TokenManager) ValidateToken(token string) (*database.DeploymentToken, error) {
	dt, err := tm.queries.GetDeploymentToken(token)
	if err != nil {
		return nil, err
	}

	if time.Now().After(dt.ValidUntil) {
		return nil, fmt.Errorf("token expired")
	}

	if dt.UsedCount >= dt.MaxUses {
		return nil, fmt.Errorf("token usage limit reached")
	}

	return dt, nil
}
