package agent

import (
	"time"
	"github.com/gorilla/websocket"
)

type Agent struct {
	ID       string
	Conn     *websocket.Conn
	LastSeen time.Time
	Response chan string
	Hostname string
	Username string
	OS       string
	Arch     string
	APIKey   string
}

func New(id string, conn *websocket.Conn, apiKey string) *Agent {
	return &Agent{
		ID:       id,
		Conn:     conn,
		Response: make(chan string, 10),
		APIKey:   apiKey,
		LastSeen: time.Now(),
	}
}

func (a *Agent) UpdateInfo(hostname, username, os, arch string) {
	a.Hostname = hostname
	a.Username = username
	a.OS = os
	a.Arch = arch
	a.LastSeen = time.Now()
}

func (a *Agent) SendMessage(msg []byte) error {
	return a.Conn.WriteMessage(websocket.TextMessage, msg)
}

func (a *Agent) Close() {
	close(a.Response)
	a.Conn.Close()
}
