package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	clients map[*Client]bool
	sync.RWMutex
	handlers map[string]EventHandler
}

func NewManager() *Manager {
	m := &Manager{
		clients: make(map[*Client]bool),
	}
	m.setUpEventHandlers()
	return m
}

func (m *Manager) setUpEventHandlers() {
	m.handlers[EventSendMessage] = RedirectMessageToReceiver
}

func (m *Manager) addClent(client *Client) {
	m.Lock()
	defer m.Unlock()
	m.clients[client] = true
}

func (m *Manager) removeClent(client *Client) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("no such event type")
	}
}

// TODO:: add database and resilience here
func RedirectMessageToReceiver(event Event, c *Client) error {
	var redirectMessageType SendMessageEventStruct
	if err := json.Unmarshal(event.Payload, &redirectMessageType); err != nil {
		return fmt.Errorf("bad payload in request %v", err)
	}
	var broadCastEvent NewMessageEvent
	broadCastEvent.From = redirectMessageType.From
	broadCastEvent.message = redirectMessageType.message
	broadCastEvent.To = redirectMessageType.To
	broadCastEvent.Sent = time.Now()
	data, err := json.Marshal(broadCastEvent)
	if err != nil {
		return fmt.Errorf("error marshalling the data %v", err)
	}
	outgoingEvent := Event{
		Payload: data,
		Type:    EventNewMessage,
	}
	for client := range c.manager.clients {
		if client.username == redirectMessageType.To {
			client.egress <- outgoingEvent
		}
	}
	return nil
}
