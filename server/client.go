package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

type SelfDescribing interface {
	TypeName() string
}

// Client is a middleman between the websocket connection and the server.
type Client struct {
	username string
	server   *Server
	channels map[string]*Channel
	mu       sync.Mutex
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound WS messages.
	send chan []byte

	// Channel of incoming client commands
	messages chan []byte
}

func createClient(conn *websocket.Conn, server *Server, username string) *Client {
	client := &Client{
		channels: make(map[string]*Channel),
		conn:     conn,
		messages: make(chan []byte),
		send:     make(chan []byte, 256),
		server:   server,
		username: username,
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
	go client.commandPump()

	return client
}

// readPump pumps messages from the websocket connection to the server.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.leaveAllChannels()
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}

			c.leaveAllChannels()
			break
		}
		message = bytes.TrimSpace(message)
		c.messages <- message
	}
}

// writePump pumps messages from the server to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The server closed the channel.
				fmt.Printf("[USER: %q] websocket.Conn.SetWriteDeadline error\n", c.username)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Printf("[USER: %q] websocket.Conn.NextWriter error: %v\n", c.username, err)
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				fmt.Printf("[USER: %q] io.WriteCloser.Close error: %v\n", c.username, err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func marshalSelfDescribing(sd SelfDescribing) ([]byte, error) {
	typeName := []byte(fmt.Sprintf("%s\n", sd.TypeName()))
	marshalled, err := json.Marshal(sd)
	if err != nil {
		return nil, err
	}

	return append(typeName, marshalled...), nil
}

func (c *Client) sendSelfDescribing(sd SelfDescribing) error {
	if payload, err := marshalSelfDescribing(sd); err == nil {
		c.send <- payload
		return nil
	} else {
		log.Printf("sendSelfDescribing error: %+v\n", err)
		return err
	}
}

func (c *Client) broadcastToChannel(channel string, sd SelfDescribing) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch, ok := c.channels[channel]
	if !ok {
		return fmt.Errorf("not in channel %q", channel)
	}

	payload, err := marshalSelfDescribing(sd)
	if err != nil {
		return err
	}

	ch._broadcast <- payload
	return nil
}

func (c *Client) sendChannelInfo(ch *Channel) error {
	chanInfo := &ChannelInfoMessage{
		Channel:  ch.name,
		Username: c.username,
		Users:    make([]string, 0, len(ch.clients)),
	}
	for client := range ch.clients {
		chanInfo.Users = append(chanInfo.Users, client.username)
	}

	return c.sendSelfDescribing(chanInfo)
}

func (c *Client) sendUserLeave(ch *Channel) error {
	userLeave := &UserLeaveMessage{
		Username: c.username,
		Channel:  ch.name,
	}

	return c.sendSelfDescribing(userLeave)
}

func (c *Client) sendUserJoin(ch *Channel) error {
	userJoin := &UserJoinMessage{
		Username: c.username,
		Channel:  ch.name,
	}

	return c.sendSelfDescribing(userJoin)
}

func (c *Client) joinChannel(name string) (*Channel, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.channels[name]; ok {
		return nil, fmt.Errorf("already in channel %q", name)
	}

	ch := c.server.getChannel(name)
	if err := ch.join(c); err != nil {
		return nil, err
	}

	c.channels[name] = ch
	go func() {
		c.sendUserJoin(ch)
		c.sendChannelInfo(ch)
	}()

	return ch, nil
}

func (c *Client) leaveChannel(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Printf("[USER: %q] leaving channel %q", c.username, name)
	ch, ok := c.channels[name]
	if !ok {
		return fmt.Errorf("not in channel %v", name)
	}

	if err := ch.leave(c); err != nil {
		return err
	}

	delete(c.channels, name)
	go c.sendUserLeave(ch)

	return nil
}

func (c *Client) sendChannelsList() error {
	clm := c.server.getChannelsListMessage()

	return c.sendSelfDescribing(&clm)
}

func (c *Client) leaveAllChannels() {
	for name := range c.channels {
		if err := c.leaveChannel(name); err != nil {
			log.Printf("WARNING: client unable to leave channel %q: %v\n", name, err)
		}
	}
}
