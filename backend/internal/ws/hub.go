package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/repository"
	"github.com/michael/language-arena/backend/internal/service"
)

const maxDisconnectedPlayers = 500

type DisconnectedPlayer struct {
	Client    *Client
	Room      *Room
	Timer     *time.Timer
	UserID    uuid.UUID
	RoomCode  string
}

type Hub struct {
	log        *slog.Logger
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Rooms      map[string]*Room
	RoomByCode map[string]*Room
	Matchmaker *Matchmaker

	vocabService *service.VocabService
	gameRepo     *repository.GameRepository
	userRepo     *repository.UserRepository

	Redis        *RedisAdapter
	proxyClients map[string]*Client // userID → proxy client on this node

	disconnectedPlayers map[uuid.UUID]*DisconnectedPlayer

	mu sync.RWMutex
}

func NewHub(vocabService *service.VocabService, gameRepo *repository.GameRepository, userRepo *repository.UserRepository, redisAdapter *RedisAdapter) *Hub {
	h := &Hub{
		log:                 slog.Default().With("component", "WS"),
		Clients:             make(map[*Client]bool),
		Register:            make(chan *Client),
		Unregister:          make(chan *Client),
		Rooms:               make(map[string]*Room),
		RoomByCode:          make(map[string]*Room),
		vocabService:        vocabService,
		gameRepo:            gameRepo,
		userRepo:            userRepo,
		Redis:               redisAdapter,
		proxyClients:        make(map[string]*Client),
		disconnectedPlayers: make(map[uuid.UUID]*DisconnectedPlayer),
	}
	h.Matchmaker = NewMatchmaker(h)
	if redisAdapter != nil {
		redisAdapter.SetHub(h)
		redisAdapter.Subscribe()
	}
	return h
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()

			// Check for reconnection (same user_id returning)
			if h.tryReconnect(client) {
				continue
			}

			h.log.Info("client connected", "player", client.Username, "user_id", client.ID, "total", len(h.Clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

			h.Matchmaker.Remove(client)

			// Grace period: if player was in a room during an active game, don't remove yet
			if room := client.GetRoom(); room != nil {
				room.mu.Lock()
				gameActive := room.State == StatePlaying || room.State == StateCountdown || room.State == StateRoundEnd
				roomCode := room.Code
				room.mu.Unlock()

				if gameActive {
					h.startGracePeriod(client, room, roomCode)
				} else {
					room.RemovePlayer(client)
				}
			}

			// Clean up any cross-instance proxy mappings for this client
			if h.Redis != nil {
				h.mu.Lock()
				for key, c := range h.proxyClients {
					if c == client {
						parts := splitProxyKey(key)
						if len(parts) == 2 {
							if ownerNode, found := h.Redis.LookupRoom(parts[1]); found {
								h.Redis.PublishToNode(ownerNode, RedisMessage{
									Type:     RedisProxyLeft,
									FromNode: h.Redis.NodeID,
									RoomCode: parts[1],
									UserID:   client.ID.String(),
									Username: client.Username,
								})
							}
						}
						delete(h.proxyClients, key)
					}
				}
				h.mu.Unlock()
			}


			h.log.Info("client disconnected", "player", client.Username, "user_id", client.ID, "total", len(h.Clients))
		}
	}
}

func (h *Hub) startGracePeriod(client *Client, room *Room, roomCode string) {
	h.mu.Lock()
	// Enforce max cap
	if len(h.disconnectedPlayers) >= maxDisconnectedPlayers {
		h.mu.Unlock()
		room.RemovePlayer(client)
		h.log.Warn("grace period rejected: max disconnected players reached", "player", client.Username)
		return
	}

	dp := &DisconnectedPlayer{
		Client:   client,
		Room:     room,
		UserID:   client.ID,
		RoomCode: roomCode,
	}

	dp.Timer = time.AfterFunc(reconnectGracePeriod, func() {
		h.mu.Lock()
		if existing, ok := h.disconnectedPlayers[client.ID]; ok && existing == dp {
			delete(h.disconnectedPlayers, client.ID)
		}
		h.mu.Unlock()
		room.RemovePlayer(client)
		h.log.Info("grace period expired, player removed", "player", client.Username, "room_code", roomCode)
	})

	h.disconnectedPlayers[client.ID] = dp
	h.mu.Unlock()

	// Save routing info to Redis for cross-pod reconnect
	if h.Redis != nil {
		h.Redis.SetReconnectInfo(client.ID.String(), roomCode)
	}

	h.log.Info("grace period started", "player", client.Username, "room_code", roomCode, "ttl_s", 30)
}

