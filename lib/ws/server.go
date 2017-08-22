package ws

import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"time"
	"bytes"
	"encoding/json"
	"fmt"
)

type handler func(message []byte, sendQueue chan []byte)
type RequestHandler func(r []byte) []byte

type WebSocketServer interface {
	ServeWebSocket(w http.ResponseWriter, r *http.Request, handler RequestHandler)
}

type webSocketServer struct {
	upgrader websocket.Upgrader
}

type connection struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *connection) runRead(handler handler) {
	defer func() {
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		go handler(bytes.TrimSpace(bytes.Replace(message, newline, space, -1)), c.send)
	}
}

func (c *connection) runWrite() {
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func NewWebSocketServer() WebSocketServer {
	return &webSocketServer{
		websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func createWebSocketRequestHandler(rh RequestHandler) handler {
	return func(message []byte, sendQueue chan []byte) {
		var r request
		err := json.Unmarshal(message, &r)
		if err != nil {
			msg := fmt.Sprintf(`Message: '%s' parse failed: %s`, message, err.Error())
			log.Println(msg)
			sendQueue <- []byte(msg)
			return
		}

		resp := response{
			RequestId: r.RequestId,
			Payload: string(rh([]byte(r.Payload))),
		}

		responseJson, err := json.Marshal(resp)
		if err != nil {
			msg := fmt.Sprintf(`Message: '%s' parse failed: %s`, message, err.Error())
			log.Println(msg)
			sendQueue <- []byte(msg)
			return
		}

		sendQueue <- responseJson
	}
}

func (s *webSocketServer) ServeWebSocket(w http.ResponseWriter, r *http.Request, rh RequestHandler) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &connection{conn: conn, send: make(chan []byte, 256)}

	go client.runWrite()
	go client.runRead(createWebSocketRequestHandler(rh))
}