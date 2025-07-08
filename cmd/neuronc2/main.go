package main

import (
	"log"
	"net/http"
	"neuronc2/config"
	"neuronc2/internal/database"
	"neuronc2/internal/server"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create C2 server
	c2Server := server.New(db, cfg)

	// Start HTTP server
	go func() {
		http.HandleFunc("/agent", c2Server.HandleAgent)
		http.HandleFunc("/activate", c2Server.HandleActivation)
		log.Printf("[C2] %s v%s started on %s", cfg.ServerName, cfg.ServerVersion, cfg.Port)
		log.Fatal(http.ListenAndServe(cfg.Port, nil))
	}()

	// Start MCP server
	mcpServer := mcpserver.NewMCPServer(cfg.ServerName, cfg.ServerVersion)
	c2Server.RegisterMCPTools(mcpServer)

	log.Println("[MCP] Ready to receive commands")
	if err := mcpserver.ServeStdio(mcpServer); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}
