package ws

import (
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHub_HandleMessage_JoinQueue(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	c := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10), Hub: h}
	h.Register <- c
	time.Sleep(10 * time.Millisecond)

	msg := WSMessage{
		Type: MsgJoinQueue,
		Data: map[string]interface{}{
			"language": "en",
			"level":    "A1",
			"quizType": "meaning_to_word",
			"mode":     "duel",
		},
	}

	h.HandleMessage(c, msg)
	time.Sleep(10 * time.Millisecond)

	// Since vocabService is nil, Matchmaker will add to queue.
	h.Matchmaker.mu.Lock()
	assert.Len(t, h.Matchmaker.queue, 1)
	h.Matchmaker.mu.Unlock()
}

func TestHub_HandleMessage_CreateRoom(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	// Suppress logs for testing error paths if needed
	slog.SetDefault(slog.New(slog.DiscardHandler))

	c := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10), Hub: h}
	
	msg := WSMessage{
		Type: MsgCreateRoom,
		Data: map[string]interface{}{
			"language": "en",
			"level":    "A1",
			"quizType": "meaning_to_word",
			"mode":     "battle",
		},
	}

	h.HandleMessage(c, msg)
	time.Sleep(10 * time.Millisecond)

	// Should create a room
	assert.NotNil(t, c.Room)
	assert.Equal(t, "p1", c.Room.GetHostUsername())
}

func TestHub_HandleMessage_JoinRoom(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	room := NewRoom("en", "A1", "battle", "meaning_to_word", nil, h)
	h.AddRoom(room)

	c := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10), Hub: h}
	
	msg := WSMessage{
		Type: MsgJoinRoom,
		Data: map[string]interface{}{
			"room_code": room.Code,
		},
	}

	h.HandleMessage(c, msg)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, room, c.Room)
}

func TestHub_HandleMessage_LeaveRoom(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	room := NewRoom("en", "A1", "battle", "meaning_to_word", nil, h)
	h.AddRoom(room)

	c := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10), Hub: h, Room: room}
	room.AddPlayer(c)

	msg := WSMessage{Type: MsgLeaveRoom}
	h.HandleMessage(c, msg)

	time.Sleep(10 * time.Millisecond)

	assert.Nil(t, c.Room)
}

func TestHub_HandleMessage_Ready(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	room := NewRoom("en", "A1", "battle", "meaning_to_word", nil, h)
	h.AddRoom(room)

	c := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10), Hub: h, Room: room}
	room.AddPlayer(c)

	msg := WSMessage{Type: MsgReady}
	h.HandleMessage(c, msg)

	room.mu.Lock()
	ready := room.Players[c].Ready
	room.mu.Unlock()

	assert.True(t, ready)
}

func TestHub_HandleMessage_StartGame(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	room := NewRoom("en", "A1", "battle", "meaning_to_word", nil, h)
	h.AddRoom(room)

	c := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10), Hub: h, Room: room}
	room.AddPlayer(c)

	msg := WSMessage{Type: MsgStartGame}
	h.HandleMessage(c, msg)

	time.Sleep(10 * time.Millisecond)
	// We just ensure it routes without panic
}

func TestHub_HandleMessage_TargetHit(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	room := NewRoom("en", "A1", "battle", "meaning_to_word", nil, h)
	h.AddRoom(room)

	c := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10), Hub: h, Room: room}
	room.AddPlayer(c)

	msg := WSMessage{
		Type: MsgTargetHit, 
		Data: map[string]interface{}{
			"targetId": "123",
			"reactionMs": float64(100),
		},
	}
	h.HandleMessage(c, msg)
}

func TestHub_ProxyAction(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	c := &Client{ID: uuid.New(), Username: "p1", Hub: h}
    msg := WSMessage{
        Type: "proxy_join",
        Data: map[string]interface{}{"room_code": "XYZ", "node_id": "test_node"},
    }
    h.HandleMessage(c, msg)
}
