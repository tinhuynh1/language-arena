# 🏗️ System Design — Lingo Sniper

> Tổng quan kiến trúc hệ thống / System architecture overview

## Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        Browser["Next.js Browser App<br/>(React + Tailwind)"]
    end

    subgraph "Edge / Ingress"
        Nginx["Nginx Ingress Controller<br/>TLS termination + WS upgrade"]
    end

    subgraph "Application Layer (GKE)"
        subgraph "Pod 1"
            BE1["Go Backend"]
            CSP1["Cloud SQL Proxy"]
        end
        subgraph "Pod 2"
            BE2["Go Backend"]
            CSP2["Cloud SQL Proxy"]
        end
        FE["Next.js SSR<br/>(standalone)"]
    end

    subgraph "Data Layer"
        PG["PostgreSQL 16<br/>(Cloud SQL)"]
        Redis["Redis 7<br/>(session affinity + pub/sub)"]
    end

    Browser -->|"HTTPS / WSS"| Nginx
    Nginx -->|"/api/*"| BE1
    Nginx -->|"/api/*"| BE2
    Nginx -->|"/*"| FE
    BE1 <-->|"Pub/Sub"| Redis
    BE2 <-->|"Pub/Sub"| Redis
    BE1 --> CSP1 --> PG
    BE2 --> CSP2 --> PG
```

## Tech Stack

| Layer | Technology | Lý do chọn / Why |
|-------|-----------|-------------------|
| **Frontend** | Next.js 15 + React 19 + Tailwind CSS v4 | SSR, file-based routing, rapid UI dev |
| **Backend** | Go 1.25 + Gin framework | High concurrency, low latency for real-time game |
| **WebSocket** | gorilla/websocket + Redis Pub/Sub | Real-time bidirectional communication, multi-pod sync |
| **Database** | PostgreSQL 16 (Cloud SQL) | Relational data, ACID, leaderboard queries |
| **Cache/PubSub** | Redis 7 | Cross-pod WS message relay, session data |
| **Container** | Docker + GKE Autopilot | Managed K8s, auto-scaling |
| **CI/CD** | GitHub Actions + WIF | Keyless auth to GCP, auto-deploy on push to `main` |
| **Ingress** | Nginx Ingress + cert-manager | TLS via Let's Encrypt, WebSocket upgrade support |

## Backend Architecture (Layered)

```mermaid
graph LR
    HTTP["HTTP Request"] --> MW["Middleware<br/>CORS, Auth, RateLimit,<br/>RequestID, Logger"]
    MW --> H["Handler<br/>auth, vocab,<br/>leaderboard, game"]
    H --> S["Service<br/>business logic,<br/>interfaces"]
    S --> R["Repository<br/>SQL queries,<br/>DB access"]
    R --> DB["PostgreSQL"]

    WS["WebSocket"] --> Hub["Hub<br/>message router"]
    Hub --> Room["Room<br/>game state machine"]
    Hub --> MM["Matchmaker<br/>auto-pairing"]
    Hub --> RA["Redis Adapter<br/>cross-pod sync"]
    Room --> R
```

> **Nguyên tắc / Principle**: Mỗi layer chỉ giao tiếp với layer ngay dưới nó. Handler không truy cập trực tiếp Repository. / Each layer only communicates with the layer directly below it.

## Frontend Architecture

```mermaid
graph LR
    Pages["Pages<br/>/, /play, /login,<br/>/leaderboard, /dashboard"]
    Pages --> Components["Components<br/>GameCanvas, Countdown,<br/>GameOverScreen, LiveLeaderboard"]
    Pages --> Hooks["Hooks<br/>useAuth, useGame,<br/>useWebSocket"]
    Hooks --> API["lib/api.ts<br/>axios instance"]
    Hooks --> WS["WebSocket<br/>connection"]
    API --> Backend["Go Backend"]
    WS --> Backend
```

## Data Flow

### REST API Flow (Ví dụ: GET /api/v1/leaderboard)

```
Browser → Nginx → [RequestID] → [Logger] → [CORS] → [RateLimit]
        → LeaderboardHandler.GetLeaderboard()
        → LeaderboardService.GetTopPlayers()
        → UserRepository.GetLeaderboard()
        → PostgreSQL → Response
```

### WebSocket Flow (Ví dụ: Player joins game)

```
Browser → WSS upgrade → GameHandler.HandleWebSocket()
        → JWT validation → User lookup
        → Client created → Hub.Register
        → Client sends "join_queue"
        → Hub.HandleMessage() → Matchmaker.Enqueue()
        → Match found → Room created → "match_found" sent to both players
```

## Database Schema

```mermaid
erDiagram
    USERS {
        uuid id PK
        string username UK
        string email UK
        string password_hash
        bigint total_score
        int games_played
        int best_reaction_ms
        timestamp created_at
    }

    GAME_SESSIONS {
        uuid id PK
        string mode
        string language
        int rounds
        int avg_reaction_ms
        uuid winner_id FK
        timestamp created_at
        timestamp finished_at
    }

    GAME_SESSION_PLAYERS {
        uuid id PK
        uuid session_id FK
        uuid user_id FK
        int score
        int avg_reaction_ms
        int best_reaction_ms
        int rank
    }

    VOCABULARIES {
        uuid id PK
        string word
        string meaning
        string language
        string level
        int difficulty
        string category
        string ipa
        string pinyin
    }

    USERS ||--o{ GAME_SESSION_PLAYERS : "plays"
    GAME_SESSIONS ||--o{ GAME_SESSION_PLAYERS : "has"
    USERS ||--o{ GAME_SESSIONS : "wins"
```

## Deployment Topology

```
GKE Cluster (asia-southeast1-a)
├── Namespace: lingo-sniper
│   ├── Deployment: backend (2 replicas)
│   │   ├── Container: go-backend (:8080)
│   │   └── Sidecar: cloud-sql-proxy
│   ├── Deployment: frontend (1 replica)
│   │   └── Container: nextjs (:3000)
│   ├── Deployment: redis (1 replica)
│   │   └── Container: redis (:6379)
│   ├── Service: backend (ClusterIP)
│   ├── Service: frontend (ClusterIP)
│   ├── Service: redis (ClusterIP)
│   └── Ingress: nginx + cert-manager (Let's Encrypt)
└── Domain: lingosniper.lol
```

## Key Design Decisions

| Decision | Choice | Alternative Considered | Lý do / Reason |
|----------|--------|----------------------|----------------|
| Monolith vs Microservice | **Monolith** | Microservice | Small team, game logic tightly coupled |
| WS library | **gorilla/websocket** | nhooyr/websocket | Battle-tested, simpler API |
| Multi-pod WS sync | **Redis Pub/Sub** | NATS, Kafka | Already using Redis, sufficient for scale |
| Auth strategy | **JWT stateless** | Session-based | Horizontal scaling, no shared session store |
| DB migration | **Embedded go:embed** | Flyway, golang-migrate CLI | Zero external dependencies at runtime |
