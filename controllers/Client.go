package controllers

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	username   string
	egress     chan Event
}

func NewClient(conn *websocket.Conn, m *Manager, username string) *Client {
	return &Client{
		connection: conn,
		manager:    m,
		egress:     make(chan Event),
		username:   username,
	}
}

func (c *Client) WriteMessageToClient() {
	func() { defer c.manager.removeClent(c) }()
	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("Closed connection", err)
				}
				return
			}
			data, err := json.Marshal(message)
			if err != nil {
				log.Println("There was an error marshalling the payload", err)
				return
			}
			err = c.connection.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Error sending the message ", err)
			}
		default:
			// TODO:: Need to add ping pong heartbeat mechanisms
		}
	}
}

// mssages coming in from a client
func (c *Client) ReadMessageFromClient() {
	func() { defer c.manager.removeClent(c) }()
	c.connection.SetReadLimit(512)
	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading messag: %v", err)
			}
			break
		}
		var data Event
		err = json.Unmarshal(payload, &data)
		if err != nil {
			log.Println("Failed to unmarshall the data", err)
			break
		}
		// data came in from the client now i need to forward it to the manager to send it to someone
		err = c.manager.routeEvent(data, c)
		if err != nil {
			log.Println("There was an error executing whatever needed to be executed", err)
		}
	}
}
