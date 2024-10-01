package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, c *Client) error

const (
	// types for SendMessage
	SendMessageEvent = "send_message"
	// types for NewMessage
	ReceiveMessageEvent = "receive_message"
	// types for ChangeRoom
	ChangeRoomEvent = "change_room"
)

// payload for SendMessage
type SendMessagePayload struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

// payload for ReceiveMessage
type ReceiveMessagePayload struct {
	SendMessagePayload
	Sent time.Time `json:"sent"`
}

// payload for ChangeRoom
type ChangeRoomPayload struct {
	Name string `json:"name"`
}

// Handler for SendMessage
func SendMessageHandler(event Event, c *Client) error {

	var sendMessagePayload SendMessagePayload
	if err := json.Unmarshal(event.Payload, &sendMessagePayload); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	var broadcastPayload ReceiveMessagePayload
	broadcastPayload.Message = sendMessagePayload.Message
	broadcastPayload.From = sendMessagePayload.From
	broadcastPayload.Sent = time.Now()

	data, err := json.Marshal(broadcastPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	broadcastEvent := Event{
		Type:    SendMessageEvent,
		Payload: data,
	}

	for client := range c.manager.clients {
		if client.chatroom == c.chatroom {
			client.egress <- broadcastEvent
		}
	}

	return nil
}

// Handler for ChangeRoom
func ChatRoomHandler(event Event, c *Client) error {
	var changeRoomPayload ChangeRoomPayload
	if err := json.Unmarshal(event.Payload, &changeRoomPayload); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// Add Client to chat room
	c.chatroom = changeRoomPayload.Name
	fmt.Println(changeRoomPayload)

	return nil
}
