package main

import (
	"encoding/json"
	"time"
)

const (
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"
	EventChangeRoom  = "change_room"
)

type Event struct {
	Type string `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// func NewEvent(event Event) *Event {
// 	return &Event{
// 		event.Type,
// 		event.Payload,
// 	}
// }

type EventHandler func(event Event, c *Client) error

type SendMessageEvent struct {
	Message string `json:"message"`
	From	string `json:"from"`	
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`	
}

type ChangeRoomEvent struct {
	Name string `json:"name"`	
}