package ws

import (
	"log"
	"sync"

	"github.com/michael/language-arena/backend/internal/model"
)

type queueEntry struct {
	Client   *Client
	Language string
	Level    string
}

type Matchmaker struct {
	queue []queueEntry
	mu    sync.Mutex
	hub   *Hub
}

func NewMatchmaker(hub *Hub) *Matchmaker {
	return &Matchmaker{
		queue: make([]queueEntry, 0),
		hub:   hub,
	}
}

func (m *Matchmaker) Enqueue(client *Client, language, level string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, entry := range m.queue {
		if entry.Language == language && entry.Level == level && entry.Client.ID != client.ID {
			opponent := entry.Client

			m.queue = append(m.queue[:i], m.queue[i+1:]...)

			vocabs := m.hub.GetVocabs(language, level, maxRounds+numTargets)

			room := NewRoom(language, level, model.ModeDuel, vocabs, m.hub)
			room.AddPlayer(opponent)
			room.AddPlayer(client)

			m.hub.AddRoom(room)

			opponent.SendMessage(WSMessage{
				Type: MsgMatchFound,
				Data: MatchFoundData{RoomID: room.ID, Opponent: client.Username, Mode: "duel"},
			})
			client.SendMessage(WSMessage{
				Type: MsgMatchFound,
				Data: MatchFoundData{RoomID: room.ID, Opponent: opponent.Username, Mode: "duel"},
			})

			log.Printf("Match found: %s vs %s in room %s (level: %s)", opponent.Username, client.Username, room.ID, level)
			return
		}
	}

	m.queue = append(m.queue, queueEntry{Client: client, Language: language, Level: level})
	client.SendMessage(WSMessage{Type: MsgQueueJoined, Data: map[string]string{"status": "waiting"}})
	log.Printf("Player %s joined queue for %s/%s", client.Username, language, level)
}

func (m *Matchmaker) Remove(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, entry := range m.queue {
		if entry.Client.ID == client.ID {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			return
		}
	}
}
