# 🎯 Lingo Sniper

A real-time multiplayer vocabulary aim trainer that combines CSGO-style reflex training with foreign language learning.

**Train your reflexes. Master languages. Challenge friends.**

## Tech Stack

| Component | Technology |
|-----------|-----------|
| **Backend** | Go 1.25 (Gin + gorilla/websocket) |
| **Frontend** | Next.js 15, TypeScript, Tailwind CSS |
| **Database** | PostgreSQL 16 |
| **Cache / Pub/Sub** | Redis 7 (room registry + cross-instance relay) |
| **Logging** | `log/slog` (structured JSON/Text) |
| **Infrastructure** | Docker Compose, Nginx (reverse proxy + LB) |

## Features

- **Solo Practice** — Word targets spawn at random positions. Read the meaning, click the correct word before time runs out
- **1v1 Real-time Duel** — WebSocket-powered matchmaking. Both players see the same targets. Fastest and most accurate wins
- **Battle Royale** — Room-based score race for up to 100 players. Host creates room, shares code, starts when ready
- **Quiz Types** — Meaning → Word, Word → Meaning, Word → IPA (English), Word → Pinyin (Chinese)
- **Multi-language** — English (A1–B2, 150+ words with IPA) and Chinese (HSK1–5, 100+ words with Pinyin)
- **Multi-instance** — Redis Pub/Sub relay enables horizontal scaling. Users on different backend instances can join the same room
- **Structured Logging** — `log/slog` with JSON output, contextual loggers per subsystem, request ID tracing
- **Leaderboard** — Global ranking by total score, games played, and best reaction time
- **Production-ready** — JWT auth, rate limiting, CORS, graceful shutdown, Docker deployment

## Architecture

```
                  ┌────────────┐
                  │   Nginx    │    (reverse proxy + load balancer)
                  └─────┬──────┘
             ┌──────────┼──────────┐
             ▼                     ▼
    ┌──────────────┐     ┌──────────────┐
    │  Go Backend  │     │  Go Backend  │   (N instances, stateless)
    │  Instance 1  │     │  Instance 2  │
    │  node:f28ce3 │     │  node:44319c │
    └──┬─────┬─────┘     └──┬─────┬─────┘
       │     │              │     │
       │     └──────┬───────┘     │
       │            ▼             │
       │     ┌──────────┐        │
       │     │  Redis   │        │   (room registry + Pub/Sub relay)
       │     └──────────┘        │
       ▼                         ▼
    ┌──────────┐          ┌──────────────┐
    │ Postgres │          │  Next.js     │
    └──────────┘          │  Frontend    │
                          └──────────────┘
```

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (for local dev)
- Node.js 20+ (for local dev)

### Run with Docker Compose

```bash
# Clone and start
git clone <repo-url>
cd language-arena
cp .env.example .env

# Single instance (default)
docker compose up --build

# Multi-instance (2 backends + nginx load balancer)
docker compose -f docker-compose.yml -f docker-compose.scale.yml up --build

# Open browser
open http://localhost
```

### Local Development

```bash
# 1. Start database & redis
docker compose up postgres redis -d

# 2. Run backend
cd backend
LOG_FORMAT=text go run ./cmd/server

# 3. Run frontend (new terminal)
cd frontend
npm install
npm run dev
```

### Seed Vocabulary Data

```bash
# Connect to postgres and run seed
docker exec -i $(docker compose ps -q postgres) psql -U lingouser -d lingodb < backend/internal/migration/004_seed_vocabularies.sql
```

## API Reference

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/auth/register` | ✗ | Register new user |
| POST | `/api/v1/auth/login` | ✗ | Login → JWT token |
| GET | `/api/v1/vocab?lang=en\|zh` | ✗ | Get vocabulary list |
| GET | `/api/v1/leaderboard` | ✗ | Top players ranking |
| GET | `/api/v1/stats/me` | ✓ | Player stats |
| GET | `/api/v1/online` | ✗ | Online player count |
| GET | `/api/v1/ws/game?token=<JWT>` | ✓ | WebSocket game connection |

## WebSocket Protocol

```json
// Client → Server
{"type": "join_queue", "data": {"mode": "duel", "language": "en", "level": "B1"}}
{"type": "create_room", "data": {"language": "en", "level": "A1"}}
{"type": "join_room", "data": {"room_code": "A8K3MN"}}
{"type": "start_game"}
{"type": "ready"}
{"type": "target_hit", "data": {"target_id": "abc", "reaction_ms": 342}}
{"type": "leave_room"}

