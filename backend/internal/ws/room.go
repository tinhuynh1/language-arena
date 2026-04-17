package ws

import (
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
)

const (
	maxRounds    = 10
	roundTimeMs  = 5000
	numTargets   = 4
	countdownMs  = 3000
	maxPlayers   = 100
)

type RoomState int

const (
	StateWaiting RoomState = iota
	StateCountdown
	StatePlaying
	StateRoundEnd
	StateFinished
)

type PlayerState struct {
	Client    *Client
	Score     int
	Reactions []int
	Ready     bool
	Answered  bool
}

type Room struct {
	ID       string
	Code     string
	Language string
	Level    string
	Mode     model.GameMode
	State    RoomState
	HostID   uuid.UUID

	Players map[*Client]*PlayerState

	CurrentRound int
	TotalRounds  int
	RoundTimer   *time.Timer
	Vocabs       []model.Vocabulary

	mu sync.Mutex
}

func NewRoom(language, level string, mode model.GameMode, vocabs []model.Vocabulary) *Room {
	return &Room{
		ID:          uuid.New().String()[:8],
		Code:        generateRoomCode(),
		Language:    language,
		Level:       level,
		Mode:        mode,
		State:       StateWaiting,
		TotalRounds: maxRounds,
		Vocabs:      vocabs,
		Players:     make(map[*Client]*PlayerState),
	}
}

func generateRoomCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func (r *Room) AddPlayer(client *Client) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != StateWaiting {
		return false
	}

	if r.Mode == model.ModeDuel && len(r.Players) >= 2 {
		return false
	}
	if len(r.Players) >= maxPlayers {
		return false
	}

	r.Players[client] = &PlayerState{
		Client:    client,
		Score:     0,
		Reactions: make([]int, 0),
		Ready:     false,
		Answered:  false,
	}
	client.SetRoom(r)
	return true
}

func (r *Room) SetReady(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ps, ok := r.Players[client]
	if !ok {
		return
	}
	ps.Ready = true

	if r.Mode == model.ModeSolo && r.allReady() {
		go r.startGame()
	} else if r.Mode == model.ModeDuel && len(r.Players) == 2 && r.allReady() {
		go r.startGame()
	}
	// Battle mode: host calls start_game explicitly
}

func (r *Room) StartByHost(client *Client) {
	r.mu.Lock()
	if client.ID != r.HostID {
		r.mu.Unlock()
		client.SendMessage(WSMessage{Type: MsgError, Data: "only host can start"})
		return
	}
	if len(r.Players) < 1 {
		r.mu.Unlock()
		client.SendMessage(WSMessage{Type: MsgError, Data: "not enough players"})
		return
	}
	// Auto-ready all players in battle mode
	for _, ps := range r.Players {
		ps.Ready = true
	}
	r.mu.Unlock()
	r.startGame()
}

func (r *Room) allReady() bool {
	for _, ps := range r.Players {
		if !ps.Ready {
			return false
		}
	}
	return len(r.Players) > 0
}

func (r *Room) startGame() {
	r.mu.Lock()
	r.State = StateCountdown
	r.CurrentRound = 0
	r.mu.Unlock()

	r.broadcast(WSMessage{Type: MsgCountdown, Data: map[string]int{"ms": countdownMs}})
	time.Sleep(time.Duration(countdownMs) * time.Millisecond)

	r.nextRound()
}

func (r *Room) nextRound() {
	r.mu.Lock()

	r.CurrentRound++
	if r.CurrentRound > r.TotalRounds {
		r.mu.Unlock()
		r.finishGame()
		return
	}

	r.State = StatePlaying

	// Reset answered state for all players
	for _, ps := range r.Players {
		ps.Answered = false
	}

	vocabIdx := (r.CurrentRound - 1) % len(r.Vocabs)
	correctVocab := r.Vocabs[vocabIdx]

	targets := r.generateTargets(correctVocab)

	roundData := RoundStartData{
		Round:    r.CurrentRound,
		Total:    r.TotalRounds,
		Question: correctVocab.Meaning,
		Targets:  targets,
		TimeMs:   roundTimeMs,
	}
	r.mu.Unlock()

	r.broadcast(WSMessage{Type: MsgRoundStart, Data: roundData})

	r.RoundTimer = time.AfterFunc(time.Duration(roundTimeMs)*time.Millisecond, func() {
		r.mu.Lock()
		if r.State == StatePlaying {
			r.State = StateRoundEnd
			r.mu.Unlock()

			r.broadcastLeaderboard()
			r.broadcast(WSMessage{Type: MsgRoundEnd, Data: map[string]string{"result": "timeout"}})

			time.Sleep(2 * time.Second)
			r.nextRound()
		} else {
			r.mu.Unlock()
		}
	})
}

