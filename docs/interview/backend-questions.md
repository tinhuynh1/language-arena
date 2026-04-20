# 🎯 Interview Questions: Backend

> Câu hỏi phỏng vấn backend qua dự án / Backend interview prep using this project

---

## Go Fundamentals

### Q1: Tại sao chọn Go cho dự án này thay vì Node.js hay Python?

**Answer:**
- **Concurrency model**: Goroutines (lightweight threads, ~2KB stack) + channels cho WebSocket game server — xử lý hàng nghìn concurrent connections hiệu quả
- **Performance**: Compiled binary, không có garbage collection pause lớn như JVM
- **Single binary deployment**: `go build` → 1 file chạy được, Docker image chỉ ~15MB
- **Standard library mạnh**: `net/http`, `database/sql`, `log/slog` — ít dependency ngoài
- 📄 Xem: `backend/cmd/server/main.go`, `backend/internal/ws/hub.go`

### Q2: Giải thích Interface trong Go và cách dùng trong dự án?

**Answer:**
- Go interface là implicit (duck typing) — struct tự động implement interface nếu có đủ methods
- Dự án dùng Interface Segregation: `UserReader`, `UserWriter` thay vì 1 interface lớn `UserRepository`
- Benefit: Service layer KHÔNG depend vào concrete repository → dễ mock khi test
```go
type UserReader interface {
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}
```
- 📄 Xem: `service/auth_service.go`, `service/leaderboard_service.go`

### Q3: Goroutine và Channel hoạt động thế nào?

**Answer:**
- Goroutine: Lightweight thread do Go runtime quản lý, khởi tạo bằng `go func()`
- Channel: Pipe để goroutines giao tiếp an toàn (synchronized)
- Trong dự án:
  - Mỗi WS client = 2 goroutines (`ReadPump`, `WritePump`)
  - Hub dùng channels: `Register`, `Unregister` để quản lý clients
  - `Client.Send` là buffered channel (`make(chan []byte, 256)`)
- 📄 Xem: `ws/client.go`, `ws/hub.go`

---

## Architecture

### Q4: Giải thích kiến trúc phân tầng (Layered Architecture)?

**Answer:**
```
Handler → Service → Repository → Database
```
- **Handler**: Parse request, validate input, format response. KHÔNG chứa business logic
- **Service**: Business logic, validation rules, orchestration. Depend vào interfaces
- **Repository**: SQL queries, data mapping. KHÔNG biết về HTTP
- **Tại sao**: Separation of Concerns — mỗi layer chỉ 1 trách nhiệm, dễ test, dễ thay đổi
- 📄 Xem: `handler/leaderboard_handler.go` → `service/leaderboard_service.go` → `repository/user_repo.go`

### Q5: Middleware chain hoạt động thế nào?

**Answer:**
- Gin middleware = chain of handlers, mỗi handler gọi `c.Next()` để pass sang handler tiếp
- Thứ tự: Recovery → RequestID → Logger → CORS → RateLimit → [Route Handler]
- RequestID inject vào context, tất cả layers sau đó truy cập được
- Logger chạy 2 lần: trước (start timer) và sau (log result)
- 📄 Xem: `middleware/request_logger.go`, `cmd/server/main.go:setupRouter()`

### Q6: Tại sao dùng Dependency Injection thay vì global variables?

**Answer:**
- **Testability**: Mock dependencies dễ dàng (`UserReader` interface → mock in tests)
- **Explicit dependencies**: Nhìn vào constructor biết ngay cần gì
- **No hidden state**: Không có global variable bất ngờ thay đổi
- Pattern: Constructor injection via `New*()` functions
```go
func NewAuthService(reader UserReader, writer UserWriter, cfg *config.JWTConfig) *AuthService
```

---

## Database

### Q7: Tại sao dùng `database/sql` thay vì ORM (GORM)?

**Answer:**
- **Performance**: Truy vấn trực tiếp SQL, không overhead ORM
- **Control**: Biết chính xác query nào được chạy
- **Learning**: Hiểu sâu SQL thay vì rely trên abstraction
- **Complexity**: Dự án không có schema phức tạp → ORM overkill
- Trade-off: Boilerplate nhiều hơn (scan manually)

### Q8: Giải thích Connection Pooling?

**Answer:**
```go
db.SetMaxOpenConns(25)        // Max 25 concurrent connections
db.SetMaxIdleConns(10)        // Keep 10 idle connections ready
db.SetConnMaxLifetime(5*min)  // Recycle connections after 5 min
```
- **Tại sao**: Opening DB connection tốn ~50ms. Pool giữ connections sẵn sàng
- **MaxOpen**: Limit tổng connections đến DB (tránh overload)
- **MaxIdle**: Giữ connections "warm" cho request tiếp theo
- **Lifetime**: Tránh stale connections

### Q9: Giải thích embedded migrations (go:embed)?

**Answer:**
```go
//go:embed *.sql
var Files embed.FS
```
- SQL migration files compiled vào binary
- Không cần COPY migration files trong Dockerfile
- Không cần external tool (Flyway, golang-migrate CLI)
- Schema versioned cùng application code

---

## Observability

### Q10: Logging strategy trong production?

**Answer:**
- **Structured JSON logging** with `slog` — mỗi log entry là JSON parseable
- **Component tags**: `REPO.User`, `SVC.Auth`, `HANDLER.Leaderboard`
- **request_id propagation**: Inject ở middleware → truyền qua context → mọi layer log cùng ID
- **Log levels**: Debug (token), Info (business), Warn (client err/slow query), Error (server)
- **Full trace**: Search `request_id=xyz` → thấy call chain Handler → Service → Repo
- 📄 Xem: `middleware/request_logger.go`, `repository/user_repo.go`
