package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMatchmaker_Enqueue_Remove(t *testing.T) {
	h := NewHub(nil, nil, nil, nil)
	m := NewMatchmaker(h)

	c1 := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10)}
	c2 := &Client{ID: uuid.New(), Username: "p2", Send: make(chan []byte, 10)}

	// Enqueue player 1
	m.Enqueue(c1, "en", "A1", model.QuizTypeMeaningToWord, model.ModeDuel)
	
	m.mu.Lock()
	assert.Len(t, m.queue, 1)
	m.mu.Unlock()

	// Wait, match should NOT happen yet
	
	// Remove player 1
	m.Remove(c1)
	m.mu.Lock()
	assert.Len(t, m.queue, 0)
	m.mu.Unlock()

	// Enqueue player 1 again
	m.Enqueue(c1, "en", "A1", model.QuizTypeMeaningToWord, model.ModeDuel)

	// Enqueue player 2
	m.Enqueue(c2, "en", "A1", model.QuizTypeMeaningToWord, model.ModeDuel)

	// They should match
	m.mu.Lock()
	assert.Len(t, m.queue, 0)
	m.mu.Unlock()

	// Since they matched, they should both receive MsgMatchFound
	// wait briefly for the messages to enter the channels
	time.Sleep(10 * time.Millisecond)

	checkMatch := func(c *Client) string {
		for {
			select {
			case msgBytes := <-c.Send:
				var m WSMessage
				json.Unmarshal(msgBytes, &m)
				if m.Type == MsgMatchFound {
					return string(m.Type)
				}
			case <-time.After(100 * time.Millisecond):
				return "timeout"
			}
		}
	}

	assert.Equal(t, string(MsgMatchFound), checkMatch(c1))
	assert.Equal(t, string(MsgMatchFound), checkMatch(c2))
}
