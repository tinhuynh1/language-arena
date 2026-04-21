package ws

import (
	"log/slog"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
)

const (
	maxRounds      = 10
	roundTimeMs    = 5000
	numTargets     = 4
	countdownMs    = 3000
	maxPlayers     = 100
	graceMs        = 500

	// Transition delays (non-blocking via time.AfterFunc)
	delayAfterCorrect   = 300 * time.Millisecond
	delayAfterTimeout   = 500 * time.Millisecond
	delayAfterAllAnswer = 800 * time.Millisecond
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
	Client       *Client
	CorrectCount int
	Reactions    []int // only correct-hit reaction times
	AllReactions []int // all reactions including penalty (5000ms) for wrong answers
	Ready        bool
	Answered     bool
}

type Room struct {
	log      *slog.Logger
	ID       string
	Code     string
	Language string
	Level    string
	Mode     model.GameMode
	QuizType model.QuizType
	State    RoomState
	HostID   uuid.UUID
	Hub      *Hub

	Players map[*Client]*PlayerState

	CurrentRound int
	TotalRounds  int
	RoundTimer    *time.Timer
	RoundStartAt  time.Time
	Vocabs        []model.Vocabulary

	mu sync.Mutex
}

func NewRoom(language, level string, mode model.GameMode, quizType model.QuizType, vocabs []model.Vocabulary, hub *Hub) *Room {
	if quizType == "" {
		quizType = model.QuizTypeMeaningToWord
	}
	roomID := uuid.New().String()[:8]
	return &Room{
		ID:          roomID,
		Code:        generateRoomCode(),
		Language:    language,
		Level:       level,
		Mode:        mode,
		QuizType:    quizType,
		State:       StateWaiting,
		TotalRounds: maxRounds,
		Vocabs:      vocabs,
		Players:     make(map[*Client]*PlayerState),
		Hub:         hub,
		log: slog.Default().With(
			"component", "GAME",
			"room_id", roomID,
		),
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
		Client:       client,
		CorrectCount: 0,
		Reactions:    make([]int, 0),
		AllReactions: make([]int, 0),
		Ready:        false,
		Answered:     false,
	}
	client.SetRoom(r)
	return true
}

func (r *Room) GetHostUsername() string {
	for c := range r.Players {
		if c.ID == r.HostID {
			return c.Username
		}
	}
	return ""
}

func (r *Room) getHostUsernameUnlocked() string {
	for c := range r.Players {
		if c.ID == r.HostID {
			return c.Username
		}
	}
	return ""
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
		Question: r.getQuestion(correctVocab),
		Targets:  targets,
		TimeMs:   roundTimeMs,
	}
	r.mu.Unlock()

	r.broadcast(WSMessage{Type: MsgRoundStart, Data: roundData})

	r.mu.Lock()
	r.RoundStartAt = time.Now()
	r.mu.Unlock()

	r.RoundTimer = time.AfterFunc(time.Duration(roundTimeMs)*time.Millisecond, func() {
		r.mu.Lock()
		if r.State == StatePlaying {
			r.State = StateRoundEnd
			r.mu.Unlock()

			r.broadcastLeaderboard()
			r.broadcast(WSMessage{Type: MsgRoundEnd, Data: RoundEndData{
				Result:   "timeout",
				NextInMs: int(delayAfterTimeout.Milliseconds()),
			}})

			time.AfterFunc(delayAfterTimeout, func() {
				r.nextRound()
			})
		} else {
			r.mu.Unlock()
		}
	})
}

func (r *Room) getQuestion(v model.Vocabulary) string {
	switch r.QuizType {
	case model.QuizTypeMeaningToWord:
		return v.Meaning
	case model.QuizTypeWordToMeaning, model.QuizTypeWordToIPA, model.QuizTypeWordToPinyin:
		return v.Word
	default:
		return v.Meaning
	}
}

func (r *Room) getLabel(v model.Vocabulary) string {
	switch r.QuizType {
	case model.QuizTypeWordToMeaning:
		return v.Meaning
	case model.QuizTypeWordToIPA:
		if v.IPA != "" {
			return v.IPA
		}
		return v.Word
	case model.QuizTypeWordToPinyin:
		if v.Pinyin != "" {
			return v.Pinyin
		}
		return v.Word
	default:
		return v.Word
	}
}

