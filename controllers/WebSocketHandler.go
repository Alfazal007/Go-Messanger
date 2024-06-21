package controllers

import (
	helper "messager/helpers"
	"messager/internal/database"
	"net/http"

	"github.com/gorilla/websocket"
)

func (m *Manager) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	webSocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
	user, ok := r.Context().Value("user").(database.User)
	if !ok {
		helper.RespondWithError(w, 400, "Issue with finding the user from the database")
		return
	}
	conn, err := webSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		helper.RespondWithError(w, 400, "Issue upgrading the connection")
		return
	}
	client := NewClient(conn, m, user.Username)
	m.addClent(client)
	go client.ReadMessageFromClient()
	go client.WriteMessageToClient()
	// add heartbeat mechanisms
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	switch origin {
	case "http://localhost:8000":
		return true
	default:
		return false
	}
}
