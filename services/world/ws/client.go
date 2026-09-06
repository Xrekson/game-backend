package ws

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/services/world/model"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow CORS for game client
	},
}

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	PlayerID string
	send     chan []byte
}

func (c *Client) ReadPump(onMessage func(c *Client, msg *model.WSMessage)) {
	defer func() {
		c.Hub.unregister <- c
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WS read error: %v", err)
			}
			break
		}

		var msg model.WSMessage
		if err := json.Unmarshal(messageBytes, &msg); err == nil {
			onMessage(c, &msg)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, playerID string, onMessage func(c *Client, msg *model.WSMessage)) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WS Upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:      hub,
		Conn:     conn,
		PlayerID: playerID,
		send:     make(chan []byte, 256),
	}

	client.Hub.register <- client

	go client.WritePump()
	go client.ReadPump(onMessage)
}
