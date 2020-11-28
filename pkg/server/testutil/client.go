package testutil

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/armsnyder/othelgo/pkg/messages"
)

type Client struct {
	handler               *Handler
	connectionID          string
	messagesSinceLastSend []interface{}
}

func (c *Client) Connect() {
	if c.connectionID != "" {
		return
	}

	var connectionIDSource [9]byte
	_, err := rand.Read(connectionIDSource[:])
	if err != nil {
		panic(err)
	}
	c.connectionID = base64.URLEncoding.EncodeToString(connectionIDSource[:])
	c.handler.invoke("CONNECT", "", c.connectionID)
}

func (c *Client) Disconnect() {
	if c.connectionID == "" {
		return
	}

	c.handler.invoke("DISCONNECT", "", c.connectionID)
	c.connectionID = ""
}

func (c *Client) Send(message interface{}) {
	if c.connectionID == "" {
		panic(errors.New("client is not connected"))
	}

	wrapper := messages.Wrapper{Message: message}

	raw, err := json.Marshal(wrapper)
	if err != nil {
		panic(err)
	}

	c.handler.invoke("MESSAGE", string(raw), c.connectionID)
}

func (c *Client) resetReceivedMessages() {
	c.messagesSinceLastSend = nil
}

func (c *Client) addReceivedMessage(data []byte) {
	var wrapper messages.Wrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		panic(err)
	}
	c.messagesSinceLastSend = append(c.messagesSinceLastSend, wrapper.Message)
}