func (r *Room) generateTargets(correct model.Vocabulary) []Target {
	targets := make([]Target, 0, numTargets)

	positions := generateSpreadPositions(numTargets)

	targets = append(targets, Target{
		ID:      correct.ID.String()[:8],
		Word:    correct.Word,
		Meaning: correct.Meaning,
		Label:   r.getLabel(correct),
		X:       positions[0][0],
		Y:       positions[0][1],
		Correct: true,
	})

	used := map[string]bool{r.getLabel(correct): true}
	posIdx := 1
	for i := 0; i < numTargets-1; i++ {
		for attempts := 0; attempts < 20; attempts++ {
			idx := rand.Intn(len(r.Vocabs))
			v := r.Vocabs[idx]
			label := r.getLabel(v)
			if !used[label] {
				used[label] = true
				targets = append(targets, Target{
					ID:      v.ID.String()[:8],
					Word:    v.Word,
					Meaning: v.Meaning,
					Label:   label,
					X:       positions[posIdx][0],
					Y:       positions[posIdx][1],
					Correct: false,
				})
				posIdx++
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

	// Accept clicks during playing OR within grace period after round end
	if r.State == StateRoundEnd {
		elapsed := time.Since(r.RoundStartAt).Milliseconds()
		if elapsed > int64(roundTimeMs+graceMs) {
			r.log.Debug("hit rejected: too late", "player", client.Username, "round", r.CurrentRound, "elapsed_ms", elapsed)
			return
		}
		r.log.Debug("hit accepted in grace period", "player", client.Username, "round", r.CurrentRound, "elapsed_ms", elapsed)
	} else if r.State != StatePlaying {
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
		ps.CorrectCount++
		ps.Reactions = append(ps.Reactions, data.ReactionMs)
		ps.AllReactions = append(ps.AllReactions, data.ReactionMs)
		r.log.Info("correct hit",
			"player", client.Username,
			"round", r.CurrentRound,
			"reaction_ms", data.ReactionMs,
			"correct_count", ps.CorrectCount,
		)
	} else {
		// Penalty: count wrong answer as max reaction time
		ps.AllReactions = append(ps.AllReactions, roundTimeMs)
		r.log.Info("wrong hit",
			"player", client.Username,
			"round", r.CurrentRound,
			"penalty_ms", roundTimeMs,
			"correct_count", ps.CorrectCount,
		)
	}

	playerAvgMs := avgReaction(ps.AllReactions)

	// Send personal score update to the player
	if r.Mode == model.ModeDuel {
		for otherClient, otherPS := range r.Players {
			var opponentCorrect, opponentAvgMs int
			for c, p := range r.Players {
				if c != otherClient {
					opponentCorrect = p.CorrectCount
					opponentAvgMs = avgReaction(p.AllReactions)
					break
				}
			}
			var reactionMs int
			if otherClient == client {
				reactionMs = data.ReactionMs
			}
			otherClient.SendMessage(WSMessage{
				Type: MsgScoreUpdate,
				Data: ScoreUpdateData{
					YourCorrect:     otherPS.CorrectCount,
					YourAvgMs:       avgReaction(otherPS.AllReactions),
					OpponentCorrect: opponentCorrect,
					OpponentAvgMs:   opponentAvgMs,
					LastHitBy:       client.Username,
					ReactionMs:      reactionMs,
					IsCorrect:       isCorrect,
				},
			})
		}
	} else {
		client.SendMessage(WSMessage{
			Type: MsgScoreUpdate,
			Data: ScoreUpdateData{
				YourCorrect: ps.CorrectCount,
				YourAvgMs:   playerAvgMs,
				ReactionMs:  data.ReactionMs,
				IsCorrect:   isCorrect,
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
		r.State = StateRoundEnd
		go func() {
			r.broadcast(WSMessage{Type: MsgRoundEnd, Data: RoundEndData{
				Result:   "correct",
				NextInMs: int(delayAfterCorrect.Milliseconds()),
			}})
			time.AfterFunc(delayAfterCorrect, func() {
				r.nextRound()
			})
		}()
	}

	// For battle: advance when ALL players answered (or timer expires)
	if r.Mode == model.ModeBattle && r.allAnswered() {
		if r.RoundTimer != nil {
			r.RoundTimer.Stop()
		}
		r.State = StateRoundEnd
		go func() {
			r.broadcastLeaderboard()
			r.broadcast(WSMessage{Type: MsgRoundEnd, Data: RoundEndData{
				Result:   "all_answered",
				NextInMs: int(delayAfterAllAnswer.Milliseconds()),
			}})
			time.AfterFunc(delayAfterAllAnswer, func() {
				r.nextRound()
			})
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
		if r.Mode != model.ModeDuel || ranking[0].CorrectCount > 0 {
			winner = ranking[0].Username
		}
		if r.Mode == model.ModeDuel && len(ranking) >= 2 && ranking[0].CorrectCount == ranking[1].CorrectCount && ranking[0].AvgReactionMs == ranking[1].AvgReactionMs {
			winner = "draw"
		}
	}

	r.mu.Unlock()

	for client, ps := range r.Players {
		var opponentCorrect, opponentAvgMs int
		if r.Mode == model.ModeDuel {
			for c, p := range r.Players {
				if c != client {
					opponentCorrect = p.CorrectCount
					opponentAvgMs = avgReaction(p.AllReactions)
					break
				}
			}
		}

		playerAvgMs := avgReaction(ps.AllReactions)
		accuracy := 0
		if r.TotalRounds > 0 {
			accuracy = ps.CorrectCount * 100 / r.TotalRounds
		}

		client.SendMessage(WSMessage{
			Type: MsgGameOver,
			Data: GameOverData{
				Winner:          winner,
				YourCorrect:     ps.CorrectCount,
				YourAvgMs:       playerAvgMs,
				OpponentCorrect: opponentCorrect,
				OpponentAvgMs:   opponentAvgMs,
				Stats: GameOverStats{
					TotalRounds:   r.TotalRounds,
					AvgReactionMs: playerAvgMs,
					Accuracy:      accuracy,
				},
				Ranking: ranking,
			},
		})
	}

	r.log.Info("game finished",
		"mode", r.Mode,
		"winner", winner,
		"player_count", len(r.Players),
		"total_rounds", r.TotalRounds,
	)

	if r.Hub != nil {
		go r.Hub.SaveGameResults(r, ranking)
		go r.Hub.RemoveRoom(r)
	}
}

func (r *Room) getRanking() []LeaderboardPlayerData {
	type entry struct {
		username     string
		correctCount int
		avgMs        int
	}
	entries := make([]entry, 0, len(r.Players))
	for _, ps := range r.Players {
		entries = append(entries, entry{
			username:     ps.Client.Username,
			correctCount: ps.CorrectCount,
			avgMs:        avgReaction(ps.AllReactions),
		})
	}
	// Sort by correct count DESC, then avg reaction ASC (lower is better)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].correctCount != entries[j].correctCount {
			return entries[i].correctCount > entries[j].correctCount
		}
		return entries[i].avgMs < entries[j].avgMs
	})

	ranking := make([]LeaderboardPlayerData, len(entries))
	for i, e := range entries {
		ranking[i] = LeaderboardPlayerData{
			Rank:          i + 1,
			Username:      e.username,
			CorrectCount:  e.correctCount,
			AvgReactionMs: e.avgMs,
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
		// Transfer host if the leaving player was the host
		if client.ID == r.HostID && len(r.Players) > 0 {
			for c := range r.Players {
				r.HostID = c.ID
				r.log.Info("host transferred", "new_host", c.Username)
				r.broadcastUnlocked(WSMessage{
					Type: MsgHostChanged,
					Data: HostChangedData{NewHost: c.Username},
				})
				break
			}
		}

		// Notify remaining players
		names := r.getPlayerNames()
		r.broadcastUnlocked(WSMessage{
			Type: MsgPlayerLeft,
			Data: PlayerJoinedData{
				Username:    client.Username,
				PlayerCount: len(r.Players),
				Players:     names,
				Host:        r.getHostUsernameUnlocked(),
			},
		})

		// Clean up empty rooms
		if len(r.Players) == 0 && r.Hub != nil {
			go r.Hub.RemoveRoom(r)
		}
	}
}

func (r *Room) ReconnectPlayer(oldClient, newClient *Client) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State == StateFinished {
		return false
	}

	ps, ok := r.Players[oldClient]
	if !ok {
		return false
	}

	// Swap connection: keep correct count/reactions, replace client
	ps.Client = newClient
	delete(r.Players, oldClient)
	r.Players[newClient] = ps
	newClient.SetRoom(r)

	r.log.Info("player reconnected",
		"player", newClient.Username,
		"round", r.CurrentRound,
		"correct_count", ps.CorrectCount,
	)

	return true
}

func (r *Room) GetGameStateSync(client *Client) GameStateSyncData {
	r.mu.Lock()
	defer r.mu.Unlock()

	var state string
	switch r.State {
	case StateWaiting:
		state = "waiting"
	case StateCountdown:
		state = "countdown"
	case StatePlaying:
		state = "playing"
	case StateRoundEnd:
		state = "round_end"
	default:
		state = "finished"
	}

	var question string
	var targets []Target
	elapsedMs := 0

	if r.CurrentRound > 0 && r.CurrentRound <= len(r.Vocabs) {
		vocabIdx := (r.CurrentRound - 1) % len(r.Vocabs)
		correctVocab := r.Vocabs[vocabIdx]
		question = r.getQuestion(correctVocab)
		targets = r.generateTargets(correctVocab)
		if r.State == StatePlaying {
			elapsedMs = int(time.Since(r.RoundStartAt).Milliseconds())
		}
	}

	var yourCorrect, yourAvgMs, opponentCorrect, opponentAvgMs int
	ps, ok := r.Players[client]
	if ok {
		yourCorrect = ps.CorrectCount
		yourAvgMs = avgReaction(ps.AllReactions)
	}

	if r.Mode == model.ModeDuel {
		for c, p := range r.Players {
			if c != client {
				opponentCorrect = p.CorrectCount
				opponentAvgMs = avgReaction(p.AllReactions)
				break
			}
		}
	}

	return GameStateSyncData{
		RoomCode:        r.Code,
		Mode:            string(r.Mode),
		State:           state,
		Round:           r.CurrentRound,
		TotalRounds:     r.TotalRounds,
		Question:        question,
		Targets:         targets,
		TimeMs:          roundTimeMs,
		ElapsedMs:       elapsedMs,
		YourCorrect:     yourCorrect,
		YourAvgMs:       yourAvgMs,
		OpponentCorrect: opponentCorrect,
		OpponentAvgMs:   opponentAvgMs,
		Players:         r.getPlayerNames(),
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

// calculateScore removed — scoring replaced by avg reaction time

// generateSpreadPositions creates non-overlapping positions by dividing the
// play area into a strict 2x3 grid. Each cell is guaranteed to have no overlap.
// Y values start from 15% to avoid the HUD area. Positions are in percentages.
func generateSpreadPositions(count int) [][2]float64 {
	// Strict 2x3 grid — no zone overlap possible
	// Format: {minX, maxX, minY, maxY}
	zones := [][4]float64{
		{5, 45, 10, 35},   // top-left
		{55, 95, 10, 35},  // top-right
		{5, 45, 40, 65},   // mid-left
		{55, 95, 40, 65},  // mid-right
		{5, 45, 70, 95},   // bottom-left
		{55, 95, 70, 95},  // bottom-right
	}

	rand.Shuffle(len(zones), func(i, j int) {
		zones[i], zones[j] = zones[j], zones[i]
	})

	positions := make([][2]float64, count)
	for i := 0; i < count; i++ {
		z := zones[i%len(zones)]
		// Place in center of zone with small random offset (±10%)
		cx := (z[0] + z[1]) / 2
		cy := (z[2] + z[3]) / 2
		offsetX := (rand.Float64() - 0.5) * (z[1] - z[0]) * 0.4
		offsetY := (rand.Float64() - 0.5) * (z[3] - z[2]) * 0.4
		positions[i] = [2]float64{cx + offsetX, cy + offsetY}
	}
	return positions
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