func (h *Hub) tryReconnect(newClient *Client) bool {
	h.mu.Lock()
	dp, ok := h.disconnectedPlayers[newClient.ID]
	if ok {
		// Same pod reconnect: swap client directly
		dp.Timer.Stop()
		delete(h.disconnectedPlayers, newClient.ID)
		h.mu.Unlock()

		if dp.Room.ReconnectPlayer(dp.Client, newClient) {
			// Clear Redis reconnect info
			if h.Redis != nil {
				h.Redis.ClearReconnectInfo(newClient.ID.String())
			}

			// Send current game state to the reconnected client
			stateSync := dp.Room.GetGameStateSync(newClient)
			newClient.SendMessage(WSMessage{Type: MsgGameStateSync, Data: stateSync})

			h.log.Info("player reconnected (same pod)",
				"player", newClient.Username,
				"room_code", dp.RoomCode,
			)
			return true
		}
		h.log.Warn("reconnect failed: room rejected player", "player", newClient.Username)
		return false
	}
	h.mu.Unlock()

	// Cross-pod reconnect: check Redis
	if h.Redis != nil {
		info, found := h.Redis.GetReconnectInfo(newClient.ID.String())
		if found && info.NodeID != h.Redis.NodeID {
			// Room is on another pod — use existing proxy join mechanism
			h.Redis.PublishToNode(info.NodeID, RedisMessage{
				Type:     RedisProxyJoin,
				FromNode: h.Redis.NodeID,
				RoomCode: info.RoomCode,
				UserID:   newClient.ID.String(),
				Username: newClient.Username,
			})
			h.Redis.ClearReconnectInfo(newClient.ID.String())

			h.mu.Lock()
			proxyKey := newClient.ID.String() + ":" + info.RoomCode
			h.proxyClients[proxyKey] = newClient
			h.mu.Unlock()

			h.log.Info("player reconnecting (cross-pod)",
				"player", newClient.Username,
				"room_code", info.RoomCode,
				"target_node", info.NodeID,
			)
			return true
		}
	}

	return false
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

	h.Matchmaker.Enqueue(client, data.Language, data.Level, model.QuizType(data.QuizType), model.GameMode(data.Mode))
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

	// Register in Redis so other instances can find this room
	if h.Redis != nil {
		h.Redis.RegisterRoom(room.Code, h.Redis.NodeID)
	}

	h.log.Info("room created", "room_id", room.ID, "room_code", room.Code, "host", client.Username, "language", data.Language, "level", data.Level)
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

	// Try local first
	h.mu.RLock()
	room, ok := h.RoomByCode[data.RoomCode]
	h.mu.RUnlock()

	if ok {
		// Room is on this instance — join directly
		if !room.AddPlayer(client) {
			// Distinguish why join failed
			errMsg := "room is full"
			if room.State != StateWaiting {
				errMsg = "game already started"
			}
			client.SendMessage(WSMessage{Type: MsgError, Data: errMsg})
			return
		}

		client.SendMessage(WSMessage{
			Type: MsgMatchFound,
			Data: MatchFoundData{
				RoomID:      room.ID,
				PlayerCount: len(room.Players),
				Mode:        string(room.Mode),
				IsHost:      client.ID == room.HostID,
				Host:        room.GetHostUsername(),
			},
		})

		names := room.getPlayerNames()
		room.broadcast(WSMessage{
			Type: MsgPlayerJoined,
			Data: PlayerJoinedData{
				Username:    client.Username,
				PlayerCount: len(room.Players),
				Players:     names,
				Host:        room.GetHostUsername(),
			},
		})

		h.log.Info("player joined room", "player", client.Username, "room_id", room.ID, "room_code", room.Code, "player_count", len(room.Players))
		return
	}

	// Room not local — check Redis for cross-instance
	if h.Redis != nil {
		if ownerNode, found := h.Redis.LookupRoom(data.RoomCode); found {
			if ownerNode == h.Redis.NodeID {
				// Stale registry entry
				client.SendMessage(WSMessage{Type: MsgError, Data: "room not found"})
				return
			}

			// Cross-instance join: register this client as a remote player
			// and send proxy_join to the owning node
			h.mu.Lock()
			h.proxyClients[client.ID.String()+":"+data.RoomCode] = client
			h.mu.Unlock()

			h.Redis.PublishToNode(ownerNode, RedisMessage{
				Type:     RedisProxyJoin,
				FromNode: h.Redis.NodeID,
				RoomCode: data.RoomCode,
				UserID:   client.ID.String(),
				Username: client.Username,
			})

			h.log.Info("cross-instance join requested", "player", client.Username, "room_code", data.RoomCode, "target_node", ownerNode)
			return
		}
	}

	client.SendMessage(WSMessage{Type: MsgError, Data: "room not found"})
}

