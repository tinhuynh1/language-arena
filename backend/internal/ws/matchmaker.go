package ws

import (
	"log/slog"
	"sync"

	"github.com/michael/language-arena/backend/internal/model"
)

type queueEntry struct {
	Client   *Client
	Language string
	Level    string
	QuizType model.QuizType
	Mode     model.GameMode
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

func (m *Matchmaker) Enqueue(client *Client, language, level string, quizType model.QuizType, mode model.GameMode) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, entry := range m.queue {
		if entry.Language == language && entry.Level == level && entry.QuizType == quizType && entry.Mode == mode && entry.Client.ID != client.ID {
			opponent := entry.Client

			m.queue = append(m.queue[:i], m.queue[i+1:]...)

			vocabs := m.hub.GetVocabs(language, level, maxRounds+numTargets)

			room := NewRoom(language, level, mode, quizType, vocabs, m.hub)
			room.AddPlayer(opponent)
			room.AddPlayer(client)

			m.hub.AddRoom(room)

			modeStr := string(mode)
			opponent.SendMessage(WSMessage{
				Type: MsgMatchFound,
				Data: MatchFoundData{RoomID: room.ID, Opponent: client.Username, Mode: modeStr},
			})
			client.SendMessage(WSMessage{
				Type: MsgMatchFound,
				Data: MatchFoundData{RoomID: room.ID, Opponent: opponent.Username, Mode: modeStr},
			})

			slog.Info("match found",
				"component", "WS",
				"player_1", opponent.Username,
				"player_2", client.Username,
				"room_id", room.ID,
				"mode", modeStr,
				"level", level,
			)
			return
		}
	}

	m.queue = append(m.queue, queueEntry{Client: client, Language: language, Level: level, QuizType: quizType, Mode: mode})
	client.SendMessage(WSMessage{Type: MsgQueueJoined, Data: map[string]string{"status": "waiting"}})
	slog.Info("player joined queue", "component", "WS", "player", client.Username, "language", language, "level", level, "mode", string(mode))
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
