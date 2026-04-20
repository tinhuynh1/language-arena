package ws

import "github.com/google/uuid"

type MessageType string

const (
	// Client → Server
	MsgJoinQueue   MessageType = "join_queue"
	MsgCreateRoom  MessageType = "create_room"
	MsgJoinRoom    MessageType = "join_room"
	MsgReady       MessageType = "ready"
	MsgStartGame   MessageType = "start_game"
	MsgTargetHit   MessageType = "target_hit"
	MsgLeaveRoom   MessageType = "leave_room"

	// Server → Client
	MsgQueueJoined     MessageType = "queue_joined"
	MsgMatchFound      MessageType = "match_found"
	MsgRoomCreated     MessageType = "room_created"
	MsgPlayerJoined    MessageType = "player_joined"
	MsgPlayerLeft      MessageType = "player_left"
	MsgCountdown       MessageType = "countdown"
	MsgRoundStart      MessageType = "round_start"
	MsgScoreUpdate     MessageType = "score_update"
	MsgLiveLeaderboard MessageType = "live_leaderboard"
	MsgRoundEnd        MessageType = "round_end"
	MsgGameOver        MessageType = "game_over"
	MsgOpponentLeft    MessageType = "opponent_left"
	MsgHostChanged     MessageType = "host_changed"
	MsgGameStateSync   MessageType = "game_state_sync"
	MsgError           MessageType = "error"
)

type RoundEndData struct {
	Result   string `json:"result"`
	NextInMs int    `json:"next_in_ms"`
}

type GameStateSyncData struct {
	RoomCode      string   `json:"room_code"`
	Mode          string   `json:"mode"`
	State         string   `json:"state"`
	Round         int      `json:"round"`
	TotalRounds   int      `json:"total_rounds"`
	Question      string   `json:"question"`
	Targets       []Target `json:"targets"`
	TimeMs        int      `json:"time_ms"`
	ElapsedMs     int      `json:"elapsed_ms"`
	YourScore     int      `json:"your_score"`
	OpponentScore int      `json:"opponent_score"`
	Players       []string `json:"players"`
}

type WSMessage struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type JoinQueueData struct {
	Language string `json:"language"`
	Level    string `json:"level"`
	Mode     string `json:"mode"`
	QuizType string `json:"quiz_type"`
}

type CreateRoomData struct {
	Language string `json:"language"`
	Level    string `json:"level"`
	QuizType string `json:"quiz_type"`
}

type JoinRoomData struct {
	RoomCode string `json:"room_code"`
}

type TargetHitData struct {
	TargetID   string `json:"target_id"`
	ReactionMs int    `json:"reaction_ms"`
}

type MatchFoundData struct {
	RoomID      string `json:"room_id"`
	Opponent    string `json:"opponent"`
	PlayerCount int    `json:"player_count"`
	Mode        string `json:"mode"`
	IsHost      bool   `json:"is_host"`
	Host        string `json:"host"`
}

type RoomCreatedData struct {
	RoomCode    string `json:"room_code"`
	RoomID      string `json:"room_id"`
	Language    string `json:"language"`
	Level       string `json:"level"`
}

type PlayerJoinedData struct {
	Username    string   `json:"username"`
	PlayerCount int      `json:"player_count"`
	Players     []string `json:"players"`
	Host        string   `json:"host"`
}

type HostChangedData struct {
	NewHost string `json:"new_host"`
}

type Target struct {
	ID      string  `json:"id"`
	Word    string  `json:"word"`
	Meaning string  `json:"meaning"`
	Label   string  `json:"label"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Correct bool    `json:"correct"`
}

type RoundStartData struct {
	Round    int      `json:"round"`
	Total    int      `json:"total"`
	Question string   `json:"question"`
	Targets  []Target `json:"targets"`
	TimeMs   int      `json:"time_ms"`
}

type ScoreUpdateData struct {
	You        int    `json:"you"`
	Opponent   int    `json:"opponent"`
	LastHitBy  string `json:"last_hit_by,omitempty"`
	ReactionMs int    `json:"reaction_ms,omitempty"`
}

type LeaderboardPlayerData struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Score    int    `json:"score"`
}

type LiveLeaderboardData struct {
	Round   int                     `json:"round"`
	Players []LeaderboardPlayerData `json:"players"`
}

type GameOverData struct {
	Winner        string          `json:"winner"`
	YourScore     int             `json:"your_score"`
	OpponentScore int             `json:"opponent_score"`
	Stats         GameOverStats   `json:"stats"`
	Ranking       []LeaderboardPlayerData `json:"ranking,omitempty"`
}

type GameOverStats struct {
	TotalRounds   int `json:"total_rounds"`
	AvgReactionMs int `json:"avg_reaction_ms"`
	Accuracy      int `json:"accuracy"`
}

type PlayerInfo struct {
	UserID   uuid.UUID
	Username string
}
