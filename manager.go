package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
	clients ClientList
	sync.RWMutex

	otps RetentionMap

	handlers map[string]EventHandler
}

type userLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type respose struct {
	OTP string `json:"otp"`
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		otps:     NewRetentionMap(ctx, 15*time.Second),
	}

	m.setupEventHandlers()

	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
	m.handlers[EventChangeRoom] = ChatRoomHandler
}

func ChatRoomHandler(event Event, c *Client) error {
	var changeRoomEvent ChangeRoomEvent
	
	if err := json.Unmarshal(event.Payload, &changeRoomEvent); err != nil {
		return fmt.Errorf("Error trying to unmarshal changeChatroomEvent: %v", err)
	}

	c.chatroom = changeRoomEvent.Name

	return nil
}

func SendMessage(event Event, c *Client) error {
	var chatEvent SendMessageEvent

	if err := json.Unmarshal(event.Payload, &chatEvent); err != nil {
		return fmt.Errorf("Error trying to unmarshal message payload: %v\n", err)
	}

	var broadcastMessage NewMessageEvent

	broadcastMessage = NewMessageEvent{		
		SendMessageEvent: SendMessageEvent {
			From : chatEvent.From,
			Message : chatEvent.Message,
		},
		Sent : time.Now(),
	}

	// broadcastMessage.Sent = time.Now()
	// broadcastMessage.From = chatEvent.From
	// broadcastMessage.Message = chatEvent.Message

	data, err := json.Marshal(broadcastMessage)

	if err != nil {
		return fmt.Errorf("Error trying to marshal message: %v\n", err)	
	}

	var outgoingEvent Event

	outgoingEvent = Event {
		Type : EventNewMessage,
		Payload: data,
	}

	for client := range c.manager.clients {
		if client.chatroom == c.chatroom {
			client.egress <- outgoingEvent
		}
	}	
	
	return nil
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	}

	return errors.New(fmt.Sprintf("No such event type supported: %v", event.Type))
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {

	otp := r.URL.Query().Get("otp")

	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Empty OTP")
		return
	}

	if !m.otps.VerifyOTP(otp) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Invalid OTP")
		return
	}

	conn, err := websocketUpgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("Error occured upgrading the connection to WSC")
		return
	}

	client := NewClient(conn, m)

	fmt.Printf("New WS connection: %v\n", client.connection.RemoteAddr())

	m.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}

func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req userLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("Error occured trying to parse login data: %v\n", err)
		http.Error(w, "Error occured trying to parse login data", http.StatusBadRequest)
		return
	}

	if req.Username == "testmest" && req.Password == "123" {
		otp := m.otps.NewOTP()
		resp := respose{
			OTP: otp.Key,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			fmt.Printf("Couldn't marshal OTP data: %v\n", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, found := m.clients[client]; found {
		client.connection.Close()
		delete(m.clients, client)
	}
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")	

	if strings.Contains(origin, "localhost:8080") {
		return true
	}	

	fmt.Println("Origin must be https://localhost:8080")
	return false
}
