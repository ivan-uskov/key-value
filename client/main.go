package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"github.com/gorilla/websocket"
	"encoding/json"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

type request struct {
	RequestId int
	Payload string
}

type response struct {
	RequestId int
	Payload string
}

type QueryResult struct {

}

type connection struct {
	requestId int
	queries map[int](chan string)
	ws *websocket.Conn
	done chan struct{}
}

func createConnection(address string, path string) *connection {
	u := url.URL{Scheme: "ws", Host: address, Path: "/" + path}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	return &connection{
		requestId: 0,
		queries: make(map[int](chan string)),
		ws: c,
		done: make(chan struct{}),
	}
}

func (c *connection) handleResponse(res *response) {
	resChan, ok := c.queries[res.RequestId]
	if ok {
		resChan <- res.Payload
		delete(c.queries, res.RequestId)
	}
}

func (c *connection) runReader() {
	go func() {
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
			}

			c.handleResponse(&response)
		}
	}()
}

func (c *connection) send(msg string) (chan string, error) {
	c.requestId++
	req := request{
		RequestId: c.requestId,
		Payload: msg,
	}
	message, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(30)
	result := make(chan string)
	c.ws.WriteMessage(websocket.TextMessage, message)
	c.queries[req.RequestId] = result
	go func () {
		select {
		case <-time.After(time.Second * timeout):
			log.Println("Request closed by timeout: ", req.RequestId)
			delete(c.queries, req.RequestId)
		}
	}()
	return nil, nil
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c := createConnection(*addr, `ws`)
	c.runReader()
	c.send(`hello`)
}