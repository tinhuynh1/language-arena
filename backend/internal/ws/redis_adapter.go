package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisAdapter handles cross-instance communication via Redis Pub/Sub.
// Each backend instance gets a unique NodeID. Rooms are registered in a
// Redis hash so any instance can look up which node owns a given room code.
type RedisAdapter struct {
	client *redis.Client
	NodeID string
	hub    *Hub
}

// Inter-node message types
const (
	RedisProxyJoin   = "proxy_join"
	RedisProxyAction = "proxy_action"
	RedisProxyLeft   = "proxy_left"
	RedisRelayWS     = "relay_ws"
	RedisMatchFound  = "match_found"
)

// RedisMessage is the envelope for all inter-node communication.
type RedisMessage struct {
	Type       string          `json:"type"`
	FromNode   string          `json:"from_node"`
	RoomCode   string          `json:"room_code,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	Username   string          `json:"username,omitempty"`
	WSMessage  json.RawMessage `json:"ws_message,omitempty"`
	ActionType string          `json:"action_type,omitempty"`
	ActionData json.RawMessage `json:"action_data,omitempty"`
}

func NewRedisAdapter(redisURL string) (*RedisAdapter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	nodeID := uuid.New().String()[:8]

	slog.Info("redis connected", "component", "BOOT", "node_id", nodeID)

	return &RedisAdapter{
		client: client,
		NodeID: nodeID,
	}, nil
}

func (ra *RedisAdapter) SetHub(hub *Hub) {
	ra.hub = hub
}

// --- Room Registry ---

const roomRegistryKey = "lingo:room_registry"

func (ra *RedisAdapter) RegisterRoom(code, nodeID string) {
	ctx := context.Background()
	ra.client.HSet(ctx, roomRegistryKey, code, nodeID)
	ra.client.Expire(ctx, roomRegistryKey, 2*time.Hour)
	slog.Info("room registered", "component", "REDIS", "room_code", code, "node_id", nodeID)
}

func (ra *RedisAdapter) UnregisterRoom(code string) {
	ctx := context.Background()
	ra.client.HDel(ctx, roomRegistryKey, code)
	slog.Debug("room unregistered", "component", "REDIS", "room_code", code)
}

func (ra *RedisAdapter) LookupRoom(code string) (string, bool) {
	ctx := context.Background()
	nodeID, err := ra.client.HGet(ctx, roomRegistryKey, code).Result()
	if err != nil {
		return "", false
	}
	return nodeID, true
}

// --- Reconnect Registry (TTL-based) ---

const reconnectGracePeriod = 30 * time.Second

type ReconnectInfo struct {
	RoomCode string `json:"room_code"`
	NodeID   string `json:"node_id"`
}

func reconnectKey(userID string) string {
	return "lingo:reconnect:" + userID
}

func (ra *RedisAdapter) SetReconnectInfo(userID, roomCode string) {
	ctx := context.Background()
	info := ReconnectInfo{RoomCode: roomCode, NodeID: ra.NodeID}
	data, err := json.Marshal(info)
	if err != nil {
		slog.Error("reconnect marshal error", "component", "REDIS", "err", err)
		return
	}
	ra.client.Set(ctx, reconnectKey(userID), data, reconnectGracePeriod)
	slog.Info("reconnect info saved", "component", "REDIS", "user_id", userID, "room_code", roomCode, "ttl_s", 30)
}

func (ra *RedisAdapter) GetReconnectInfo(userID string) (*ReconnectInfo, bool) {
	ctx := context.Background()
	data, err := ra.client.Get(ctx, reconnectKey(userID)).Result()
	if err != nil {
		return nil, false
	}
	var info ReconnectInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, false
	}
	return &info, true
}

func (ra *RedisAdapter) ClearReconnectInfo(userID string) {
	ctx := context.Background()
	ra.client.Del(ctx, reconnectKey(userID))
}

// --- Pub/Sub ---

func nodeChannel(nodeID string) string {
	return "lingo:node:" + nodeID
}

func (ra *RedisAdapter) PublishToNode(targetNodeID string, msg RedisMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("redis marshal error", "component", "REDIS", "err", err)
		return
	}
	ctx := context.Background()
	ra.client.Publish(ctx, nodeChannel(targetNodeID), data)
}

func (ra *RedisAdapter) Subscribe() {
	ctx := context.Background()
	channel := nodeChannel(ra.NodeID)
	sub := ra.client.Subscribe(ctx, channel)
	ch := sub.Channel()

	slog.Info("subscribing to node channel", "component", "REDIS", "channel", channel)

	go func() {
		for msg := range ch {
			var rmsg RedisMessage
			if err := json.Unmarshal([]byte(msg.Payload), &rmsg); err != nil {
				slog.Error("redis unmarshal error", "component", "REDIS", "err", err)
				continue
			}
			ra.hub.HandleRedisMessage(rmsg)
		}
	}()
}

func (ra *RedisAdapter) Close() {
	ra.client.Close()
}

// --- Matchmaking Queue ---

const (
	matchQueuePrefix = "lingo:queue:"
	matchQueueTTL    = 60 // seconds
)

// RedisQueueEntry is stored in the per-bucket Redis hash while a player is waiting.
type RedisQueueEntry struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	NodeID    string `json:"node_id"`
	Timestamp int64  `json:"timestamp"`
}

// enqueueOrMatchScript atomically finds an opponent or adds the caller to the queue.
// Returns the opponent's JSON entry if matched, or false if now waiting.
var enqueueOrMatchScript = redis.NewScript(`
local key = KEYS[1]
local myUserID = ARGV[1]
local myEntry = ARGV[2]
local now = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

local entries = redis.call('HGETALL', key)
for i = 1, #entries, 2 do
    local uid = entries[i]
    local entryJson = entries[i+1]
    if uid ~= myUserID then
        local ok, entry = pcall(cjson.decode, entryJson)
        if ok and entry and (now - entry.timestamp) <= ttl then
            redis.call('HDEL', key, uid)
            return entryJson
        else
            redis.call('HDEL', key, uid)
        end
    end
end

redis.call('HSET', key, myUserID, myEntry)
redis.call('EXPIRE', key, ttl * 2)
return false
`)

func matchQueueKey(language, level, quizType, mode string) string {
	return matchQueuePrefix + language + ":" + level + ":" + quizType + ":" + mode
}

// EnqueueOrMatch atomically either matches the caller with a waiting player or
// adds them to the queue. Returns the opponent entry on match, nil when waiting.
func (ra *RedisAdapter) EnqueueOrMatch(userID, username, language, level, quizType, mode string) (*RedisQueueEntry, error) {
	entry := RedisQueueEntry{
		UserID:    userID,
		Username:  username,
		NodeID:    ra.NodeID,
		Timestamp: time.Now().Unix(),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	key := matchQueueKey(language, level, quizType, mode)
	result, err := enqueueOrMatchScript.Run(ctx, ra.client, []string{key},
		userID, string(data), time.Now().Unix(), matchQueueTTL,
	).Text()

	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var opponent RedisQueueEntry
	if err := json.Unmarshal([]byte(result), &opponent); err != nil {
		return nil, err
	}
	return &opponent, nil
}

// RemoveFromQueue removes a player from the Redis matchmaking queue bucket.
func (ra *RedisAdapter) RemoveFromQueue(userID, language, level, quizType, mode string) {
	ctx := context.Background()
	ra.client.HDel(ctx, matchQueueKey(language, level, quizType, mode), userID)
}