// Server → Client
{"type": "queue_joined", "data": {"status": "waiting"}}
{"type": "match_found", "data": {"room_id": "xyz", "opponent": "Player2", "mode": "duel"}}
{"type": "room_created", "data": {"room_code": "A8K3MN", "room_id": "xyz", "language": "en", "level": "A1"}}
{"type": "player_joined", "data": {"username": "Player2", "player_count": 3, "players": [...]}}
{"type": "countdown", "data": {"ms": 3000}}
{"type": "round_start", "data": {"round": 1, "total": 10, "question": "Xin chào", "targets": [...], "time_ms": 5000}}
{"type": "score_update", "data": {"you": 450, "opponent": 380, "reaction_ms": 342}}
{"type": "live_leaderboard", "data": {"round": 3, "players": [{"rank":1,"username":"P1","score":500}]}}
{"type": "round_end", "data": {"result": "timeout"}}
{"type": "game_over", "data": {"winner": "Player1", "your_score": 450, "stats": {...}, "ranking": [...]}}
{"type": "opponent_left"}
{"type": "error", "data": "not in a room"}
```

## Backend Flows (for Backend Developers)

> All message types are defined in `backend/internal/ws/message.go`  
> Core game logic lives in `backend/internal/ws/room.go`

---

### Flow 1: Authentication (REST)

```
Client                              Server (auth_handler.go → auth_service.go → user_repo.go)
  │                                    │
  ├─ POST /api/v1/auth/register ──────►│  Validate fields → Hash password (bcrypt)
  │  {username, email, password}       │  → INSERT into users table
  │◄─── 201 {token, user} ────────────┤  → Generate JWT token (HS256, 72h expiry)
  │                                    │
  ├─ POST /api/v1/auth/login ─────────►│  Find user by email → Compare bcrypt hash
  │  {email, password}                 │  → Generate JWT token
  │◄─── 200 {token, user} ────────────┤
```

**Key files:** `handler/auth_handler.go` → `service/auth_service.go` → `repository/user_repo.go`

---

### Flow 2: WebSocket Connection Lifecycle

```
Client                              Server (main.go → client.go → hub.go)
  │                                    │
  ├─ GET /api/v1/ws/game?token=JWT ───►│  middleware.AuthWS() extracts userID from JWT
  │                                    │  Upgrade HTTP → WebSocket (gorilla/websocket)
  │                                    │  Create Client{ID, Username, Conn, Send channel}
  │                                    │  hub.Register ← client
  │                                    │  Launch goroutines: client.ReadPump(), client.WritePump()
  │◄─── WS connected ─────────────────┤
  │                                    │
  │  ... game messages ...             │
  │                                    │
  ├─ [connection closes] ─────────────►│  ReadPump exits → hub.Unregister ← client
  │                                    │  → matchmaker.Remove(client)
  │                                    │  → room.RemovePlayer(client)
  │                                    │  → close(client.Send)
```

**Goroutine model:** Each client spawns 2 goroutines:
- `ReadPump`: Reads JSON messages from WS → calls `hub.HandleMessage(client, msg)` → routes to handler by `msg.Type`
- `WritePump`: Drains `client.Send` channel → writes to WS connection. Also sends periodic Ping frames (every 54s) for keepalive.

**Key files:** `cmd/server/main.go` (WS upgrade), `ws/client.go` (Read/WritePump), `ws/hub.go` (message router)

---

### Flow 3: Solo Practice

```
Client                  Hub (hub.go)                    Room (room.go)
  │                        │                               │
  ├─ join_queue ──────────►│ mode == "solo"                 │
  │  {mode:"solo",         │ GetVocabs(lang, level, 14)     │
  │   language:"en",       │ NewRoom(ModeSolo, vocabs)      │
  │   level:"A1"}          │ room.AddPlayer(client)         │
  │                        │                               ┌┤
  │◄─ match_found ────────┤  {room_id, mode:"solo"}       ││
  │                        │                               ││
  ├─ ready ───────────────►│ → room.SetReady(client)       ││
  │                        │   allReady()? → startGame()   ││
  │◄─ countdown ──────────┤◄──────────────────────────────┘│
  │   {ms: 3000}           │                               │
  │   [3 second wait]      │                               │
  │◄─ round_start ────────┤◄── nextRound() ───────────────┤
  │   {round, question,    │   generateTargets(correct)    │
  │    targets[], time_ms} │   broadcast to all players    │
  │                        │                               │
  │  [See Flow 6: Game Round Loop]                         │
