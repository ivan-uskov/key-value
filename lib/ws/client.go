package ws

import (
	"net/url"
	"github.com/gorilla/websocket"
	"encoding/json"
	"time"
	"errors"
	"sync"
)

type ClientConnection interface {
	Send(msg string) (<-chan string, error)
	SendSync(msg string, timeout time.Duration) (string, error)
	Close()
}

type clientConnection struct {
	requestID int64
	queries   sync.Map
	ws        *websocket.Conn
	done      chan bool
}

func (c *clientConnection) handleResponse(res *response) {
	resChan, ok := c.queries.Load(res.RequestID)
	if ok {
		resChan.(chan string) <- res.Payload
		c.queries.Delete(res.RequestID)
	}
}

func (c *clientConnection) runReader() {
	go func() {
		for {
			_, message, err := c.ws.ReadMessage()
			if err != nil {
				return
			}
			resp := response{}
			err = json.Unmarshal(message, &resp)
			if err != nil {
				return
			}

			c.handleResponse(&resp)
			select {
			case <-c.done:
				return
			default:
				continue
			}
		}
	}()
}

func (c *clientConnection) Close() {
	c.done <- true
	c.ws.Close()
}

const TIMEOUT  = 60

func (c *clientConnection) Send(msg string) (<-chan string, error) {
	c.requestID++
	req := request{
		RequestID: c.requestID,
		Payload:   msg,
	}
	message, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resultChan := make(chan string, 1)
	c.ws.WriteMessage(websocket.TextMessage, message)
	c.queries.Store(req.RequestID, resultChan)
	go func() {
		select {
		case <-time.After(time.Second * TIMEOUT):
			c.queries.Delete(req.RequestID)
		}
	}()
	return resultChan, nil
}

func (c *clientConnection) SendSync(msg string, timeout time.Duration) (string, error) {
	mc, err := c.Send(msg)
	if err != nil {
		return "", err
	}

	select {
	case <-time.After(timeout):
		return ``, errors.New("timeout")
	case result := <-mc:
		return result, nil
	}
}

func NewClient(address string, path string) (ClientConnection, error) {
	u := url.URL{Scheme: "ws", Host: address, Path: "/" + path}
	rawConnection, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	con := &clientConnection{
		requestID: 0,
		queries:   sync.Map{},
		ws:        rawConnection,
		done:      make(chan bool, 1),
	}

	con.runReader()
	return con, nil
}
