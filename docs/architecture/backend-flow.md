# 🔄 Backend Flow — Lingo Sniper

> Chi tiết luồng xử lý backend / Detailed backend request lifecycle

## Layered Architecture

```
cmd/server/main.go          ← Entry point, DI, server setup
├── internal/middleware/     ← Cross-cutting concerns (CORS, auth, rate limit, logging)
├── internal/handler/        ← HTTP transport layer (request parsing, response formatting)
├── internal/service/        ← Business logic (interfaces, validation)
├── internal/repository/     ← Data access (SQL queries, DB interaction)
├── internal/model/          ← Domain models (User, GameSession, Vocabulary)
├── internal/config/         ← Environment configuration
├── internal/ws/             ← WebSocket engine (Hub, Room, Client, Matchmaker)
├── internal/migration/      ← SQL migration files (embedded via go:embed)
└── pkg/                     ← Shared packages (logger, response)
```

## Startup Flow / Luồng khởi động

```mermaid
sequenceDiagram
    participant Main as main.go
    participant Config as config.Load()
    participant Logger as logger.Setup()
    participant DB as PostgreSQL
    participant Migration as runMigrations()
    participant Hub as ws.Hub
    participant Router as gin.Engine
    participant Server as http.Server

    Main->>Config: Load env vars
    Config->>Config: Validate (panic if JWT_SECRET default in release)
    Main->>Logger: Setup slog (JSON/Text, log level)
    Main->>DB: sql.Open() + Ping
    Main->>Migration: Run embedded SQL migrations
    Main->>Hub: NewHub() + go hub.Run()
    Main->>Router: setupRouter() middleware + routes
    Main->>Server: ListenAndServe(:8080)
    Note over Server: Graceful shutdown on SIGINT/SIGTERM
```

## REST API Request Lifecycle / Vòng đời request

```mermaid
sequenceDiagram
    participant Client as Browser
    participant MW as Middleware Chain
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant DB as PostgreSQL

    Client->>MW: GET /api/v1/leaderboard?limit=10
    MW->>MW: RequestID() → inject X-Request-ID
    MW->>MW: RequestLogger() → log start
    MW->>MW: CORSMiddleware() → validate origin
    MW->>MW: RateLimiter() → check + set X-RateLimit-* headers
    MW->>H: LeaderboardHandler.GetLeaderboard()
    H->>H: Parse & validate query params
    H->>S: LeaderboardService.GetTopPlayers(ctx, 10)
    S->>S: Validate limit bounds (0 < limit ≤ 100)
    S->>R: UserRepository.GetLeaderboard(ctx, 10)
    R->>DB: SELECT ... ORDER BY total_score DESC LIMIT 10
    DB-->>R: rows
    R->>R: Log if slow (>100ms)
    R-->>S: []LeaderboardEntry
    S-->>H: []LeaderboardEntry
    H-->>Client: {"success": true, "data": [...]}
    MW->>MW: RequestLogger() → log complete (status, latency, bytes)
```

## Middleware Chain / Chuỗi Middleware

Thứ tự thực thi (trên xuống) / Execution order (top to bottom):

| # | Middleware | Chức năng / Purpose |
|---|-----------|---------------------|
| 1 | `gin.Recovery()` | Recover panics, trả 500 thay vì crash |
| 2 | `RequestID()` | Inject unique ID vào context + response header |
| 3 | `RequestLogger()` | Log method, path, status, latency, bytes, request_id |
| 4 | `CORSMiddleware(origins)` | Validate browser origin từ config |
| 5 | `RateLimiter(200/min)` | IP-based rate limiting + RFC 6585 headers |
| 6 | `AuthMiddleware()` | JWT validation (chỉ cho protected routes) |

## WebSocket Game Flow / Luồng game real-time

### Connection Establishment / Thiết lập kết nối

```mermaid
sequenceDiagram
    participant B as Browser
    participant GH as GameHandler
    participant Auth as AuthService
    participant Hub as Hub
    participant C as Client

    B->>GH: GET /api/v1/ws/game?token=xxx
    GH->>Auth: ValidateToken(token)
    Auth-->>GH: userID
    GH->>GH: userRepo.FindByID(userID)
    GH->>GH: upgrader.Upgrade() (check origin)
    GH->>C: NewClient(hub, conn, userID, username)
    GH->>Hub: hub.Register <- client
    C->>C: go WritePump() // goroutine gửi message
    C->>C: go ReadPump()  // goroutine nhận message
```