```

**Key point:** Solo creates a Room with 1 player. No matchmaking needed. `allReady()` returns true when the single player hits READY.

---

### Flow 4: 1v1 Duel Matchmaking

```
Player A                Matchmaker (matchmaker.go)         Player B
  │                           │                               │
  ├─ join_queue ─────────────►│ queue is empty                │
  │  {mode:"duel",            │ Enqueue(A, "en", "B1")       │
  │   language:"en",          │ queue = [{A, en, B1}]        │
  │   level:"B1"}             │                               │
  │◄─ queue_joined ──────────┤                               │
  │   {status:"waiting"}      │                               │
  │                           │                               │
  │                           │◄────── join_queue ────────────┤
  │                           │  {mode:"duel", language:"en", │
  │                           │   level:"B1"}                 │
  │                           │                               │
  │                           │ Match! Same lang + level      │
  │                           │ Remove A from queue           │
  │                           │ GetVocabs("en", "B1", 14)    │
  │                           │ NewRoom(ModeDuel, vocabs)     │
  │                           │ room.AddPlayer(A)             │
  │                           │ room.AddPlayer(B)             │
  │                           │                               │
  │◄─ match_found ───────────┤── match_found ────────────────►│
  │  {room_id, opponent:"B"} │  {room_id, opponent:"A"}      │
  │                           │                               │
  ├─ ready ──────────────────►│                               │
  │                           │◄───────────── ready ──────────┤
  │                           │                               │
  │  [Both ready + 2 players  │  → startGame() triggered]     │
  │◄─ countdown ─────────────┤── countdown ──────────────────►│
```

**Matching rules (matchmaker.go):**  
- Queue is an in-memory `[]queueEntry` protected by `sync.Mutex`
- Match condition: `entry.Language == language && entry.Level == level && entry.Client.ID != client.ID`
- First match wins (FIFO). No Elo or skill-based matching.

---

### Flow 5: Battle Royale (Score Race, up to 100 players)

```
Host                    Hub (hub.go)                    Player X
  │                        │                               │
  ├─ create_room ─────────►│ NewRoom(ModeBattle, vocabs)   │
  │  {language, level}     │ room.HostID = host.ID         │
  │                        │ room.Code = "A8K3MN" (random) │
  │◄─ room_created ───────┤ AddRoom → RoomByCode map      │
  │  {room_code:"A8K3MN"} │                               │
  │                        │                               │
  │  [Host shares code     │                               │
  │   to friends]          │                               │
  │                        │◄───── join_room ──────────────┤
  │                        │  {room_code:"A8K3MN"}         │
  │                        │  RoomByCode["A8K3MN"] lookup  │
  │                        │  room.AddPlayer(X)            │
  │                        │                               │
  │◄─ player_joined ──────┤── match_found ────────────────►│
  │  {username:"X",        │  {room_id, mode:"battle"}     │
  │   player_count: 5,     │                               │
  │   players: [...]}      │── player_joined ─────────────►│
  │                        │                               │
  │  [...more players join up to 100]                      │
  │                        │                               │
  ├─ start_game ──────────►│ HostID check                  │
  │                        │ Auto-ready all players        │
  │                        │ startGame()                   │
  │◄─ countdown ──────────┤── countdown ──────────────────►│
  │   {ms: 3000}           │                               │
```

**Key differences from Duel:**
- Room capacity: 100 (vs 2 for Duel). Constant `maxPlayers` in `room.go`.
- Room code: 6-char alphanumeric (`ABCDEFGHJKLMNPQRSTUVWXYZ23456789`), stored in `hub.RoomByCode` map.
- Host-only start: Only the player whose `client.ID == room.HostID` can trigger `start_game`.
- No matchmaker involvement: Room is created explicitly, players join by code.

---

### Flow 6: Game Round Loop (all modes)

```
                        Room (room.go)
                           │
   startGame()             │
   ├─ State = Countdown    │
   ├─ broadcast: countdown │→ All players get {ms: 3000}
   ├─ time.Sleep(3s)       │
   │                       │
   nextRound()             │
   ├─ CurrentRound++       │
   ├─ State = Playing      │
   ├─ vocabIdx = (round-1) % len(vocabs)    ← Correct answer from vocab list
   ├─ generateTargets(correct)               ← 1 correct + 3 wrong, grid positions
   ├─ broadcast: round_start                 → {round, total, question, targets[], time_ms}
   ├─ Start RoundTimer (5s)                  ← AfterFunc goroutine
   │                       │
   │  Player hits target:  │
   ├─ HandleHit(client, {target_id, reaction_ms})
   │  ├─ Check: state == Playing? Already answered?
   │  ├─ Compare target_id with correct vocab ID
   │  ├─ Correct: score += (5000 - reactionMs) * 1000 / 5000  (min 100)
   │  ├─ Wrong:   score -= 50  (floor at 0)
   │  ├─ Send: score_update to player(s)
   │  │
   │  ├─ [Solo/Duel] If correct → stop timer → sleep 1s → nextRound()
   │  ├─ [Battle]    If all answered → stop timer → broadcastLeaderboard → sleep 2s → nextRound()
   │  ├─ [Battle]    After each hit → broadcastLeaderboard (top 5 live ranking)
   │                       │
   │  RoundTimer fires:    │
   ├─ State = RoundEnd     │
   ├─ broadcastLeaderboard │
   ├─ broadcast: round_end │→ {result: "timeout"}
   ├─ sleep 2s             │
   ├─ nextRound()          │
   │                       │
   │  After round 10:      │
   finishGame()            │
   ├─ State = Finished     │
   ├─ getRanking()         │→ Sort players by score DESC
   ├─ Determine winner     │→ Duel: check for draw
   ├─ Send: game_over      │→ Each player gets personalized {your_score, opponent_score, ranking}
