package controllers

import (
	"encoding/json"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, c *Client) error

type SendMessageEventStruct struct {
	From    string `json:"from"`
	To      string `json:"to"`
	message string `json:"message"`
}

type NewMessageEvent struct {
	SendMessageEventStruct
	Sent time.Time
}

const (
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"
	EventChangeRoom  = "change_room"
)
