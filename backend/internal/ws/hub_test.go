package ws

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHub_BasicOperations(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	assert.NotNil(t, h)
	assert.NotNil(t, h.Matchmaker)

	// Test Add Room
	room := &Room{
		ID:          "room-123",
		Code:        "ABCD",
		Language:    "en",
		TotalRounds: 5,
	}
	h.AddRoom(room)

	h.mu.RLock()
	assert.Equal(t, room, h.Rooms["room-123"])
	assert.Equal(t, room, h.RoomByCode["ABCD"])
	h.mu.RUnlock()

	// Test Remove Room
	h.RemoveRoom(room)

	h.mu.RLock()
	_, found1 := h.Rooms["room-123"]
	_, found2 := h.RoomByCode["ABCD"]
	assert.False(t, found1)
	assert.False(t, found2)
	h.mu.RUnlock()

	// Test Online Count
	assert.Equal(t, 0, h.GetOnlineCount())
	c := &Client{ID: uuid.New()}
	h.mu.Lock()
	h.Clients[c] = true
	h.mu.Unlock()
	assert.Equal(t, 1, h.GetOnlineCount())
}



func TestDisconnectedPlayer_GracePeriod(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)

	room := &Room{Code: "GRACE", Players: make(map[*Client]*PlayerState)}
	client := &Client{ID: uuid.New(), Username: "test_dc"}
	
	h.startGracePeriod(client, room, room.Code)

	h.mu.RLock()
	dp, exists := h.disconnectedPlayers[client.ID]
	h.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, client, dp.Client)
	assert.Equal(t, room, dp.Room)

	// Test TryReconnect same pod
	newClient := &Client{ID: client.ID, Username: "test_reconnected", Send: make(chan []byte, 10)}
	
	// Wait, room ReconnectPlayer mock
	room.Players[client] = &PlayerState{Client: client}
	
	// Mock ReconnectPlayer method on room
	dp.Room.Players[newClient] = dp.Room.Players[client]
	delete(dp.Room.Players, client)
	
	// Actually tryReconnect calls dp.Room.ReconnectPlayer
	// let's just make sure tryReconnect doesn't panic
	h.tryReconnect(newClient)
}