### Game State Machine / Máy trạng thái game

```mermaid
stateDiagram-v2
    [*] --> Waiting: Room created
    Waiting --> Playing: Host starts game<br/>(all players ready)
    Playing --> Playing: Round cycle<br/>(round_start → target_hit → round_end)
    Playing --> Finished: All rounds complete
    Finished --> [*]: Results saved to DB

    Waiting --> [*]: All players leave
    Playing --> [*]: All players disconnect
```

### Room Lifecycle / Vòng đời phòng chơi

```
1. CREATE_ROOM → Room created (code: "ABC123") → Host joins
2. JOIN_ROOM   → Player joins by code | Matchmaker auto-pairs
3. READY       → Each player marks ready
4. START_GAME  → Host triggers (solo: auto-start)
5. COUNTDOWN   → 3...2...1... broadcast
6. ROUND_START → Question + targets sent to all players
7. TARGET_HIT  → Player clicks → score calculated → broadcast update
8. ROUND_END   → After timeout → live leaderboard shown
9. → Repeat steps 6-8 for all rounds
10. GAME_OVER  → Final ranking → results saved to DB → stats updated
```

### Cross-Instance Flow (Redis Pub/Sub) / Đồng bộ đa pod

```mermaid
sequenceDiagram
    participant P1 as Player A (Pod 1)
    participant Hub1 as Hub (Pod 1)
    participant Redis as Redis Pub/Sub
    participant Hub2 as Hub (Pod 2)
    participant P2 as Player B (Pod 2)

    Note over Hub1: Room "ABC" lives on Pod 1
    P2->>Hub2: join_room "ABC"
    Hub2->>Redis: LookupRoom("ABC") → Pod 1
    Hub2->>Redis: Publish: proxy_join to Pod 1
    Redis->>Hub1: proxy_join received
    Hub1->>Hub1: Create ProxyClient for Player B
    Hub1->>Redis: Relay: match_found to Pod 2
    Redis->>Hub2: match_found received
    Hub2->>P2: Forward WS message

    Note over P2: Player B hits a target
    P2->>Hub2: target_hit
    Hub2->>Redis: proxy_action to Pod 1
    Redis->>Hub1: Process hit on real Room
    Hub1->>Redis: Relay: score_update to Pod 2
    Redis->>Hub2: Forward to Player B
    Hub1->>P1: Direct WS: score_update
```

## Logging Strategy / Chiến lược logging

Mỗi layer có component tag riêng. Khi incident xảy ra, search bằng `request_id` để thấy toàn bộ call chain:

```
Layer           │ Component Tag         │ Log Level
────────────────┼───────────────────────┼──────────────────
Middleware      │ HTTP                  │ INFO (all requests)
Handler         │ HANDLER.Auth/Game/... │ WARN (client err), ERROR (server err)
Service         │ SVC.Auth/Vocab/...    │ INFO (lifecycle), ERROR (failures)
Repository      │ REPO.User/Game/...    │ ERROR (query fail), WARN (slow >100ms)
WebSocket       │ WS                    │ INFO (connect/disconnect), ERROR (upgrade fail)
Redis           │ REDIS                 │ INFO (sync events), ERROR (connection)
```

## Error Handling Pattern / Quy ước xử lý lỗi

```go
// Repository: log error + return
if err != nil {
    r.log.Error("find user failed", "op", "FindByID", "user_id", id, "err", err)
    return nil, err
}

// Service: log context + return
if err != nil {
    s.log.Error("get top players failed", "op", "GetTopPlayers", "err", err)
    return nil, err
}

// Handler: log + return generic message to client (NEVER expose internal errors)
if err != nil {
    h.log.Error("get leaderboard failed", "err", err, "request_id", reqID)
    response.InternalError(c, "failed to fetch leaderboard") // Generic!
    return
}
```

> **Nguyên tắc / Principle**: Internal errors NEVER leak to client. Client chỉ nhận generic error message. Detail nằm trong structured logs.