```

**Scoring formula (`calculateScore`):**
```go
bonus := (roundTimeMs - reactionMs) * 1000 / roundTimeMs
// roundTimeMs = 5000, so: faster click = higher score
// Example: 500ms → (5000-500)*1000/5000 = 900 points
// Example: 4500ms → (5000-4500)*1000/5000 = 100 points (minimum)
```

**Target generation (`generateSpreadPositions`):**
- Canvas divided into 6 grid zones to prevent overlapping
- 4 targets per round: 1 correct + 3 distractors (random from vocab pool)
- Zones are shuffled each round so correct answer position is unpredictable

---

### Flow 7: Disconnect / Leave

```
Client                  Hub (hub.go)                    Room (room.go)
  │                        │                               │
  ├─ leave_room ──────────►│ room.RemovePlayer(client)     │
  │  (voluntary)           │ matchmaker.Remove(client)     │
  │                        │ client.SetRoom(nil)           │
  │                        │                               │
  │  [connection drops] ──►│ ReadPump exits                │
  │  (involuntary)         │ Unregister ← client           │
  │                        │ matchmaker.Remove(client)     │
  │                        │ room.RemovePlayer(client)     │
  │                        │                               │
  │                        │ [Duel mode:]                   │
  │                        │ broadcast: opponent_left       │
  │                        │ RoundTimer.Stop()              │
  │                        │ State = Finished               │
  │                        │                               │
  │                        │ [Battle mode:]                 │
  │                        │ broadcast: player_left         │
  │                        │ {username, player_count,       │
  │                        │  players[]}                    │
```

## Project Structure

```
language-arena/
├── backend/                       # Go backend
│   ├── cmd/server/main.go         # Entry point, router, migrations
│   ├── internal/
│   │   ├── config/                # Environment config (LOG_LEVEL, LOG_FORMAT, etc.)
│   │   ├── model/                 # Domain models (game modes, quiz types)
│   │   ├── repository/            # Data access (PostgreSQL)
│   │   ├── service/               # Business logic
│   │   ├── handler/               # HTTP handlers
│   │   ├── ws/                    # WebSocket game engine
│   │   │   ├── hub.go             # Central message router
│   │   │   ├── room.go            # Game loop (rounds, scoring, timer)
│   │   │   ├── client.go          # WS client (ReadPump/WritePump)
│   │   │   ├── matchmaker.go      # Duel queue matching
│   │   │   └── redis_adapter.go   # Cross-instance Pub/Sub relay
│   │   ├── middleware/            # Auth, CORS, rate limit, request logger
│   │   └── migration/            # SQL migrations (embedded via go:embed)
│   └── pkg/
│       ├── logger/                # slog init (JSON/Text, log level)
│       └── response/              # HTTP response helpers
├── frontend/                      # Next.js frontend
│   └── src/
│       ├── app/                   # Pages (App Router)
│       ├── components/            # UI + Game components
│       ├── hooks/                 # useAuth, useWebSocket, useGame
│       └── lib/                   # API client
├── docs/                          # Technical documentation
├── docker-compose.yml             # Standard deployment
├── docker-compose.scale.yml       # Multi-instance override (2 backends)
├── nginx/
│   ├── default.conf               # Single-instance proxy
│   └── default.scale.conf         # Upstream pool (multi-instance)
└── README.md
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend server port |
| `DB_URL` | `postgres://lingouser:lingopass@localhost:5432/lingodb?sslmode=disable` | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection (optional — graceful fallback) |
| `JWT_SECRET` | `dev-secret-change-in-prod` | JWT signing key |
| `JWT_EXPIRATION` | `24h` | Token expiry |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` (production) or `text` (development) |
| `GIN_MODE` | `debug` | `debug` or `release` |

## Documentation

| Document | Description |
|----------|-------------|
| [BACKEND_TECHNICAL_VI.md](docs/BACKEND_TECHNICAL_VI.md) | Detailed backend technical documentation (Vietnamese) |
| [INTERVIEW_QUESTIONS_VI.md](docs/INTERVIEW_QUESTIONS_VI.md) | Interview preparation Q&A (Vietnamese) |

## License

MIT