func (r *Room) generateTargets(correct model.Vocabulary) []Target {
	targets := make([]Target, 0, numTargets)

	targets = append(targets, Target{
		ID:      correct.ID.String()[:8],
		Word:    correct.Word,
		Meaning: correct.Meaning,
		X:       randomPosition(),
		Y:       randomPosition(),
		Correct: true,
	})

	used := map[string]bool{correct.Word: true}
	for i := 0; i < numTargets-1; i++ {
		for attempts := 0; attempts < 20; attempts++ {
			idx := rand.Intn(len(r.Vocabs))
			v := r.Vocabs[idx]
			if !used[v.Word] {
				used[v.Word] = true
				targets = append(targets, Target{
					ID:      v.ID.String()[:8],
					Word:    v.Word,
					Meaning: v.Meaning,
					X:       randomPosition(),
					Y:       randomPosition(),
					Correct: false,
				})
				break
			}
		}
	}

	rand.Shuffle(len(targets), func(i, j int) {
		targets[i], targets[j] = targets[j], targets[i]
	})

	return targets
}

func (r *Room) HandleHit(client *Client, data TargetHitData) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != StatePlaying {
		return
	}

	ps, ok := r.Players[client]
	if !ok || ps.Answered {
		return
	}

	vocabIdx := (r.CurrentRound - 1) % len(r.Vocabs)
	correctVocab := r.Vocabs[vocabIdx]
	correctID := correctVocab.ID.String()[:8]

	isCorrect := data.TargetID == correctID
	ps.Answered = true

	if isCorrect {
		points := calculateScore(data.ReactionMs)
		ps.Score += points
		ps.Reactions = append(ps.Reactions, data.ReactionMs)
	} else {
		ps.Score -= 50
		if ps.Score < 0 {
			ps.Score = 0
		}
	}

	// Send personal score update to the player
	if r.Mode == model.ModeDuel {
		// For duel, send opponent score
		for otherClient, otherPS := range r.Players {
			var opponentScore int
			for c, p := range r.Players {
				if c != otherClient {
					opponentScore = p.Score
					break
				}
			}
			otherClient.SendMessage(WSMessage{
				Type: MsgScoreUpdate,
				Data: ScoreUpdateData{
					You:        otherPS.Score,
					Opponent:   opponentScore,
					LastHitBy:  client.Username,
					ReactionMs: data.ReactionMs,
				},
			})
		}
	} else {
		// Solo/Battle: just send personal score
		client.SendMessage(WSMessage{
			Type: MsgScoreUpdate,
			Data: ScoreUpdateData{
				You:        ps.Score,
				ReactionMs: data.ReactionMs,
			},
		})
	}

	// In battle mode, broadcast live leaderboard after each hit
	if r.Mode == model.ModeBattle {
		go r.broadcastLeaderboard()
	}

	// For solo/duel: advance round when correct answer is given
	if isCorrect && (r.Mode == model.ModeSolo || r.Mode == model.ModeDuel) {
		if r.RoundTimer != nil {
			r.RoundTimer.Stop()
		}
		go func() {
			time.Sleep(1 * time.Second)
			r.nextRound()
		}()
	}

	// For battle: advance when ALL players answered (or timer expires)
	if r.Mode == model.ModeBattle && r.allAnswered() {
		if r.RoundTimer != nil {
			r.RoundTimer.Stop()
		}
		go func() {
			r.broadcastLeaderboard()
			time.Sleep(2 * time.Second)
			r.nextRound()
		}()
	}
}

func (r *Room) allAnswered() bool {
	for _, ps := range r.Players {
		if !ps.Answered {
			return false
		}
	}
	return true
}

