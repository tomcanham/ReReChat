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

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// ConnectedMessage tells the client they've connected (with other info)
type ConnectedMessage struct {
	Type     string `json:"type"`
	Username string `json:"username"`
}

// ChatMessage is the deserialized JSON object representing normal chat messages
type ChatMessage struct {
	Type    string `json:"type"`
	Sender  string `json:"sender"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

// ChannelInfoMessage is the initial channel information sent to newly-joined users
type ChannelInfoMessage struct {
	Type     string   `json:"type"`
	Channel  string   `json:"channel"`
	Username string   `json:"username"`
	Users    []string `json:"users"`
}

// ChannelListMessage is an instance of the abbreviated channel information
// sent in response to a channels.list message
type ChannelListMessage struct {
	Channel string `json:"channel"`
}

// ChannelListResponse is a list of ChannelListMessage objects send in
// response to a channels.list message
type ChannelListResponse struct {
	Type     string                `json:"type"`
	Channels []*ChannelListMessage `json:"channels"`
}

// UserJoinMessage is a direct notification to a user that they have left a channel
type UserJoinMessage struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Channel  string `json:"channel"`
}

// UserLeaveMessage is a direct notification to a user that they have left a channel
type UserLeaveMessage struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Channel  string `json:"channel"`
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
	go client.messagePump()
	go client.sendConnectedMessage()

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

// messagePump pumps decoded JSON messages from the connection
//
// A message can indicate a control command (join channel, leave channel, set channel privileges)
// or it can represent a simple chat message
func (c *Client) messagePump() {
	for {
		bytes := <-c.messages

		log.Printf("Message: %s\n", string(bytes))
		var msg map[string]interface{}
		json.Unmarshal(bytes, &msg)

		msgType, ok := msg["type"]
		if !ok {
			log.Printf("[Client.messagePump] Unable to determine message type")
		} else {
			switch msgType.(string) {
			case "channels.list":
				go c.sendChannelsList()
				log.Printf("[USER: %q] Sent channels list\n", c.username)

			case "channel.chat":
				channelName := msg["channel"]
				if channelName == nil {
					log.Println("ERROR: channel broadcast command missing channel name")
					continue
				}

				text := msg["message"]
				if text == nil {
					log.Printf("ERROR: channel broadcast to channel %q missing message\n", channelName.(string))
					continue
				}

				chatMessage := &ChatMessage{Type: "channel.chat", Channel: channelName.(string), Sender: c.username, Message: text.(string)}
				if err := c.broadcastToChannel(channelName.(string), chatMessage); err != nil {
					log.Printf("ERROR: could not marshal message to channel %q -- %v\n", channelName.(string), err)
				}

			case "channel.join":
				channelName := msg["channel"]
				if channelName == nil {
					log.Println("ERROR: channel join command missing channel name")
					continue
				}

				if _, err := c.joinChannel(channelName.(string)); err != nil {
					log.Printf("ERROR: could not join channel %q -- %v\n", channelName.(string), err)
				} else {
					log.Printf("[USER: %q] Joined channel %q\n", c.username, channelName.(string))
				}

			case "channel.leave":
				channelName := msg["channel"]
				if channelName == nil {
					log.Println("ERROR: channel leave command missing channel name")
					continue
				}

				if err := c.leaveChannel(channelName.(string)); err != nil {
					log.Printf("ERROR: could not leave channel %q -- %v\n", channelName.(string), err)
				} else {
					log.Printf("[USER: %q] Left channel %q\n", c.username, channelName.(string))
				}

			default:
				log.Printf("Unhandled message type %q", msgType)
			}
		}
	}
}

func (c *Client) sendConnectedMessage() error {
	msg := ConnectedMessage{
		Type:     "user.connected",
		Username: c.username,
	}

	marshalled, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.send <- marshalled
	return nil
}

func (c *Client) broadcastToChannel(channel string, msg interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch, ok := c.channels[channel]
	if !ok {
		return fmt.Errorf("Not in channel %q", channel)
	}

	marshalled, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ch._broadcast <- marshalled
	return nil
}

func (c *Client) sendChannelsList() {
	server := getServer()
	channels := make([]*ChannelListMessage, len(server.channels))

	idx := 0
	server.mu.Lock()
	defer server.mu.Unlock()

	for _, c := range server.channels {
		channels[idx] = &ChannelListMessage{
			Channel: c.name,
		}
		idx++
	}

	response := &ChannelListResponse{
		Type:     "channels.list",
		Channels: channels,
	}

	marshalled, _ := json.Marshal(response)
	c.send <- marshalled
}

func (c *Client) sendChannelInfo(ch *Channel) {
	chanInfo := &ChannelInfoMessage{
		Type:     "channel.info",
		Channel:  ch.name,
		Username: c.username,
		Users:    make([]string, 0, len(ch.clients)),
	}
	for client := range ch.clients {
		chanInfo.Users = append(chanInfo.Users, client.username)
	}
	marshalled, _ := json.Marshal(chanInfo)
	c.send <- marshalled
}

func (c *Client) sendUserLeave(ch *Channel) {
	userLeave := &UserLeaveMessage{
		Type:     "user.leave",
		Username: c.username,
		Channel:  ch.name,
	}

	marshalled, _ := json.Marshal(userLeave)
	c.send <- marshalled
}

func (c *Client) sendUserJoin(ch *Channel) {
	userJoin := &UserJoinMessage{
		Type:     "user.join",
		Username: c.username,
		Channel:  ch.name,
	}

	marshalled, _ := json.Marshal(userJoin)
	c.send <- marshalled
}

func (c *Client) joinChannel(name string) (*Channel, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.channels[name]; ok {
		return nil, fmt.Errorf("Already in channel %q", name)
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
		return fmt.Errorf("Not in channel %v", name)
	}

	if err := ch.leave(c); err != nil {
		return err
	}

	c.sendUserLeave(ch)
	delete(c.channels, name)

	return nil
}

func (c *Client) leaveAllChannels() {
	for name := range c.channels {
		if err := c.leaveChannel(name); err != nil {
			log.Printf("ERROR: client unable to leave channel %q: %v", name, err)
		}
	}
}
