// Server maintains the set of active clients and broadcasts messages to the
// clients.

package main

import (
	"log"
)

// ChannelLeaveMessage describes a user leaving a channel
type ChannelLeaveMessage struct {
	Channel  string `json:"channel"`
	Username string `json:"username"`
}

func (msg *ChannelLeaveMessage) TypeName() string {
	return "channel.left"
}

// ChannelJoinMessage describes a user joining a channel
type ChannelJoinMessage struct {
	Channel  string `json:"channel"`
	Username string `json:"username"`
}

func (msg *ChannelJoinMessage) TypeName() string {
	return "channel.joined"
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

func (ch *Channel) broadcastSelfDescribing(sd SelfDescribing) error {
	if payload, err := marshalSelfDescribing(sd); err == nil {
		ch.broadcast(payload)
		return nil
	} else {
		return err
	}
}

func (ch *Channel) run() {
	for {
		select {
		case client := <-ch.register:
			ch.clients[client] = true
			log.Printf("[CHANNEL: %q] user %q joined channel\n", ch.name, client.username)
			msg := &ChannelJoinMessage{Channel: ch.name, Username: client.username}
			ch.broadcastSelfDescribing(msg)

		case client := <-ch.unregister:
			if _, ok := ch.clients[client]; ok {
				delete(ch.clients, client)
				log.Printf("[CHANNEL: %q] user %q left channel\n", ch.name, client.username)
				msg := &ChannelLeaveMessage{Channel: ch.name, Username: client.username}
				ch.broadcastSelfDescribing(msg)
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
