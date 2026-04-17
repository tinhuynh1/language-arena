package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/service"
)

type Hub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Rooms      map[string]*Room
	RoomByCode map[string]*Room
	Matchmaker *Matchmaker

	vocabService *service.VocabService

	mu sync.RWMutex
}

func NewHub(vocabService *service.VocabService) *Hub {
	h := &Hub{
		Clients:      make(map[*Client]bool),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Rooms:        make(map[string]*Room),
		RoomByCode:   make(map[string]*Room),
		vocabService: vocabService,
	}
	h.Matchmaker = NewMatchmaker(h)
	return h
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected: %s (%s). Total: %d", client.Username, client.ID, len(h.Clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

			h.Matchmaker.Remove(client)
			if room := client.GetRoom(); room != nil {
				room.RemovePlayer(client)
			}
			log.Printf("Client disconnected: %s. Total: %d", client.Username, len(h.Clients))
		}
	}
}

func (h *Hub) HandleMessage(client *Client, msg WSMessage) {
	switch msg.Type {
	case MsgJoinQueue:
		h.handleJoinQueue(client, msg)
	case MsgCreateRoom:
		h.handleCreateRoom(client, msg)
	case MsgJoinRoom:
		h.handleJoinRoom(client, msg)
	case MsgStartGame:
		h.handleStartGame(client)
	case MsgReady:
		h.handleReady(client)
	case MsgTargetHit:
		h.handleTargetHit(client, msg)
	case MsgLeaveRoom:
		h.handleLeaveRoom(client)
	default:
		client.SendMessage(WSMessage{Type: MsgError, Data: "unknown message type"})
	}
}

func (h *Hub) handleJoinQueue(client *Client, msg WSMessage) {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		client.SendMessage(WSMessage{Type: MsgError, Data: "invalid join data"})
		return
	}

	var data JoinQueueData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		client.SendMessage(WSMessage{Type: MsgError, Data: "invalid join data"})
		return
	}

	if data.Language == "" {
		data.Language = "en"
	}
	if data.Level == "" {
		data.Level = "A1"
	}

	if data.Mode == "solo" {
		vocabs := h.GetVocabs(data.Language, data.Level, maxRounds+numTargets)
		room := NewRoom(data.Language, data.Level, model.ModeSolo, vocabs)
		room.AddPlayer(client)
		h.AddRoom(room)

		client.SendMessage(WSMessage{
			Type: MsgMatchFound,
			Data: MatchFoundData{RoomID: room.ID, Opponent: "", Mode: "solo"},
		})
		return
	}

	h.Matchmaker.Enqueue(client, data.Language, data.Level)
}

func (h *Hub) handleCreateRoom(client *Client, msg WSMessage) {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		client.SendMessage(WSMessage{Type: MsgError, Data: "invalid create data"})
		return
	}

	var data CreateRoomData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		client.SendMessage(WSMessage{Type: MsgError, Data: "invalid create data"})
		return
	}

	if data.Language == "" {
		data.Language = "en"
	}
	if data.Level == "" {
		data.Level = "A1"
	}

	vocabs := h.GetVocabs(data.Language, data.Level, maxRounds+numTargets)
	room := NewRoom(data.Language, data.Level, model.ModeBattle, vocabs)
	room.HostID = client.ID
	room.AddPlayer(client)
	h.AddRoom(room)

	client.SendMessage(WSMessage{
		Type: MsgRoomCreated,
		Data: RoomCreatedData{
			RoomCode: room.Code,
			RoomID:   room.ID,
			Language: data.Language,
			Level:    data.Level,
		},
	})

	log.Printf("Battle room %s (code: %s) created by %s", room.ID, room.Code, client.Username)
}

func (h *Hub) handleJoinRoom(client *Client, msg WSMessage) {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		client.SendMessage(WSMessage{Type: MsgError, Data: "invalid join data"})
		return
	}

	var data JoinRoomData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		client.SendMessage(WSMessage{Type: MsgError, Data: "invalid join data"})
		return
	}

	h.mu.RLock()
	room, ok := h.RoomByCode[data.RoomCode]
	h.mu.RUnlock()

	if !ok {
		client.SendMessage(WSMessage{Type: MsgError, Data: "room not found"})
		return
	}

	if !room.AddPlayer(client) {
		client.SendMessage(WSMessage{Type: MsgError, Data: "room is full or game started"})
		return
	}

	// Notify joining player
	client.SendMessage(WSMessage{
		Type: MsgMatchFound,
		Data: MatchFoundData{
			RoomID:      room.ID,
			PlayerCount: len(room.Players),
			Mode:        string(room.Mode),
		},
	})

	// Notify all players about new join
	names := room.getPlayerNames()
	room.broadcast(WSMessage{
		Type: MsgPlayerJoined,
		Data: PlayerJoinedData{
			Username:    client.Username,
			PlayerCount: len(room.Players),
			Players:     names,
		},
	})

	log.Printf("Player %s joined room %s (code: %s). Total: %d", client.Username, room.ID, room.Code, len(room.Players))
}

func (h *Hub) handleStartGame(client *Client) {
	room := client.GetRoom()
	if room == nil {
		client.SendMessage(WSMessage{Type: MsgError, Data: "not in a room"})
		return
	}
	room.StartByHost(client)
}

func (h *Hub) handleReady(client *Client) {
	room := client.GetRoom()
	if room == nil {
		client.SendMessage(WSMessage{Type: MsgError, Data: "not in a room"})
		return
	}
	room.SetReady(client)
}

func (h *Hub) handleTargetHit(client *Client, msg WSMessage) {
	room := client.GetRoom()
	if room == nil {
		return
	}

	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return
	}

	var data TargetHitData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return
	}

	room.HandleHit(client, data)
}

func (h *Hub) handleLeaveRoom(client *Client) {
	room := client.GetRoom()
	if room != nil {
		room.RemovePlayer(client)
		client.SetRoom(nil)
	}
	h.Matchmaker.Remove(client)
}

func (h *Hub) GetVocabs(language, level string, count int) []model.Vocabulary {
	vocabs, err := h.vocabService.GetRandomSet(context.Background(), language, level, count)
	if err != nil || len(vocabs) == 0 {
		log.Printf("failed to get vocabs: %v", err)
		return []model.Vocabulary{
			{Word: "hello", Meaning: "xin chào", Language: "en", Level: "A1"},
			{Word: "world", Meaning: "thế giới", Language: "en", Level: "A1"},
		}
	}
	return vocabs
}

func (h *Hub) AddRoom(room *Room) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Rooms[room.ID] = room
	h.RoomByCode[room.Code] = room
}

func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}
