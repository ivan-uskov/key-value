package routers

import (
	"key-value/lib/ws"
	"encoding/json"
)

type client struct {
	con ws.ClientConnection
}

type Client interface {
	Send(r Request) (<-chan string, error)
	SendSync(r Request) (string, error)
	Close()
}

func NewClient(address string, path string) (Client, error) {
	con, err := ws.NewClient(address, path)
	if err != nil {
		return nil, err
	}

	return &client{con}, nil
}

func (c * client) Close() {
	c.con.Close()
}

func (c *client) Send(r Request) (<-chan string, error) {
	messageData, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return c.con.Send(string(messageData))
}

func (c *client) SendSync(r Request) (string, error) {
	messageData, err := json.Marshal(r)
	if err != nil {
		return ``, err
	}

	return c.con.SendSync(string(messageData))
}