package mcptools

import (
	"context"
	"fmt"
	"neuronc2/internal/utils"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func GenerateDeploymentToken(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		notes := "default deployment"
		if n, ok := req.Params.Arguments["notes"].(string); ok {
			notes = n
		}

		maxUses := 1
		if m, ok := req.Params.Arguments["max_uses"].(float64); ok {
			maxUses = int(m)
		}

		duration := 24 * time.Hour
		if d, ok := req.Params.Arguments["duration"].(string); ok {
			if parsed, err := time.ParseDuration(d); err == nil {
				duration = parsed
			}
		}

		tokenManager := c2.GetTokenManager()
		token, err := tokenManager.GenerateDeploymentToken(notes, maxUses, duration)
		if err != nil {
			return utils.FormatJSONResponse("generate_deployment_token", nil, err, startTime)
		}

		data := map[string]interface{}{
			"token": map[string]interface{}{
				"token":      token,
				"expires":    time.Now().Add(duration),
				"max_uses":   maxUses,
				"notes":      notes,
				"created_at": time.Now(),
			},
			"instructions": fmt.Sprintf("Use this token when building the agent:\n./build_agent.ps1 -Token \"%s\"", token),
		}

		return utils.FormatJSONResponse("generate_deployment_token", data, nil, startTime)
	}
}

func ListTokens(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		queries := c2.GetQueries()
		tokens, err := queries.ListTokens()
		if err != nil {
			return utils.FormatJSONResponse("list_tokens", nil, err, startTime)
		}

		// Count tokens by status
		statusCount := map[string]int{
			"active":    0,
			"expired":   0,
			"exhausted": 0,
		}
		for _, token := range tokens {
			statusCount[token.Status]++
		}

		data := map[string]interface{}{
			"tokens":  tokens,
			"count":   len(tokens),
			"summary": statusCount,
		}

		return utils.FormatJSONResponse("list_tokens", data, nil, startTime)
	}
}

func RevokeToken(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		token, ok := req.Params.Arguments["token"].(string)
		if !ok {
			return utils.FormatJSONResponse("revoke_token", nil, fmt.Errorf("token parameter missing"), startTime)
		}

		queries := c2.GetQueries()
		err := queries.RevokeToken(token)
		if err != nil {
			return utils.FormatJSONResponse("revoke_token", nil, err, startTime)
		}

		data := map[string]interface{}{
			"token":   token,
			"status":  "revoked",
			"message": fmt.Sprintf("Token %s has been revoked", token),
		}

		return utils.FormatJSONResponse("revoke_token", data, nil, startTime)
	}
}
