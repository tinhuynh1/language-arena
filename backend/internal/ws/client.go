package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	ID       uuid.UUID
	Username string
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	Room     *Room
	mu       sync.Mutex

	// Proxy support: if set, SendMessage uses this instead of WebSocket
	RelayFunc func(WSMessage)
	IsProxy   bool
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, username string) *Client {
	return &Client{
		ID:       userID,
		Username: username,
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("ws read error", "component", "WS", "user_id", c.ID, "player", c.Username, "err", err)
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.SendMessage(WSMessage{Type: MsgError, Data: "invalid message format"})
			continue
		}

		c.Hub.HandleMessage(c, msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) SendMessage(msg WSMessage) {
	if c.RelayFunc != nil {
		c.RelayFunc(msg)
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.Send <- data:
	default:
		slog.Warn("client send buffer full", "component", "WS", "user_id", c.ID, "player", c.Username)
	}
}

func (c *Client) SetRoom(room *Room) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Room = room
}

func (c *Client) GetRoom() *Room {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Room
}
