package routers

import (
	"key-value/lib/ws"
	"encoding/json"
	"fmt"
)

type client struct {
	con ws.ClientConnection
}

type Client interface {
	SendSync(r Request) (*Response, error)
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

func (c *client) SendSync(r Request) (*Response, error) {
	messageData, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	respStr, err := c.con.SendSync(string(messageData))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp := Response{}
	err = json.Unmarshal([]byte(respStr), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}