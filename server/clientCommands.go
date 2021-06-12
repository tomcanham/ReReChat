package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// ChatMessage is the deserialized JSON object representing normal chat messages
type ChatMessage struct {
	Sender  string `json:"sender"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

func (m *ChatMessage) TypeName() string {
	return "channel.chat"
}

// ChannelsListMessage is the command to retrieve joined channels
type ChannelsListMessage struct {
	Channels []string `json:"channels"`
}

func (m *ChannelsListMessage) TypeName() string {
	return "channels.list"
}

// ChannelInfoMessage is the initial channel information sent to newly-joined users
type ChannelInfoMessage struct {
	Channel  string   `json:"name"`
	Username string   `json:"username"`
	Users    []string `json:"users"`
}

func (m *ChannelInfoMessage) TypeName() string {
	return "channel.info"
}

// UserJoinMessage is a direct notification to a user that they have left a channel
type UserJoinMessage struct {
	Username string `json:"username"`
	Channel  string `json:"channel"`
}

func (m *UserJoinMessage) TypeName() string {
	return "user.join"
}

// UserLeaveMessage is a direct notification to a user that they have left a channel
type UserLeaveMessage struct {
	Username string `json:"username"`
	Channel  string `json:"channel"`
}

func (m *UserLeaveMessage) TypeName() string {
	return "user.leave"
}

func (c *Client) unmarshalSelfDescribing(bytes []byte) (interface{}, error) {
	lines := strings.SplitN(string(bytes), "\n", 2)
	msgType := lines[0]
	rest := []byte(lines[1])

	switch msgType {
	case "channels.list":
		return ChannelsListMessage{}, nil

	case "channel.chat":
		var chatMessage ChatMessage
		if err := json.Unmarshal(rest, &chatMessage); err != nil {
			return nil, err
		}

		chatMessage.Sender = c.username
		return chatMessage, nil

	case "channel.join":
		var joinMsg ChannelJoinMessage
		if err := json.Unmarshal(rest, &joinMsg); err != nil {
			return nil, err
		}

		joinMsg.Username = c.username
		return joinMsg, nil

	case "channel.leave":
		var leaveMsg ChannelLeaveMessage
		if err := json.Unmarshal(rest, &leaveMsg); err != nil {
			return nil, err
		}

		leaveMsg.Username = c.username
		return leaveMsg, nil

	default:
		return nil, error(fmt.Errorf("unhandled message type %q", msgType))
	}
}

// commandPump pumps decoded JSON messages from the connection
//
// A message can indicate a control command (join channel, leave channel, set channel privileges)
// or it can represent a simple chat message
func (c *Client) commandPump() {
	for {
		bytes := <-c.messages

		msg, err := c.unmarshalSelfDescribing(bytes)
		if err != nil {
			log.Printf("error unmarshaling client command -- %v\n", err)
		}

		switch msg := msg.(type) {
		case ChannelsListMessage:
			c.sendChannelsList()

		case ChatMessage:
			if err := c.broadcastToChannel(msg.Channel, &msg); err != nil {
				log.Printf("error: could not broadcast message to channel -- %v\n", err)
			}

		case ChannelJoinMessage:
			if _, err := c.joinChannel(msg.Channel); err != nil {
				log.Printf("error: could not join channel %q -- %v\n", msg.Channel, err)
			} else {
				log.Printf("[USER: %q] Joined channel %q\n", c.username, msg.Channel)
			}

		case ChannelLeaveMessage:
			if err := c.leaveChannel(msg.Channel); err != nil {
				log.Printf("error: could not leave channel %q -- %v\n", msg.Channel, err)
			} else {
				log.Printf("[USER: %q] Left channel %q\n", c.username, msg.Channel)
			}

		default:
			log.Printf("error -- unknown message type: %+v\n", msg)
		}
	}
}
