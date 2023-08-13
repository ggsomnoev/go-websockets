package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager *Manager
	chatroom string

	egress chan Event
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager: manager,
		egress: make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer func () {
		c.manager.removeClient(c)
	}()

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		fmt.Printf("SetReadDeadline error: %v\n", err)
		return
	}

	c.connection.SetReadLimit(512)

	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("Websocket connection closed during read abnormaly: %v\n", c.connection.RemoteAddr())
			}
			fmt.Println(payload)
			fmt.Printf("Websocket connection closed: %v\n", err)
			break;
		}

		var req Event
		
		if err :=json.Unmarshal(payload, &req); err != nil {
			fmt.Printf("Error unmarshaling event: %v\n", err)
		}
		
		if err := c.manager.routeEvent(req, c); err != nil {
			fmt.Printf("Error routing the event: %v\n", err)
		}
	}
}

func (c *Client) writeMessages() {
	defer func () {
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					fmt.Println("Couldn't send message back to the client: ", err)
				}
				return
			}

			data, err := json.Marshal(message)

			if err != nil {
				fmt.Printf("Error marshaling the event: %v\n", err)
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Println("Couldn't send message back to the client: ", err)
				return	
			}

			fmt.Printf("Message Sent:%v\n", string(data))
		case <-ticker.C:
			//fmt.Println("Ping")

			if err := c.connection.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				fmt.Println("Couldn't send ping message to the client: ", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string)  error {
	//fmt.Println("Pong")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}