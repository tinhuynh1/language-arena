# 📘 Learning Guide: Go Backend Development

> Hướng dẫn học Go backend qua dự án Lingo Sniper / Learn Go backend through this project

## Prerequisites / Kiến thức cần có

- Basic programming (biến, hàm, vòng lặp)
- HTTP basics (request, response, status codes)
- SQL basics (SELECT, INSERT, JOIN)

---

## 1. Go Fundamentals / Cơ bản Go

### Tài liệu học / Resources

| Resource | Type | Link |
|----------|------|------|
| Go Tour | Interactive | https://go.dev/tour |
| Go by Example | Examples | https://gobyexample.com |
| Effective Go | Docs | https://go.dev/doc/effective_go |

### Khái niệm chính trong dự án / Key concepts used

| Concept | File ví dụ / Example | Giải thích / Why |
|---------|---------------------|------------------|
| **Structs & Methods** | `repository/user_repo.go` | Mọi layer dùng struct + method receivers |
| **Interfaces** | `service/auth_service.go` | `UserReader`, `UserWriter` cho Dependency Inversion |
| **Goroutines** | `ws/client.go` | `go client.ReadPump()` — concurrent WS read/write |
| **Channels** | `ws/hub.go` | `Register`, `Unregister` channels cho client management |
| **Context** | `handler/`, `service/`, `repository/` | Request-scoped data + cancellation |
| **Error handling** | Everywhere | `if err != nil { return err }` pattern |
| **Embedding (go:embed)** | `migration/embed.go` | SQL files embedded vào binary |

### Bài tập / Exercises

1. ✏️ Đọc `cmd/server/main.go` → vẽ sơ đồ DI (dependency injection)
2. ✏️ Trace 1 request từ `GET /api/v1/leaderboard` qua tất cả layers
3. ✏️ Tìm tất cả goroutines trong project → giải thích mỗi cái làm gì

---

## 2. Gin Framework / Framework Gin

### Tài liệu / Resources
- Official docs: https://gin-gonic.com/docs/
- Xem file: `cmd/server/main.go` → `setupRouter()`

### Patterns used / Các pattern sử dụng

```go
// Route grouping — nhóm routes
api := r.Group("/api/v1")
auth := api.Group("/auth")
protected := api.Group("").Use(middleware.AuthMiddleware(authService))

// Middleware chain — chuỗi middleware
r.Use(gin.Recovery())
r.Use(middleware.RequestID())
r.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))

// Handler pattern — xử lý endpoint
func (h *AuthHandler) Login(c *gin.Context) {
    var req model.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil { ... }
    result, err := h.authService.Login(c.Request.Context(), req)
    response.OK(c, result)
}

// Context values — truyền data qua middleware
c.Set("user_id", userID)           // Set
userID := c.Get("user_id")         // Get
ctx := c.Request.Context()         // Access Go context
```

---

## 3. Structured Logging with slog

### Tài liệu / Resources
- Official: https://pkg.go.dev/log/slog
- Xem file: `pkg/logger/logger.go`, `middleware/request_logger.go`

### Pattern trong dự án / Project patterns

```go
// Component-tagged logger — logger với tag component
log := slog.Default().With("component", "REPO.User")

// Structured key-value logging
log.Error("query failed", "op", "FindByID", "user_id", id, "err", err, "duration_ms", 150)

// Output (JSON):
// {"level":"ERROR","component":"REPO.User","op":"FindByID","user_id":"abc-123","err":"connection refused","duration_ms":150}
```

### Quy ước log level / Log level conventions
- `Debug`: Token validation, successful fetches (chỉ hiện khi `LOG_LEVEL=debug`)
- `Info`: Business events (login, register, room created, game over)
- `Warn`: Client errors, slow queries (>100ms)
- `Error`: Server errors, DB failures

---

## 4. Database with database/sql

### Tài liệu / Resources
- https://go.dev/doc/database/
- Xem file: `repository/user_repo.go`

### Patterns

```go
// QueryRow — single row
err := r.db.QueryRowContext(ctx, "SELECT ... WHERE id = $1", id).Scan(&user.ID, &user.Name)

// Query — multiple rows
rows, err := r.db.QueryContext(ctx, "SELECT ... LIMIT $1", limit)
defer rows.Close()
for rows.Next() { rows.Scan(&entry.ID, ...) }

// Exec — write operations
_, err := r.db.ExecContext(ctx, "UPDATE users SET score = $2 WHERE id = $1", id, score)

// Connection pooling (set in main.go)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## 5. Concurrency Patterns / Các pattern concurrency

| Pattern | File | Mô tả / Description |
|---------|------|---------------------|
| **Fan-out goroutines** | `client.go` | 2 goroutines per connection (read + write) |
| **Channel-based coordination** | `hub.go` | Register/Unregister channels |
| **Mutex protection** | `hub.go`, `room.go` | `sync.RWMutex` for shared state (Clients, Rooms) |
| **Select statement** | `hub.go:Run()` | Listen on multiple channels simultaneously |
| **Graceful shutdown** | `main.go` | `signal.Notify` + context timeout |

### Bài tập / Exercises

1. ✏️ Tại sao `hub.Run()` dùng `select` thay vì `if`?
2. ✏️ Nếu bỏ `defer rows.Close()` thì điều gì xảy ra?
3. ✏️ Giải thích tại sao `Client.Send` là buffered channel (`make(chan []byte, 256)`)
