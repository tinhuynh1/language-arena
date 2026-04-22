package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRoom_Initialization(t *testing.T) {
	vocabs := []model.Vocabulary{
		{ID: uuid.New(), Word: "apple", Meaning: "quả táo", Language: "en", Level: "A1"},
	}

	room := NewRoom("en", "A1", model.ModeBattle, model.QuizTypeMeaningToWord, vocabs, nil)
	
	assert.NotNil(t, room)
	assert.Equal(t, "en", room.Language)
	assert.Equal(t, model.ModeBattle, room.Mode)
	assert.Equal(t, StateWaiting, room.State)
	assert.Len(t, room.Vocabs, 1)
	assert.NotEqual(t, "", room.Code)
}

func TestRoom_AddPlayer(t *testing.T) {
	vocabs := []model.Vocabulary{{ID: uuid.New(), Word: "test"}}
	room := NewRoom("en", "A1", model.ModeDuel, model.QuizTypeMeaningToWord, vocabs, nil)
	c1 := &Client{ID: uuid.New(), Username: "p1"}
	c2 := &Client{ID: uuid.New(), Username: "p2"}
	c3 := &Client{ID: uuid.New(), Username: "p3"}

	// Success
	assert.True(t, room.AddPlayer(c1))
	assert.True(t, room.AddPlayer(c2))

	// Fail because duel max is 2
	assert.False(t, room.AddPlayer(c3))

	// Remove player
	room.RemovePlayer(c1)
	assert.Len(t, room.Players, 1)
}

func TestRoom_SetReady(t *testing.T) {
	vocabs := []model.Vocabulary{{ID: uuid.New(), Word: "test"}}
	room := NewRoom("en", "A1", model.ModeDuel, model.QuizTypeMeaningToWord, vocabs, nil)
	c1 := &Client{ID: uuid.New(), Username: "p1"}
	c2 := &Client{ID: uuid.New(), Username: "p2"}

	room.AddPlayer(c1)
	room.AddPlayer(c2)

	assert.False(t, room.Players[c1].Ready)
	
	room.SetReady(c1)
	assert.True(t, room.Players[c1].Ready)

	// Since Mode is Duel, setting both ready should automatically trigger StartGame
	// Let's test the state changes to StateCountdown
	room.SetReady(c2)
	
	// Because startGame runs in a separate goroutine, we sleep briefly
	time.Sleep(10 * time.Millisecond)
	
	room.mu.Lock()
	state := room.State
	room.mu.Unlock()
	
	assert.Equal(t, StateCountdown, state)
}

func TestRoom_StartByHost(t *testing.T) {
	vocabs := []model.Vocabulary{{ID: uuid.New(), Word: "test"}}
	room := NewRoom("en", "A1", model.ModeBattle, model.QuizTypeMeaningToWord, vocabs, nil)
	host := &Client{ID: uuid.New(), Username: "host", Send: make(chan []byte, 10)}
	guest := &Client{ID: uuid.New(), Username: "guest"}
	
	room.HostID = host.ID
	room.AddPlayer(host)
	room.AddPlayer(guest)

	room.StartByHost(guest) // Guest tries to start
	
	room.mu.Lock()
	assert.Equal(t, StateWaiting, room.State) // Should fail
	room.mu.Unlock()

	room.StartByHost(host) // Host starts

	time.Sleep(10 * time.Millisecond)

	room.mu.Lock()
	assert.Equal(t, StatePlaying, room.State)
	assert.True(t, room.Players[guest].Ready)
	room.mu.Unlock()
}

func TestRoom_HandleHit(t *testing.T) {
	vocabID := uuid.New()
	vocabs := []model.Vocabulary{
		{ID: vocabID, Word: "apple", Meaning: "quả táo", Language: "en", Level: "A1"},
		{ID: uuid.New(), Word: "cat", Meaning: "mèo", Language: "en", Level: "A1"},
	}

	room := NewRoom("en", "A1", model.ModeSolo, model.QuizTypeMeaningToWord, vocabs, nil)
	c := &Client{ID: uuid.New(), Username: "solo_player", Send: make(chan []byte, 50)}
	room.AddPlayer(c)

	// Force into playing state
	room.State = StatePlaying
	room.CurrentRound = 1
	room.RoundStartAt = time.Now()

	// Correct hit
	hitData := TargetHitData{
		TargetID:   vocabID.String()[:8],
		ReactionMs: 1200,
	}

	room.HandleHit(c, hitData)
	
	room.mu.Lock()
	ps := room.Players[c]
	assert.True(t, ps.Answered)
	assert.Equal(t, 1, ps.CorrectCount)
	assert.Len(t, ps.Reactions, 1)
	room.mu.Unlock()
}

