# 🎯 Lingo Sniper

A real-time multiplayer vocabulary aim trainer that combines CSGO-style reflex training with foreign language learning.

**Train your reflexes. Master languages. Challenge friends.**

## Tech Stack

| Component | Technology |
|-----------|-----------|
| **Backend** | Go 1.22 (Gin + gorilla/websocket) |
| **Frontend** | Next.js 15, TypeScript, Tailwind CSS |
| **Database** | PostgreSQL 16 |
| **Cache** | Redis 7 |
| **Infrastructure** | Docker Compose |

## Features

- **Solo Practice** — Word targets spawn at random positions. Read the meaning, click the correct word before time runs out
- **1v1 Real-time Duel** — WebSocket-powered matchmaking. Both players see the same targets. Fastest and most accurate wins
- **Leaderboard** — Global ranking by total score, games played, and best reaction time
- **Multi-language** — English (general + CSGO callouts) and Chinese (basic → idioms)
- **Production-ready** — JWT auth, rate limiting, CORS, graceful shutdown, Docker deployment

## Architecture

```
┌──────────────┐     ┌──────────────────────────┐
│  Next.js     │ WS  │  Go Backend (Gin)        │
│  Frontend    ├────►│  ├── REST API (auth,vocab)│
│  (SSR)       │     │  ├── WebSocket Hub        │
│              │ API │  ├── Game Rooms            │
│              ├────►│  └── Matchmaker            │
└──────────────┘     └─────┬──────────┬──────────┘
                           │          │
                     ┌─────▼──┐  ┌────▼────┐
                     │Postgres│  │  Redis   │
                     └────────┘  └─────────┘
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

# Start all services
docker compose up --build

# Open browser
open http://localhost:3000
```

### Local Development

```bash
# 1. Start database & redis
docker compose up postgres redis -d

# 2. Run backend
cd backend
go run ./cmd/server

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
{"type": "join_queue", "data": {"mode": "duel", "language": "en"}}
{"type": "ready"}
{"type": "target_hit", "data": {"target_id": "abc", "reaction_ms": 342}}

// Server → Client
{"type": "match_found", "data": {"room_id": "xyz", "opponent": "Player2"}}
{"type": "round_start", "data": {"round": 1, "question": "Xin chào", "targets": [...]}}
{"type": "score_update", "data": {"you": 450, "opponent": 380}}
{"type": "game_over", "data": {"winner": "Player1", "stats": {...}}}
```

## Project Structure

```
language-arena/
├── backend/                   # Go backend
│   ├── cmd/server/main.go     # Entry point
│   ├── internal/
│   │   ├── config/            # Environment config
│   │   ├── model/             # Domain models
│   │   ├── repository/        # Data access (PostgreSQL)
│   │   ├── service/           # Business logic
│   │   ├── handler/           # HTTP handlers
│   │   ├── ws/                # WebSocket game engine
│   │   ├── middleware/        # Auth, CORS, rate limit
│   │   └── migration/         # SQL migrations + seed
│   └── pkg/                   # Shared utilities
├── frontend/                  # Next.js frontend
│   ├── src/
│   │   ├── app/               # Pages (App Router)
│   │   ├── components/        # UI + Game components
│   │   ├── hooks/             # useAuth, useWebSocket, useGame
│   │   └── lib/               # API client
├── docker-compose.yml
└── README.md
```

## License

MIT
