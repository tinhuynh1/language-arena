# Lingo Sniper

A real-time multiplayer vocabulary aim trainer that combines CSGO-style reflex training with foreign language learning.

**Train your reflexes. Master languages. Challenge friends.**

Live: [https://lingosniper.lol](https://lingosniper.lol)

---

## Table of Contents

- [Tech Stack](#tech-stack)
- [Features](#features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Production Deployment (GKE)](#production-deployment-gke)
- [CI/CD Pipeline](#cicd-pipeline)
- [API Reference](#api-reference)
- [WebSocket Protocol](#websocket-protocol)
- [Backend Flows](#backend-flows)
- [Database](#database)
- [Environment Variables](#environment-variables)
- [Project Structure](#project-structure)

---

## Tech Stack

| Component | Technology |
|-----------|-----------|
| **Backend** | Go 1.22 · Gin · gorilla/websocket |
| **Frontend** | Next.js 15 · TypeScript · Tailwind CSS v4 |
| **Database** | PostgreSQL 16 (Cloud SQL on GCP) |
| **Cache / Pub/Sub** | Redis 7 — room registry, distributed matchmaking queue, cross-instance relay |
| **Container Orchestration** | Kubernetes (GKE Autopilot) |
| **Image Registry** | Google Artifact Registry (asia-southeast1) |
| **TLS / Ingress** | nginx-ingress-controller · cert-manager · Let's Encrypt |
| **Logging** | `log/slog` — structured JSON, per-subsystem loggers, request ID tracing |
| **Local Dev** | Docker Compose · Nginx (reverse proxy + load balancer) |

---

## Features

- **Solo Practice** — Word targets spawn at random canvas positions. Read the prompt, click the correct target before time runs out.
- **1v1 Duel** — Real-time matchmaking via WebSocket. Both players see the same targets simultaneously; fastest and most accurate wins.
- **Battle Royale** — Room-based score race for up to 100 players. Host creates room, shares a 6-character code, starts when ready.
- **Quiz Types** — Meaning → Word · Word → Meaning · Word → IPA (English) · Word → Pinyin (Chinese)
- **Multi-language** — English (A1–B2, 150+ words with IPA) and Chinese (HSK 1–5, 100+ words with Pinyin)
- **Horizontal Scaling** — Redis atomic queue enables cross-instance matchmaking; Redis Pub/Sub relays WebSocket messages between backend pods.
- **Structured Logging** — JSON logs with `log/slog`, per-subsystem context loggers, request ID propagation through the full call chain.
- **Leaderboard** — Global ranking by average reaction time, total correct answers, and best reaction time.
- **Production-ready** — JWT auth (HS256, 72h expiry) · bcrypt password hashing · rate limiting · CORS · graceful shutdown · DB connection retry loop.

---

## Architecture

```
                        Internet
                            │
                   ┌────────▼────────┐
                   │  GKE Ingress    │  (nginx-ingress + cert-manager TLS)
                   │  lingosniper.lol│  Cookie-based session affinity
                   └────────┬────────┘
                 /api        │           /
           ┌────────┐        │      ┌────────────┐
           │        │        │      │  Next.js   │
           │        ▼        │      │  Frontend  │
           │  ┌───────────┐  │      └────────────┘
           │  │ Backend   │◄─┤
           │  │  Pod 1    │  │
           │  │ node:xxxx │  │
           │  └─────┬─────┘  │
           │        │        │
           │  ┌─────▼─────┐  │  ┌───────────┐
           │  │ Backend   │◄─┘  │  Backend  │
           │  │  Pod 2    │─────►│  Pod N    │
           │  └─────┬─────┘     └─────┬─────┘
           │        │                 │
           │        └────────┬────────┘
           │                 │
           │        ┌────────▼────────┐
           │        │     Redis       │  Matchmaking queue · room registry
           │        │  (GKE in-pod)   │  Pub/Sub relay between pods
           │        └─────────────────┘
           │
           │        ┌─────────────────┐
           └──────► │  Cloud SQL      │  PostgreSQL 16 (managed)
                    │  (private IP)   │  Auth via Cloud SQL Proxy sidecar
                    └─────────────────┘
```

Each backend pod runs with a **Cloud SQL Auth Proxy native sidecar** (`initContainer` with `restartPolicy: Always`). The proxy starts and becomes ready before the backend container receives any traffic, ensuring the `127.0.0.1:5432` socket is available at startup.

The nginx ingress uses **cookie-based session affinity** (`LINGO_BACKEND` cookie) so a user's WebSocket connection always routes to the same pod. Redis bridges users on different pods at the matchmaking layer.

---

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (local dev only)
- Node.js 20+ (local dev only)

### Run with Docker Compose

```bash
git clone https://github.com/tinhuynh1/language-arena.git
cd language-arena
cp .env.example .env          # edit secrets as needed

docker compose up --build
# App available at http://localhost
```

Services started:

| Service | Role |
|---------|------|
| `postgres` | PostgreSQL 16 with health check |
| `redis` | Redis 7 (in-memory queue + pub/sub) |
| `migrate` | One-shot migration runner — exits after schema is applied |
| `backend` | Go API server on port 8080 |
| `frontend` | Next.js on port 3000 |
| `nginx` | Reverse proxy on port 80 |

### Local Development

```bash
# 1. Start infrastructure
docker compose up postgres redis -d

# 2. Run backend
cd backend
LOG_FORMAT=text go run ./cmd/server

# 3. Run frontend (separate terminal)
cd frontend
npm install
npm run dev
```

### Seed Vocabulary Data

```bash
docker exec -i $(docker compose ps -q postgres) \
  psql -U lingouser -d lingodb \
  < backend/internal/migration/002_seed.sql
```

---

## Production Deployment (GKE)

### Infrastructure Overview

| Resource | Details |
|----------|---------|
| Cluster | GKE Autopilot · `lingo-sniper-cluster` · `us-central1-a` |
| Namespace | `lingo-sniper` |
| Registry | `asia-southeast1-docker.pkg.dev/lingo-sniper-prod/lingo-sniper` |
| Database | Cloud SQL PostgreSQL 16 · `lingo-sniper-prod:us-central1:lingo-sniper-db` |
| TLS | cert-manager + Let's Encrypt (`letsencrypt-prod` ClusterIssuer) |
| Secrets | `kubectl create secret generic lingo-secrets` (manual, not in repo) |

### Kubernetes Manifests

```
k8s/
├── namespace.yaml
├── backend/
│   ├── deployment.yaml   # 2 replicas, RollingUpdate, Cloud SQL Proxy sidecar
│   └── service.yaml
├── frontend/
│   ├── deployment.yaml
│   └── service.yaml
├── redis/
│   └── deployment.yaml
├── ingress.yaml          # nginx-ingress + TLS + cookie affinity
├── cluster-issuer.yaml   # cert-manager ClusterIssuer
└── kustomization.yaml    # image tag management
```

### Deploy Manually

```bash
# Authenticate
gcloud auth login
gcloud container clusters get-credentials lingo-sniper-cluster \
  --zone us-central1-a --project lingo-sniper-prod

# Build and push images
docker buildx build --platform linux/amd64 \
  -t asia-southeast1-docker.pkg.dev/lingo-sniper-prod/lingo-sniper/backend:latest \
  ./backend --push

docker buildx build --platform linux/amd64 \
  -t asia-southeast1-docker.pkg.dev/lingo-sniper-prod/lingo-sniper/frontend:latest \
  ./frontend --push

# Apply with Kustomize
cd k8s
kustomize build . | kubectl apply -f -

# Wait for rollout
kubectl rollout status deployment/backend -n lingo-sniper
kubectl rollout status deployment/frontend -n lingo-sniper
```

### Cloud SQL Auth Proxy (Native Sidecar)

The proxy runs as a `restartPolicy: Always` init container (K8s 1.29+ native sidecar pattern). This guarantees:

1. Proxy starts and passes its `tcpSocket:5432` readiness probe **before** the backend container starts.
2. If the proxy crashes, K8s restarts it independently without restarting the backend.
3. The backend's startup retry loop (10 attempts × 3s) provides an additional safety net.

```yaml
initContainers:
  - name: cloud-sql-proxy
    image: gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.14.1
    restartPolicy: Always          # native sidecar
    args:
      - "--structured-logs"
      - "--auto-iam-authn"
      - "lingo-sniper-prod:us-central1:lingo-sniper-db"
    readinessProbe:
      tcpSocket:
        port: 5432
      initialDelaySeconds: 2
      periodSeconds: 5
```

### Ingress Notes

- WebSocket connections require `proxy-read-timeout: 86400` and `proxy-send-timeout: 86400` (24h).
- Cookie-based affinity (`LINGO_BACKEND`) pins each user's WS connection to a fixed pod. Redis handles cross-pod state, so a pod restart is handled gracefully.

### Secrets

```bash
kubectl create secret generic lingo-secrets \
  --namespace lingo-sniper \
  --from-literal=db-url='postgres://user:pass@127.0.0.1:5432/lingodb' \
  --from-literal=redis-url='redis://redis:6379' \
  --from-literal=jwt-secret='your-secret-here'
```

---

## CI/CD Pipeline

Two GitHub Actions workflows run in sequence on every push to `main`:

```
Push to main
    │
    ▼
┌──────────────────────────────┐
│  CI/CD Quality Gate          │  (ci.yml)
│                              │
│  backend-quality-gate        │
│  ├─ golangci-lint (ReviewDog)│
│  ├─ go test ./...            │
│  └─ 50% coverage threshold   │
│                              │
│  frontend-quality-gate       │
│  ├─ npm ci                   │
│  └─ ESLint (ReviewDog)       │
└──────────────┬───────────────┘
               │ (on success)
               ▼
┌──────────────────────────────┐
│  Deploy to GKE               │  (deploy.yml)
│                              │
│  ├─ WIF → Google Cloud auth  │
│  ├─ Docker build + push      │
│  │  (backend + frontend)     │
│  │  tagged with git SHA      │
│  ├─ kustomize edit set image │
│  ├─ kustomize build | kubectl│
│  │  apply -f -               │
│  └─ kubectl rollout status   │
└──────────────────────────────┘
```

**Authentication:** Workload Identity Federation (WIF) — no long-lived service account keys stored in GitHub. The deploy workflow exchanges a short-lived GitHub OIDC token for a Google Cloud access token.

**Image tagging:** Each deploy tags images with the git commit SHA (`github.sha`) and updates `kustomization.yaml` via `kustomize edit set image`. The registry always has a traceable build per commit.

---

## API Reference

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/auth/register` | — | Register new user |
| POST | `/api/v1/auth/login` | — | Login → returns JWT |
| GET | `/api/v1/vocab?lang=en\|zh&level=A1` | — | Vocabulary list |
| GET | `/api/v1/leaderboard?limit=10&offset=0` | — | Top players by avg reaction time |
| GET | `/api/v1/stats/me` | JWT | Authenticated player's stats |
| GET | `/api/v1/games/history` | JWT | Game history for authenticated player |
| GET | `/api/v1/online` | — | Current online player count |
| GET | `/health` | — | Liveness probe `{"status":"ok"}` |
| GET | `/ready` | — | Readiness probe (checks DB + Redis) |
| GET | `/api/v1/ws/game?token=<JWT>` | JWT (query) | Upgrade to WebSocket game connection |

---

## WebSocket Protocol

All messages are JSON. The connection URL is `wss://lingosniper.lol/api/v1/ws/game?token=<JWT>`.

### Client → Server

```jsonc
// Enter matchmaking for a 1v1 duel
{"type": "join_queue", "data": {"mode": "duel", "language": "en", "level": "B1", "quiz_type": "meaning_to_word"}}

// Solo practice session
{"type": "join_queue", "data": {"mode": "solo", "language": "zh", "level": "HSK2"}}

// Create a Battle Royale room
{"type": "create_room", "data": {"language": "en", "level": "A1"}}

// Join an existing room by code
{"type": "join_room", "data": {"room_code": "A8K3MN"}}

// Signal ready (solo/duel: triggers countdown when all ready)
{"type": "ready"}

// Host only: start Battle Royale
{"type": "start_game"}

// Submit a target hit
{"type": "target_hit", "data": {"target_id": "uuid", "reaction_ms": 342}}

// Leave current room / cancel queue
{"type": "leave_room"}
```

### Server → Client

```jsonc
// Entered matchmaking queue
{"type": "queue_joined", "data": {"status": "waiting"}}

// Match found (duel/solo) or room joined (battle)
{"type": "match_found", "data": {"room_id": "uuid", "opponent": "Player2", "mode": "duel"}}

// Battle Royale room created
{"type": "room_created", "data": {"room_code": "A8K3MN", "room_id": "uuid", "language": "en", "level": "A1"}}

// Player joined the Battle Royale room
{"type": "player_joined", "data": {"username": "Player2", "player_count": 3, "players": [...]}}

// A player left the room mid-game (Battle only)
{"type": "player_left", "data": {"username": "Player2", "player_count": 2, "players": [...]}}

// Pre-game countdown begins
{"type": "countdown", "data": {"ms": 3000}}

// New round started
{"type": "round_start", "data": {"round": 1, "total": 10, "question": "hello", "targets": [{"id":"uuid","word":"你好","x":120,"y":340}], "time_ms": 5000}}

// Score update after a hit (duel/solo)
{"type": "score_update", "data": {"you": 450, "opponent": 380, "reaction_ms": 342}}

// Live leaderboard snapshot (battle — sent after every hit)
{"type": "live_leaderboard", "data": {"round": 3, "players": [{"rank":1,"username":"P1","score":500}]}}

// Round ended (timeout or all answered)
{"type": "round_end", "data": {"result": "timeout"}}

// Game finished
{"type": "game_over", "data": {"winner": "Player1", "your_score": 450, "opponent_score": 380, "stats": {...}, "ranking": [...]}}

// Opponent disconnected (duel only)
{"type": "opponent_left"}

// Server-side error
{"type": "error", "data": "not in a room"}
```

**Quiz types:** `meaning_to_word` · `word_to_meaning` · `word_to_ipa` · `word_to_pinyin`

---

## Backend Flows

> Message types: `backend/internal/ws/message.go`  
> Game logic: `backend/internal/ws/room.go`

---

### Flow 1: Authentication (REST)

```
Client                              Server
  │                                    │
  ├─ POST /api/v1/auth/register ──────►│  Validate → bcrypt hash
  │  {username, email, password}       │  INSERT users → RETURNING created_at
  │◄─── 201 {token, user} ────────────┤  Generate JWT (HS256, 72h)
  │                                    │
  ├─ POST /api/v1/auth/login ─────────►│  FindByEmail → bcrypt.CompareHash
  │  {email, password}                 │  Generate JWT
  │◄─── 200 {token, user} ────────────┤
```

`handler/auth_handler.go` → `service/auth_service.go` → `repository/user_repo.go`

---

### Flow 2: WebSocket Connection Lifecycle

```
Client                              Server
  │                                    │
  ├─ GET /api/v1/ws/game?token=JWT ───►│  AuthWS middleware — validate JWT
  │                                    │  Upgrade HTTP → WebSocket
  │                                    │  Create Client{ID, Username, Send chan}
  │                                    │  hub.Register ← client
  │                                    │  Launch ReadPump + WritePump goroutines
  │◄─── WS connected ─────────────────┤
  │                                    │
  │  ... game messages ...             │
  │                                    │
  ├─ [connection closes] ─────────────►│  ReadPump exits
  │                                    │  hub.Unregister ← client
  │                                    │  matchmaker.Remove(client)
  │                                    │  room.RemovePlayer(client)
  │                                    │  close(client.Send)
```

Each client has two goroutines:
- **ReadPump** — reads JSON frames → `hub.HandleMessage()` → routes by `msg.Type`
- **WritePump** — drains `client.Send` channel → writes to socket; sends Ping every 54s

---

### Flow 3: Solo Practice

```
Client                  Hub                         Room
  │                      │                            │
  ├─ join_queue ─────────►│  mode == "solo"            │
  │  {mode:"solo",...}    │  GetVocabs(lang, level, 14)│
  │                       │  NewRoom(ModeSolo, vocabs) │
  │                       │  AddPlayer(client)         │
  │◄─ match_found ───────┤  {room_id, mode:"solo"}    │
  │                       │                            │
  ├─ ready ──────────────►│  SetReady → allReady()     │
  │                       │  → startGame()             │
  │◄─ countdown ─────────┤◄───────────────────────────┤
  │◄─ round_start ────────┤◄── nextRound() ────────────┤
  │  [see Round Loop]     │                            │
```

Solo creates a single-player Room with no matchmaking. `allReady()` returns true as soon as that one player sends `ready`.

---

### Flow 4: 1v1 Duel — Distributed Matchmaking

With Redis enabled (2+ backend pods), matchmaking uses an atomic Redis hash queue. Without Redis, the local in-memory queue is used.

```
Player A (Pod 1)           Redis Queue              Player B (Pod 2)
  │                            │                        │
  ├─ join_queue ──────────────►│  HGETALL — empty       │
  │                            │  HSET player_A         │
  │◄─ queue_joined ───────────┤  EXPIRE 30s             │
  │                            │                        │
  │                            │◄─── join_queue ────────┤
  │                            │  HGETALL — finds A     │
  │                            │  HDEL player_A         │
  │                            │  (atomic Lua script)   │
  │                            │                        │
  │                            │  Pod 2 creates Room    │
  │                            │  PUBLISH match_found   │
  │                            │  → Pod 1 channel       │
  │◄─ match_found ────────────►│── match_found ─────────►│
  │  {room_id, opponent:"B"}   │  {room_id, opponent:"A"}│
  │                            │                        │
  ├─ ready ───────────────────►│◄────── ready ──────────┤
  │◄─ countdown ──────────────►│── countdown ───────────►│
```

**Atomic Lua script** — `HGETALL` → find valid opponent → `HDEL` opponent → return JSON. If no opponent: `HSET` self + `EXPIRE`. Runs as a single atomic Redis command to prevent double-match races.

**Same pod:** Matched player found directly in local `pending` map.  
**Cross pod:** Matched player notified via `PUBLISH match_found` → receiving pod iterates its `hub.Clients`, proxies the room connection via `RedisProxyJoin`.

**Match conditions:** same `language` + same `level` + TTL < 30s (stale entries are skipped).

---

### Flow 5: Battle Royale (up to 100 players)

```
Host                    Hub                         Player X
  │                      │                            │
  ├─ create_room ────────►│  NewRoom(ModeBattle)       │
  │  {language, level}    │  HostID = host.ID          │
  │                       │  Code = "A8K3MN" (random)  │
  │◄─ room_created ──────┤  RoomByCode["A8K3MN"] = rm │
  │                       │                            │
  │  [host shares code]   │◄──── join_room ────────────┤
  │                       │  RoomByCode lookup         │
  │                       │  AddPlayer(X)              │
  │◄─ player_joined ─────┤── match_found ─────────────►│
  │  {player_count, [...]}│── player_joined ────────────►│
  │                       │                            │
  ├─ start_game ──────────►│  HostID check              │
  │  (host only)          │  Auto-ready all players    │
  │                       │  → startGame()             │
  │◄─ countdown ─────────┤── countdown ───────────────►│
```

Room code: 6-char alphanumeric (`ABCDEFGHJKLMNPQRSTUVWXYZ23456789`). Max players: 100. Only the host can trigger `start_game`.

---

### Flow 6: Game Round Loop (all modes)

```
Room
  │
  startGame()
  ├─ State = Countdown
  ├─ broadcast: countdown {ms: 3000}
  ├─ sleep 3s
  │
  nextRound()  (called for each of 10 rounds)
  ├─ CurrentRound++
  ├─ State = Playing
  ├─ vocabIdx = (round - 1) % len(vocabs)      ← correct answer
  ├─ generateTargets(correct)                   ← 1 correct + 3 distractors
  ├─ broadcast: round_start {round, question, targets[], time_ms: 5000}
  ├─ AfterFunc(5s) → roundTimer fires if unanswered
  │
  │  On target_hit:
  ├─ HandleHit(client, {target_id, reaction_ms})
  │  ├─ Guard: state==Playing AND not already answered
  │  ├─ Correct → score += (5000 - reactionMs) × 1000 / 5000  [min 100]
  │  ├─ Wrong   → score -= 50  [floor 0]
  │  ├─ Send: score_update
  │  │
  │  ├─ [Solo/Duel]  correct → stop timer → sleep 1s → nextRound()
  │  └─ [Battle]     all answered → stop timer → leaderboard → sleep 2s → nextRound()
  │
  │  On roundTimer:
  ├─ broadcast: round_end {result: "timeout"}
  ├─ sleep 2s → nextRound()
  │
  After round 10:
  finishGame()
  ├─ State = Finished
  ├─ sort players by score DESC → assign ranks
  ├─ Duel: detect draw
  ├─ broadcast: game_over (personalized per player)
  └─ Async: SaveGameSession → UpdateStats
```

**Scoring formula:**
```
score = (roundTimeMs - reactionMs) × 1000 / roundTimeMs
      = (5000 - reactionMs) × 1000 / 5000

Examples:
  500ms reaction  → (5000 - 500) × 1000 / 5000 = 900 pts
  2500ms reaction → (5000 - 2500) × 1000 / 5000 = 500 pts
  4900ms reaction → (5000 - 4900) × 1000 / 5000 = 20 pts → clamped to 100
```

**Target layout:** Canvas divided into 6 grid zones (shuffled each round) to prevent overlapping and make the correct target position unpredictable.

---

### Flow 7: Disconnect / Leave

```
Client                  Hub                         Room
  │                      │                            │
  ├─ leave_room ─────────►│  RemovePlayer(client)      │
  │  (voluntary)          │  matchmaker.Remove(client) │
  │                       │  client.SetRoom(nil)       │
  │                       │                            │
  ├─ [drops connection] ──►│  ReadPump exits            │
  │  (involuntary)        │  Unregister ← client       │
  │                       │  matchmaker.Remove         │
  │                       │  room.RemovePlayer         │
  │                       │                            │
  │                       │  [Duel]                    │
  │                       │  broadcast: opponent_left  │
  │                       │  stop timer → Finished     │
  │                       │                            │
  │                       │  [Battle]                  │
  │                       │  broadcast: player_left    │
  │                       │  {username, player_count}  │
```

Grace period: a player's slot stays in `room.Players` for a brief window after disconnect, so in-flight messages don't crash. `client.SendMessage` recovers from a send-on-closed-channel panic with `defer recover()`.

---

## Database

### Schema

```sql
-- Users
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(50) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    avg_reaction_ms INT DEFAULT 0,
    total_correct   INT DEFAULT 0,
    games_played    INT DEFAULT 0,
    best_reaction_ms INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Vocabulary
CREATE TABLE vocabularies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word        VARCHAR(100) NOT NULL,
    meaning     VARCHAR(255) NOT NULL,
    language    VARCHAR(5) NOT NULL,       -- 'en' | 'zh'
    level       VARCHAR(10) DEFAULT 'A1',  -- A1–B2 | HSK1–5
    difficulty  INT DEFAULT 1,             -- 1–3
    category    VARCHAR(50),
    ipa         VARCHAR(150),              -- English phonetics
    pinyin      VARCHAR(150)               -- Chinese romanization
);

-- Game sessions
CREATE TABLE game_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mode            VARCHAR(10) NOT NULL,   -- 'solo' | 'duel' | 'battle'
    language        VARCHAR(5) NOT NULL,
    winner_id       UUID REFERENCES users(id),
    rounds          INT DEFAULT 10,
    avg_reaction_ms INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    finished_at     TIMESTAMPTZ
);

-- Per-player results per session
CREATE TABLE game_session_players (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id),
    score           INT DEFAULT 0,
    avg_reaction_ms INT DEFAULT 0,
    best_reaction_ms INT DEFAULT 0,
    rank            INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

### Indexes

| Index | Columns | Type | Purpose |
|-------|---------|------|---------|
| `idx_users_email` | `users(email)` | B-tree | Login lookup |
| `idx_users_leaderboard` | `users(avg_reaction_ms ASC) WHERE games_played > 0 AND avg_reaction_ms > 0` | Partial | Leaderboard query — index-only scan, no heap fetch |
| `idx_vocabularies_level` | `vocabularies(language, level)` | B-tree | Vocab fetch by language + level |
| `idx_gsp_session` | `game_session_players(session_id)` | B-tree | JOIN from sessions side |
| `idx_gsp_user_session` | `game_session_players(user_id, session_id)` | Covering | Game history query — satisfies WHERE and JOIN without heap fetch |
| `idx_game_sessions_created` | `game_sessions(created_at DESC)` | B-tree | Recent sessions ordering |

The leaderboard uses a window function to fetch count and rows in a single query:

```sql
SELECT id, username, avg_reaction_ms, total_correct, games_played, best_reaction_ms,
       COUNT(*) OVER() AS total_count
FROM users
WHERE games_played > 0 AND avg_reaction_ms > 0
ORDER BY avg_reaction_ms ASC
LIMIT $1 OFFSET $2
```

---

## Environment Variables

### Backend

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server listen port |
| `DB_URL` | `postgres://lingouser:lingopass@localhost:5432/lingodb?sslmode=disable` | PostgreSQL DSN |
| `REDIS_URL` | `redis://localhost:6379` | Redis URL — optional, falls back to single-instance mode |
| `JWT_SECRET` | `dev-secret-change-in-prod` | JWT HMAC signing key |
| `JWT_EXPIRATION` | `72h` | Token lifetime |
| `LOG_LEVEL` | `info` | `debug` · `info` · `warn` · `error` |
| `LOG_FORMAT` | `json` | `json` (production) · `text` (development) |
| `GIN_MODE` | `debug` | `release` in production |
| `CORS_ORIGINS` | `http://localhost:3000` | Comma-separated allowed HTTP origins |
| `ALLOWED_WS_ORIGINS` | `http://localhost:3000` | Comma-separated allowed WebSocket origins |

### Frontend

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost/api/v1` | Backend API base URL |
| `NEXT_PUBLIC_WS_URL` | `ws://localhost/api/v1/ws/game` | WebSocket endpoint |

---

## Project Structure

```
language-arena/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go        # Entry point — router, migrations, graceful shutdown
│   │   └── migrate/main.go       # Standalone migration CLI (for Docker / DevOps)
│   ├── internal/
│   │   ├── config/               # Environment config loader
│   │   ├── model/                # Domain types (GameMode, QuizType, User, Vocab…)
│   │   ├── repository/           # PostgreSQL data access (user, vocab, game)
│   │   ├── service/              # Business logic (auth, vocab, leaderboard)
│   │   ├── handler/              # HTTP handlers (auth, vocab, leaderboard, game/ws)
│   │   ├── ws/
│   │   │   ├── hub.go            # Central message router + client registry
│   │   │   ├── room.go           # Game loop (rounds, scoring, timers)
│   │   │   ├── client.go         # WS client — ReadPump / WritePump
│   │   │   ├── matchmaker.go     # Duel queue — local + Redis distributed
│   │   │   ├── redis_adapter.go  # Redis queue, Pub/Sub relay, atomic Lua script
│   │   │   └── message.go        # WS message type constants
│   │   ├── middleware/           # Auth, CORS, rate limiter, request logger, locale
│   │   └── migration/            # SQL migration files (go:embed)
│   │       ├── 001_schema.sql
│   │       ├── 002_seed.sql
│   │       ├── 002_reaction_scoring.sql
│   │       ├── 003_backfill_reaction.sql
│   │       └── 004_indexes.sql
│   └── pkg/
│       ├── logger/               # slog initializer (JSON/text, level)
│       └── response/             # Standardised HTTP response helpers
│
├── frontend/
│   └── src/
│       ├── app/                  # Next.js App Router pages
│       ├── components/           # UI components + game canvas
│       ├── hooks/                # useAuth, useWebSocket, useGame
│       └── lib/                  # API client, type definitions
│
├── k8s/
│   ├── namespace.yaml
│   ├── backend/deployment.yaml   # 2 replicas, Cloud SQL Proxy native sidecar
│   ├── backend/service.yaml
│   ├── frontend/deployment.yaml
│   ├── frontend/service.yaml
│   ├── redis/deployment.yaml
│   ├── ingress.yaml              # nginx + TLS + cookie affinity
│   ├── cluster-issuer.yaml
│   └── kustomization.yaml
│
├── .github/workflows/
│   ├── ci.yml                    # Quality gate — lint + test + coverage
│   └── deploy.yml                # Build → push → GKE deploy (WIF auth)
│
├── nginx/
│   └── default.conf              # Local dev reverse proxy
│
└── docker-compose.yml            # Local full-stack deployment
```

---
`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` (production) or `text` (development) |
| `GIN_MODE` | `debug` | `debug` or `release` |

## License

MIT