func TestRoom_GetGameStateSync(t *testing.T) {
	room := NewRoom("en", "A1", model.ModeBattle, model.QuizTypeMeaningToWord, []model.Vocabulary{{ID: uuid.New(), Word: "apple", Meaning: "táo"}}, nil)
	c := &Client{ID: uuid.New(), Username: "sync_player", Send: make(chan []byte, 10)}
	room.AddPlayer(c)
	
	room.State = StatePlaying
	room.CurrentRound = 1
	room.TotalRounds = 5

	syncData := room.GetGameStateSync(c)
	assert.Equal(t, "playing", syncData.State)
	assert.Equal(t, 1, syncData.Round)
	assert.Equal(t, 5, syncData.TotalRounds)
	assert.Len(t, syncData.Players, 1)
}

func TestRoom_FinishGame(t *testing.T) {
	vocabs := []model.Vocabulary{{ID: uuid.New(), Word: "test"}}
	room := NewRoom("en", "A1", model.ModeDuel, model.QuizTypeMeaningToWord, vocabs, nil)
	c1 := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10)}
	c2 := &Client{ID: uuid.New(), Username: "p2", Send: make(chan []byte, 10)}
	room.AddPlayer(c1)
	room.AddPlayer(c2)

	// artificially set scores
	room.Players[c1].CorrectCount = 5
	room.Players[c1].AllReactions = []int{1000, 1000} // avg 1000
	
	room.Players[c2].CorrectCount = 3
	room.Players[c2].AllReactions = []int{2000, 2000} // avg 2000

	room.TotalRounds = 5

	room.finishGame()

	assert.Equal(t, StateFinished, room.State)

	// Consume message
	msg1Bytes := <-c1.Send
	var m WSMessage
	json.Unmarshal(msg1Bytes, &m)
	assert.Equal(t, MsgGameOver, m.Type)
}

func TestRoom_AllAnswered(t *testing.T) {
	room := NewRoom("en", "A1", model.ModeDuel, model.QuizTypeMeaningToWord, nil, nil)
	c1 := &Client{ID: uuid.New(), Username: "p1"}
	c2 := &Client{ID: uuid.New(), Username: "p2"}
	room.AddPlayer(c1)
	room.AddPlayer(c2)

	assert.False(t, room.allAnswered())

	room.Players[c1].Answered = true
	assert.False(t, room.allAnswered())

	room.Players[c2].Answered = true
	assert.True(t, room.allAnswered())
}

func TestRoom_BroadcastLeaderboard(t *testing.T) {
	vocabs := []model.Vocabulary{{ID: uuid.New(), Word: "test"}}
	room := NewRoom("en", "A1", model.ModeDuel, model.QuizTypeMeaningToWord, vocabs, nil)
	c1 := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 10)}
	room.AddPlayer(c1)

	room.Players[c1].CorrectCount = 1
	room.broadcastLeaderboard()

	msgBytes := <-c1.Send
	var m WSMessage
	json.Unmarshal(msgBytes, &m)
	
	assert.Equal(t, MsgLiveLeaderboard, m.Type)
}

func TestRoom_StartGameAndNextRound(t *testing.T) {
	vocabs := []model.Vocabulary{
		{ID: uuid.New(), Word: "apple", Meaning: "táo"},
		{ID: uuid.New(), Word: "banana", Meaning: "chuối"},
	}
	room := NewRoom("en", "A1", model.ModeDuel, model.QuizTypeMeaningToWord, vocabs, nil)
	c1 := &Client{ID: uuid.New(), Username: "p1", Send: make(chan []byte, 100)}
	c2 := &Client{ID: uuid.New(), Username: "p2", Send: make(chan []byte, 100)}
	room.AddPlayer(c1)
	room.AddPlayer(c2)
	
	room.startGame() // Directly trigger
	
	room.mu.Lock()
	assert.Equal(t, StatePlaying, room.State)
	assert.Equal(t, 1, room.CurrentRound)
	room.mu.Unlock()
	
	room.nextRound() // Directly trigger next round
	
	room.mu.Lock()
	assert.Equal(t, 2, room.CurrentRound)
	room.mu.Unlock()
}
