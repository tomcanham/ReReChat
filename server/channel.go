// Server maintains the set of active clients and broadcasts messages to the
// clients.

package main

import (
	"encoding/json"
	"log"
)

// ChannelLeaveMessage describes a user leaving a channel
type ChannelLeaveMessage struct {
	Type     string `json:"type"`
	Channel  string `json:"channel"`
	Username string `json:"username"`
}

// ChannelJoinMessage describes a user joining a channel
type ChannelJoinMessage struct {
	Type     string `json:"type"`
	Channel  string `json:"channel"`
	Username string `json:"username"`
}

// Channel manages all broadcasting for a chat channel
type Channel struct {
	// Channel name.
	name string

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	_broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newChannel(name string) *Channel {
	return &Channel{
		name:       name,
		_broadcast: make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (ch *Channel) broadcast(message []byte) {
	for client := range ch.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(ch.clients, client)
		}
	}
}

func (ch *Channel) run() {
	for {
		select {
		case client := <-ch.register:
			ch.clients[client] = true
			log.Printf("[CHANNEL: %q] user %q joined channel\n", ch.name, client.username)
			msg := &ChannelJoinMessage{Type: "channel.joined", Channel: ch.name, Username: client.username}
			marshalled, _ := json.Marshal(msg)
			ch.broadcast(marshalled)

		case client := <-ch.unregister:
			if _, ok := ch.clients[client]; ok {
				log.Printf("[CHANNEL: %q] user %q left channel\n", ch.name, client.username)
				msg := &ChannelLeaveMessage{Type: "channel.left", Channel: ch.name, Username: client.username}
				marshalled, _ := json.Marshal(msg)
				ch.broadcast(marshalled)
				delete(ch.clients, client)
			} else {
				log.Printf("ERROR leaving channel %q\n", ch.name)
			}
		case bytes := <-ch._broadcast:
			ch.broadcast(bytes)
		}
	}
}

func (ch *Channel) join(c *Client) error {
	ch.register <- c

	return nil
}

func (ch *Channel) leave(c *Client) error {
	ch.unregister <- c

	return nil
}
