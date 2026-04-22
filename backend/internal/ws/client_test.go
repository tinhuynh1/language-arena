package ws

import (
	"testing"
	
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestClient_Initialization(t *testing.T) {
	hub := NewHub(nil, nil, nil, nil)
	userID := uuid.New()
	c := NewClient(hub, nil, userID, "testuser")

	assert.Equal(t, userID, c.ID)
	assert.Equal(t, "testuser", c.Username)
	assert.Equal(t, hub, c.Hub)
	assert.NotNil(t, c.Send)
}

func TestClient_Room(t *testing.T) {
	c := &Client{}
	
	// Room shouldn't be set
	assert.Nil(t, c.GetRoom())

	room := &Room{Code: "ROOM1"}
	c.SetRoom(room)

	assert.Equal(t, room, c.GetRoom())
}

func TestClient_SendMessage(t *testing.T) {
	c := &Client{Send: make(chan []byte, 10)}

	c.SendMessage(WSMessage{Type: "test_msg", Data: "hello"})
	
	msg := <-c.Send
	assert.Contains(t, string(msg), "test_msg")
	assert.Contains(t, string(msg), "hello")
}

func TestClient_RelayFunc(t *testing.T) {
	c := &Client{}

	relayed := false
	c.RelayFunc = func(msg WSMessage) {
		relayed = true
		assert.Equal(t, "proxy_test", string(msg.Type))
	}

	c.SendMessage(WSMessage{Type: "proxy_test"})
	assert.True(t, relayed)
}

func TestClient_Pumps(t *testing.T) {
	c := &Client{Send: make(chan []byte)}
	close(c.Send)
	assert.Panics(t, func() {
		c.ReadPump()
	})
	assert.Panics(t, func() {
		c.WritePump()
	})
}