func (h *Hub) handleStartGame(client *Client) {
	room := client.GetRoom()
	if room != nil {
		room.StartByHost(client)
		return
	}
	// Cross-instance: forward to room owner
	h.forwardProxyAction(client, "start_game", nil)
}

func (h *Hub) handleReady(client *Client) {
	room := client.GetRoom()
	if room != nil {
		room.SetReady(client)
		return
	}
	h.forwardProxyAction(client, "ready", nil)
}

func (h *Hub) handleTargetHit(client *Client, msg WSMessage) {
	room := client.GetRoom()
	if room != nil {
		dataBytes, err := json.Marshal(msg.Data)
		if err != nil {
			return
		}
		var data TargetHitData
		if err := json.Unmarshal(dataBytes, &data); err != nil {
			return
		}
		room.HandleHit(client, data)
		return
	}
	// Cross-instance: forward the action data
	rawData, _ := json.Marshal(msg.Data)
	h.forwardProxyAction(client, "target_hit", rawData)
}

func (h *Hub) handleLeaveRoom(client *Client) {
	room := client.GetRoom()
	if room != nil {
		room.RemovePlayer(client)
		client.SetRoom(nil)
	}
	h.Matchmaker.Remove(client)
	h.forwardProxyAction(client, "leave", nil)
}

// forwardProxyAction sends an action via Redis to the room-owning instance.
func (h *Hub) forwardProxyAction(client *Client, actionType string, actionData json.RawMessage) {
	if h.Redis == nil {
		return
	}

	userID := client.ID.String()

	h.mu.RLock()
	var roomCode string
	for key, c := range h.proxyClients {
		if c == client {
			parts := splitProxyKey(key)
			if len(parts) == 2 {
				roomCode = parts[1]
			}
			break
		}
	}
	h.mu.RUnlock()

	if roomCode == "" {
		return
	}

	ownerNode, found := h.Redis.LookupRoom(roomCode)
	if !found {
		return
	}

	h.Redis.PublishToNode(ownerNode, RedisMessage{
		Type:       RedisProxyAction,
		FromNode:   h.Redis.NodeID,
		RoomCode:   roomCode,
		UserID:     userID,
		Username:   client.Username,
		ActionType: actionType,
		ActionData: actionData,
	})

	// If leaving, clean up local proxy mapping
	if actionType == "leave" {
		h.mu.Lock()
		delete(h.proxyClients, userID+":"+roomCode)
		h.mu.Unlock()
	}
}


func splitProxyKey(key string) []string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == ':' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{key}
}

func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return data
}

