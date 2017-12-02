package ws

import (
	"net/url"
	"log"
	"github.com/gorilla/websocket"
	"encoding/json"
	"time"
	"errors"
)

type ClientConnection interface {
	Send(msg string) (<-chan string, error)
	SendSync(msg string) (string, error)
}

type clientConnection struct {
	requestId int64
	queries   map[int64]chan string
	ws        *websocket.Conn
	done      chan struct{}
}

func (c *clientConnection) handleResponse(res *response) {
	resChan, ok := c.queries[res.RequestId]
	if ok {
		resChan <- res.Payload
		delete(c.queries, res.RequestId)
	} else {
		log.Println(`request not found `, res.RequestId)
	}
}

func (c *clientConnection) runReader() {
	defer c.ws.Close()
	defer close(c.done)
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		response := response{}
		err = json.Unmarshal(message, &response)
		if err != nil {
			log.Println("unmarshall:", err)
			return
		}

		c.handleResponse(&response)
	}
}

func (c *clientConnection) Send(msg string) (<-chan string, error) {
	c.requestId++
	req := request{
		RequestId: c.requestId,
		Payload:   msg,
	}
	message, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(10)
	resultChan := make(chan string)
	c.ws.WriteMessage(websocket.TextMessage, message)
	c.queries[req.RequestId] = resultChan
	go func() {
		select {
		case <-time.After(time.Second * timeout):
			_, ok := c.queries[req.RequestId]
			if ok {
				log.Println("Request closed by timeout: ", req.RequestId)
				delete(c.queries, req.RequestId)
			}
		}
	}()
	return resultChan, nil
}

func (c *clientConnection) SendSync(msg string) (string, error) {
	mc, err := c.Send(msg)
	if err != nil {
		return "", err
	}

	select {
	case <-time.After(10 * time.Second):
		return "", errors.New("timeout")
	case result := <-mc:
		return result, nil
	}
}

func NewClient(address string, path string) ClientConnection {
	u := url.URL{Scheme: "ws", Host: address, Path: "/" + path}
	log.Printf("connecting to %s", u.String())

	rawConnection, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	con := &clientConnection{
		requestId: 0,
		queries:   make(map[int64](chan string)),
		ws:        rawConnection,
		done:      make(chan struct{}),
	}

	go con.runReader()
	return con
}
