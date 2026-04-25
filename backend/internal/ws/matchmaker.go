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
	queue   []queueEntry
	pending map[string]queueEntry // userID → entry for players waiting in Redis queue
	mu      sync.Mutex
	hub     *Hub
}

func NewMatchmaker(hub *Hub) *Matchmaker {
	return &Matchmaker{
		queue:   make([]queueEntry, 0),
		pending: make(map[string]queueEntry),
		hub:     hub,
	}
}

func (m *Matchmaker) Enqueue(client *Client, language, level string, quizType model.QuizType, mode model.GameMode) {
	if m.hub.Redis != nil {
		m.enqueueRedis(client, language, level, quizType, mode)
		return
	}
	m.enqueueLocal(client, language, level, quizType, mode)
}

func (m *Matchmaker) enqueueRedis(client *Client, language, level string, quizType model.QuizType, mode model.GameMode) {
	opponent, err := m.hub.Redis.EnqueueOrMatch(
		client.ID.String(), client.Username,
		language, level, string(quizType), string(mode),
	)
	if err != nil {
		slog.Error("matchmaker redis error, falling back to local queue", "component", "WS", "err", err)
		m.enqueueLocal(client, language, level, quizType, mode)
		return
	}

	if opponent == nil {
		// No match yet — track locally so Remove() can clean up Redis
		m.mu.Lock()
		m.pending[client.ID.String()] = queueEntry{Client: client, Language: language, Level: level, QuizType: quizType, Mode: mode}
		m.mu.Unlock()

		client.SendMessage(WSMessage{Type: MsgQueueJoined, Data: map[string]string{"status": "waiting"}})
		slog.Info("player joined redis queue", "component", "WS", "player", client.Username, "language", language, "level", level, "mode", string(mode))
		return
	}

	// Match found: create room owned by this instance
	vocabs := m.hub.GetVocabs(language, level, maxRounds+numTargets)
	room := NewRoom(language, level, mode, quizType, vocabs, m.hub)
	room.AddPlayer(client)
	m.hub.AddRoom(room)
	if m.hub.Redis != nil {
		m.hub.Redis.RegisterRoom(room.Code, m.hub.Redis.NodeID)
	}

	modeStr := string(mode)

	if opponent.NodeID == m.hub.Redis.NodeID {
		// Both players are on this instance.
		// Try pending map first; fall back to h.Clients to handle the race where
		// the opponent's goroutine hasn't yet written to pending after EnqueueOrMatch returned.
		m.mu.Lock()
		localEntry, ok := m.pending[opponent.UserID]
		if ok {
			delete(m.pending, opponent.UserID)
		}
		m.mu.Unlock()

		var localClient *Client
		if ok {
			localClient = localEntry.Client
		} else {
			m.hub.mu.RLock()
			for c := range m.hub.Clients {
				if c.ID.String() == opponent.UserID {
					localClient = c
					break
				}
			}
			m.hub.mu.RUnlock()
		}

		if localClient != nil {
			room.AddPlayer(localClient)
			localClient.SendMessage(WSMessage{
				Type: MsgMatchFound,
				Data: MatchFoundData{RoomID: room.ID, Opponent: client.Username, Mode: modeStr},
			})
		}
	} else {
		// Opponent is on a different instance — ask that node to proxy-join them
		m.hub.Redis.PublishToNode(opponent.NodeID, RedisMessage{
			Type:       RedisMatchFound,
			FromNode:   m.hub.Redis.NodeID,
			RoomCode:   room.Code,
			UserID:     opponent.UserID,
			Username:   client.Username, // our username = their opponent
			ActionType: modeStr,
		})
	}

	client.SendMessage(WSMessage{
		Type: MsgMatchFound,
		Data: MatchFoundData{RoomID: room.ID, Opponent: opponent.Username, Mode: modeStr},
	})

	slog.Info("match found via redis",
		"component", "WS",
		"player_1", client.Username,
		"player_2", opponent.Username,
		"room_id", room.ID,
		"mode", modeStr,
		"level", level,
	)
}

func (m *Matchmaker) enqueueLocal(client *Client, language, level string, quizType model.QuizType, mode model.GameMode) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Dedup: skip if already waiting
	for _, entry := range m.queue {
		if entry.Client.ID == client.ID {
			return
		}
	}

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

	// Clean up Redis queue entry if present
	if m.hub.Redis != nil {
		if entry, ok := m.pending[client.ID.String()]; ok {
			delete(m.pending, client.ID.String())
			m.hub.Redis.RemoveFromQueue(
				client.ID.String(),
				entry.Language, entry.Level,
				string(entry.QuizType), string(entry.Mode),
			)
		}
	}

	// Clean up local queue
	for i, entry := range m.queue {
		if entry.Client.ID == client.ID {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			return
		}
	}
}

func (m *Matchmaker) removePending(userID string) {
	m.mu.Lock()
	delete(m.pending, userID)
	m.mu.Unlock()
}