// HandleRedisMessage processes messages from other instances via Redis Pub/Sub.
func (h *Hub) HandleRedisMessage(msg RedisMessage) {
	switch msg.Type {
	case RedisProxyJoin:
		h.handleProxyJoin(msg)
	case RedisProxyAction:
		h.handleProxyAction(msg)
	case RedisRelayWS:
		h.handleRelayWS(msg)
	case RedisProxyLeft:
		h.handleProxyLeft(msg)
	default:
		h.log.Warn("unknown redis message type", "type", msg.Type, "from_node", msg.FromNode)
	}
}

// handleProxyJoin: a player on another instance wants to join a room on THIS instance.
// We create a ProxyClient whose SendMessage publishes back to the joiner's node.
func (h *Hub) handleProxyJoin(msg RedisMessage) {
	h.mu.RLock()
	room, ok := h.RoomByCode[msg.RoomCode]
	h.mu.RUnlock()

	if !ok {
		h.log.Warn("proxy_join: room not found", "room_code", msg.RoomCode)
		return
	}

	userID, err := uuid.Parse(msg.UserID)
	if err != nil {
		h.log.Error("proxy_join: invalid user_id", "user_id", msg.UserID)
		return
	}

	fromNode := msg.FromNode
	proxyUserID := msg.UserID
	roomCode := msg.RoomCode

	// Create a ProxyClient — its SendMessage relays via Redis to the joiner's node
	proxy := &Client{
		ID:       userID,
		Username: msg.Username,
		Hub:      h,
		Send:     make(chan []byte, 256),
		IsProxy:  true,
	}
	proxy.RelayFunc = func(wsMsg WSMessage) {
		wsData, err := json.Marshal(wsMsg)
		if err != nil {
			return
		}
		h.Redis.PublishToNode(fromNode, RedisMessage{
			Type:      RedisRelayWS,
			FromNode:  h.Redis.NodeID,
			RoomCode:  roomCode,
			UserID:    proxyUserID,
			WSMessage: wsData,
		})
	}

	if !room.AddPlayer(proxy) {
		// Relay error back to the joiner's node
		errMsg := "room is full"
		if room.State != StateWaiting {
			errMsg = "game already started"
		}
		h.Redis.PublishToNode(fromNode, RedisMessage{
			Type:      RedisRelayWS,
			FromNode:  h.Redis.NodeID,
			RoomCode:  roomCode,
			UserID:    proxyUserID,
			WSMessage: mustMarshal(WSMessage{Type: MsgError, Data: errMsg}),
		})
		h.log.Warn("proxy_join: room full or started", "room_code", msg.RoomCode)
		return
	}

	// Track proxy for action forwarding
	h.mu.Lock()
	h.proxyClients[proxyUserID+":"+roomCode] = proxy
	h.mu.Unlock()

	// Send match_found via relay
	proxy.SendMessage(WSMessage{
		Type: MsgMatchFound,
		Data: MatchFoundData{
			RoomID:      room.ID,
			PlayerCount: len(room.Players),
			Mode:        string(room.Mode),
			IsHost:      proxy.ID == room.HostID,
			Host:        room.GetHostUsername(),
		},
	})

	// Broadcast player_joined to all (including proxy)
	names := room.getPlayerNames()
	room.broadcast(WSMessage{
		Type: MsgPlayerJoined,
		Data: PlayerJoinedData{
			Username:    msg.Username,
			PlayerCount: len(room.Players),
			Players:     names,
			Host:        room.GetHostUsername(),
		},
	})

	h.log.Info("proxy client added", "player", msg.Username, "room_code", msg.RoomCode, "from_node", msg.FromNode)
}

// handleProxyAction: a player on another instance performed an action (hit, ready, start, leave)
func (h *Hub) handleProxyAction(msg RedisMessage) {
	key := msg.UserID + ":" + msg.RoomCode

	h.mu.RLock()
	proxy, ok := h.proxyClients[key]
	h.mu.RUnlock()

	if !ok {
		h.log.Warn("proxy_action: no proxy found", "user_id", msg.UserID, "room_code", msg.RoomCode)
		return
	}

	room := proxy.GetRoom()
	if room == nil {
		return
	}

	switch msg.ActionType {
	case "target_hit":
		var data TargetHitData
		if err := json.Unmarshal(msg.ActionData, &data); err == nil {
			room.HandleHit(proxy, data)
		}
	case "ready":
		room.SetReady(proxy)
	case "start_game":
		room.StartByHost(proxy)
	case "leave":
		room.RemovePlayer(proxy)
		h.mu.Lock()
		delete(h.proxyClients, key)
		h.mu.Unlock()
	}
}