func (r *Room) finishGame() {
	r.mu.Lock()
	r.State = StateFinished

	ranking := r.getRanking()

	var winner string
	if len(ranking) > 0 {
		if r.Mode != model.ModeDuel || ranking[0].Score > 0 {
			winner = ranking[0].Username
		}
		if r.Mode == model.ModeDuel && len(ranking) >= 2 && ranking[0].Score == ranking[1].Score {
			winner = "draw"
		}
	}

	r.mu.Unlock()

	// Send personalized game over to each player
	for client, ps := range r.Players {
		var opponentScore int
		if r.Mode == model.ModeDuel {
			for c, p := range r.Players {
				if c != client {
					opponentScore = p.Score
					break
				}
			}
		}

		avgReaction := avgReaction(ps.Reactions)
		accuracy := 0
		if r.TotalRounds > 0 {
			accuracy = len(ps.Reactions) * 100 / r.TotalRounds
		}

		client.SendMessage(WSMessage{
			Type: MsgGameOver,
			Data: GameOverData{
				Winner:        winner,
				YourScore:     ps.Score,
				OpponentScore: opponentScore,
				Stats: GameOverStats{
					TotalRounds:   r.TotalRounds,
					AvgReactionMs: avgReaction,
					Accuracy:      accuracy,
				},
				Ranking: ranking,
			},
		})
	}

	log.Printf("Game %s finished. Mode: %s, Winner: %s, Players: %d",
		r.ID, r.Mode, winner, len(r.Players))
}

func (r *Room) getRanking() []LeaderboardPlayerData {
	type entry struct {
		username string
		score    int
	}
	entries := make([]entry, 0, len(r.Players))
	for _, ps := range r.Players {
		entries = append(entries, entry{username: ps.Client.Username, score: ps.Score})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].score > entries[j].score
	})

	ranking := make([]LeaderboardPlayerData, len(entries))
	for i, e := range entries {
		ranking[i] = LeaderboardPlayerData{
			Rank:     i + 1,
			Username: e.username,
			Score:    e.score,
		}
	}
	return ranking
}

func (r *Room) broadcastLeaderboard() {
	r.mu.Lock()
	ranking := r.getRanking()
	round := r.CurrentRound
	r.mu.Unlock()

	// Show top 5 for live updates
	top := ranking
	if len(top) > 5 {
		top = top[:5]
	}

	r.broadcast(WSMessage{
		Type: MsgLiveLeaderboard,
		Data: LiveLeaderboardData{
			Round:   round,
			Players: top,
		},
	})
}

func (r *Room) RemovePlayer(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Players, client)

	if r.State == StateFinished {
		return
	}

	if r.Mode == model.ModeDuel {
		for c := range r.Players {
			c.SendMessage(WSMessage{Type: MsgOpponentLeft})
		}
		if r.RoundTimer != nil {
			r.RoundTimer.Stop()
		}
		r.State = StateFinished
	} else if r.Mode == model.ModeBattle {
		// Notify remaining players
		names := r.getPlayerNames()
		r.broadcastUnlocked(WSMessage{
			Type: MsgPlayerLeft,
			Data: PlayerJoinedData{
				Username:    client.Username,
				PlayerCount: len(r.Players),
				Players:     names,
			},
		})
	}
}

func (r *Room) getPlayerNames() []string {
	names := make([]string, 0, len(r.Players))
	for c := range r.Players {
		names = append(names, c.Username)
	}
	return names
}

func (r *Room) broadcast(msg WSMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.broadcastUnlocked(msg)
}

func (r *Room) broadcastUnlocked(msg WSMessage) {
	for client := range r.Players {
		client.SendMessage(msg)
	}
}

func calculateScore(reactionMs int) int {
	if reactionMs <= 0 {
		return 0
	}
	base := 1000
	bonus := (roundTimeMs - reactionMs) * base / roundTimeMs
	if bonus < 100 {
		bonus = 100
	}
	return bonus
}

func randomPosition() float64 {
	return 10 + rand.Float64()*80
}

func avgReaction(reactions []int) int {
	if len(reactions) == 0 {
		return 0
	}
	sum := 0
	for _, r := range reactions {
		sum += r
	}
	return sum / len(reactions)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	_ = strings.Contains
}
