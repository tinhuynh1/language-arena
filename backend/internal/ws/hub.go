package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/repository"
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
	gameRepo     *repository.GameRepository
	userRepo     *repository.UserRepository

	mu sync.RWMutex
}

func NewHub(vocabService *service.VocabService, gameRepo *repository.GameRepository, userRepo *repository.UserRepository) *Hub {
	h := &Hub{
		Clients:      make(map[*Client]bool),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Rooms:        make(map[string]*Room),
		RoomByCode:   make(map[string]*Room),
		vocabService: vocabService,
		gameRepo:     gameRepo,
		userRepo:     userRepo,
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
		room := NewRoom(data.Language, data.Level, model.ModeSolo, model.QuizType(data.QuizType), vocabs, h)
		room.AddPlayer(client)
		h.AddRoom(room)

		client.SendMessage(WSMessage{
			Type: MsgMatchFound,
			Data: MatchFoundData{RoomID: room.ID, Opponent: "", Mode: "solo"},
		})
		return
	}

	h.Matchmaker.Enqueue(client, data.Language, data.Level, model.QuizType(data.QuizType))
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
	room := NewRoom(data.Language, data.Level, model.ModeBattle, model.QuizType(data.QuizType), vocabs, h)
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

func (h *Hub) SaveGameResults(room *Room, ranking []LeaderboardPlayerData) {
	if h.gameRepo == nil || h.userRepo == nil {
		return
	}

	ctx := context.Background()

	// Collect player info
	players := make([]*PlayerState, 0, len(room.Players))
	for _, ps := range room.Players {
		players = append(players, ps)
	}

	// Create game session (session-level data only)
	session := &model.GameSession{
		Mode:     room.Mode,
		Language: room.Language,
		Rounds:   room.TotalRounds,
	}

	if err := h.gameRepo.Create(ctx, session); err != nil {
		log.Printf("Failed to create game session: %v", err)
		return
	}

	// Determine winner
	var winnerID *uuid.UUID
	if len(ranking) > 0 && ranking[0].Score > 0 {
		for _, ps := range players {
			if ps.Client.Username == ranking[0].Username {
				id := ps.Client.ID
				winnerID = &id
				break
			}
		}
	}

	// Calculate avg reaction across all players
	totalReaction := 0
	totalCount := 0
	for _, ps := range players {
		for _, r := range ps.Reactions {
			totalReaction += r
			totalCount++
		}
	}
	avgReaction := 0
	if totalCount > 0 {
		avgReaction = totalReaction / totalCount
	}

	if err := h.gameRepo.Finish(ctx, session.ID, avgReaction, winnerID); err != nil {
		log.Printf("Failed to finish game session: %v", err)
	}

	// Build rank map from ranking
	rankMap := make(map[string]int)
	for _, r := range ranking {
		rankMap[r.Username] = r.Rank
	}

	// Insert per-player results into game_session_players
	for _, ps := range players {
		bestReaction := 0
		playerAvgReaction := 0
		if len(ps.Reactions) > 0 {
			bestReaction = ps.Reactions[0]
			sum := 0
			for _, r := range ps.Reactions {
				sum += r
				if r < bestReaction {
					bestReaction = r
				}
			}
			playerAvgReaction = sum / len(ps.Reactions)
		}

		playerResult := &model.GameSessionPlayer{
			SessionID:      session.ID,
			UserID:         ps.Client.ID,
			Score:          ps.Score,
			AvgReactionMs:  playerAvgReaction,
			BestReactionMs: bestReaction,
			Rank:           rankMap[ps.Client.Username],
		}
		if err := h.gameRepo.CreatePlayerResult(ctx, playerResult); err != nil {
			log.Printf("Failed to save player result for %s: %v", ps.Client.Username, err)
		}

		// Update user stats
		if err := h.userRepo.UpdateStats(ctx, ps.Client.ID, int64(ps.Score), bestReaction); err != nil {
			log.Printf("Failed to update stats for %s: %v", ps.Client.Username, err)
		}
	}

	log.Printf("Game %s results saved. Players: %d, Winner: %v", room.ID, len(players), winnerID)
}