// handleRelayWS: the room-owning instance sent a WS message for a player on THIS instance.
// Find the real local client and forward the message.
func (h *Hub) handleRelayWS(msg RedisMessage) {
	key := msg.UserID + ":" + msg.RoomCode

	h.mu.RLock()
	realClient, ok := h.proxyClients[key]
	h.mu.RUnlock()

	if !ok {
		return
	}

	var wsMsg WSMessage
	if err := json.Unmarshal(msg.WSMessage, &wsMsg); err != nil {
		h.log.Error("relay_ws: unmarshal error", "err", err)
		return
	}

	// Send directly via the real WebSocket (bypass RelayFunc)
	data, err := json.Marshal(wsMsg)
	if err != nil {
		return
	}
	select {
	case realClient.Send <- data:
	default:
		h.log.Warn("relay_ws: send buffer full", "user_id", msg.UserID)
	}
}

// handleProxyLeft: a player on another instance disconnected
func (h *Hub) handleProxyLeft(msg RedisMessage) {
	key := msg.UserID + ":" + msg.RoomCode

	h.mu.RLock()
	proxy, ok := h.proxyClients[key]
	h.mu.RUnlock()

	if !ok {
		return
	}

	room := proxy.GetRoom()
	if room != nil {
		room.RemovePlayer(proxy)
	}

	h.mu.Lock()
	delete(h.proxyClients, key)
	h.mu.Unlock()

	h.log.Info("proxy client removed", "player", msg.Username, "room_code", msg.RoomCode)
}

func (h *Hub) RemoveRoom(room *Room) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.Rooms, room.ID)
	delete(h.RoomByCode, room.Code)

	if h.Redis != nil {
		h.Redis.UnregisterRoom(room.Code)
	}
}

func (h *Hub) GetVocabs(language, level string, count int) []model.Vocabulary {
	vocabs, err := h.vocabService.GetRandomSet(context.Background(), language, level, count)
	if err != nil || len(vocabs) == 0 {
		h.log.Error("failed to get vocabs, using fallback", "err", err, "language", language, "level", level)
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
		h.log.Error("failed to create game session", "err", err, "room_id", room.ID)
		return
	}

	// Determine winner
	var winnerID *uuid.UUID
	if len(ranking) > 0 && ranking[0].CorrectCount > 0 {
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
	sessionAvg := 0
	if totalCount > 0 {
		sessionAvg = totalReaction / totalCount
	}

	if err := h.gameRepo.Finish(ctx, session.ID, sessionAvg, winnerID); err != nil {
		h.log.Error("failed to finish game session", "err", err, "session_id", session.ID)
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

		// Avg including penalties (for AllReactions)
		allAvg := avgReaction(ps.AllReactions)

		playerResult := &model.GameSessionPlayer{
			SessionID:      session.ID,
			UserID:         ps.Client.ID,
			Score:          ps.CorrectCount,
			AvgReactionMs:  playerAvgReaction,
			BestReactionMs: bestReaction,
			Rank:           rankMap[ps.Client.Username],
		}
		if err := h.gameRepo.CreatePlayerResult(ctx, playerResult); err != nil {
				h.log.Error("failed to save player result", "err", err, "player", ps.Client.Username)
		}

		// Update user stats with avg reaction (including penalties) and correct count
		if err := h.userRepo.UpdateStats(ctx, ps.Client.ID, allAvg, ps.CorrectCount, bestReaction); err != nil {
				h.log.Error("failed to update user stats", "err", err, "player", ps.Client.Username)
		}
	}

	h.log.Info("game results saved", "room_id", room.ID, "players", len(players), "winner_id", winnerID)
}
