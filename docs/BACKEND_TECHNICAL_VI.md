# Tài Liệu Kỹ Thuật Backend — Lingo Sniper

> Mô tả chi tiết cách hoạt động từng luồng xử lý phía backend theo hướng kỹ thuật.

---

## 1. Khởi Động Server

### Call chain

```
main() → config.Load() → sql.Open() → db.Ping() → runMigrations()
       → repository.New*() → service.New*() → ws.NewHub()
       → go hub.Run()
       → setupRouter() → http.Server.ListenAndServe()
       → signal.Notify(SIGINT/SIGTERM) → srv.Shutdown()
```

### Chi tiết

**`cmd/server/main.go`**

1. `config.Load()` đọc biến môi trường (`DB_HOST`, `DB_PORT`, `JWT_SECRET`, ...) rồi trả về struct `config.Config`.
2. Mở connection pool tới PostgreSQL bằng `database/sql`. Cấu hình pool:
   - `MaxOpenConns` — số connection tối đa đồng thời
   - `MaxIdleConns` — số connection rảnh giữ lại
   - `ConnMaxLifetime` — thời gian tối đa mỗi connection tồn tại
3. `runMigrations(db)` chạy tuần tự 6 câu SQL `CREATE TABLE IF NOT EXISTS` + `CREATE INDEX IF NOT EXISTS`. Nếu bảng đã tồn tại thì idempotent (không lỗi).
4. Khởi tạo 3 tầng theo thứ tự dependency:
   ```
   Repository (truy vấn DB) → Service (logic nghiệp vụ) → Handler (HTTP endpoint)
   ```
5. `ws.NewHub(vocabService)` tạo Hub quản lý WebSocket. `go hub.Run()` chạy event loop trong goroutine riêng.
6. `setupRouter()` đăng ký các route với Gin, gắn middleware CORS và Rate Limiter.
7. Graceful shutdown: lắng nghe `SIGINT`/`SIGTERM`, gọi `srv.Shutdown(ctx)` với timeout 10s.

### Database schema (tạo bởi `runMigrations`)

| Bảng | Cột chính | Index |
|---|---|---|
| `users` | `id` (UUID PK), `username` (UNIQUE), `email` (UNIQUE), `password_hash`, `total_score`, `games_played`, `best_reaction_ms` | `idx_users_email`, `idx_users_total_score DESC` |
| `vocabularies` | `id` (UUID PK), `word`, `meaning`, `language`, `level`, `difficulty` (1-3), `category` | `idx_vocabularies_language`, `idx_vocabularies_level(language, level)` |
| `game_sessions` | `id` (UUID PK), `mode` (CHECK: solo/duel/battle), `language`, `player1_id` (FK users), `player2_id`, `winner_id`, `rounds`, `avg_reaction_ms` | — |

---

## 2. Đăng Ký Tài Khoản

### Endpoint
```
POST /api/v1/auth/register
Content-Type: application/json
Body: {"username": "abc", "email": "abc@mail.com", "password": "123456"}
```

### Call chain
```
AuthHandler.Register()
  → c.ShouldBindJSON(&req)           // Gin binding, validate: username min=3 max=50, email format, password min=6
  → authService.Register(ctx, req)
    → userRepo.FindByEmail(ctx, email)  // SELECT * FROM users WHERE email = $1
    → bcrypt.GenerateFromPassword(password, DefaultCost=10)  // hash password, cost=10 → ~100ms
    → userRepo.Create(ctx, user)        // INSERT INTO users (username, email, password_hash) VALUES (...)
    → authService.generateToken(userID)
      → jwt.NewWithClaims(HS256, {user_id, exp: now+72h, iat: now})
      → token.SignedString([]byte(jwt_secret))
  → response.Created(c, {token, user})  // HTTP 201
```

### Luồng dữ liệu
```
Request JSON → model.RegisterRequest (Gin validate bằng struct tag `binding`)
             → bcrypt hash (10 rounds, ~100ms)
             → INSERT PostgreSQL (gen_random_uuid() cho id)
             → JWT signed với HS256 (claim: user_id + exp)
             → Response JSON {token: "eyJ...", user: {id, username, email, ...}}
```

### Xử lý lỗi
- `userRepo.FindByEmail` trả về user → return `ErrUserExists` → HTTP 400
- `userRepo.Create` lỗi duplicate username (UNIQUE constraint) → HTTP 400
- Bất kỳ lỗi nào khác → HTTP 500 `"registration failed"`

---

## 3. Đăng Nhập

### Endpoint
```
POST /api/v1/auth/login
Body: {"email": "abc@mail.com", "password": "123456"}
```

### Call chain
```
AuthHandler.Login()
  → authService.Login(ctx, req)
    → userRepo.FindByEmail(ctx, email)                          // SELECT * FROM users WHERE email = $1
    → bcrypt.CompareHashAndPassword(user.PasswordHash, password) // so sánh hash, ~100ms
    → authService.generateToken(user.ID)                         // JWT HS256
  → response.OK(c, {token, user})                               // HTTP 200
```

### Xử lý lỗi
- Email không tồn tại → `ErrInvalidCredentials` → HTTP 401
- Password sai (bcrypt compare fail) → `ErrInvalidCredentials` → HTTP 401

---

## 4. Xác Thực JWT (Middleware)

### 2 cơ chế xác thực

**a) REST API — `middleware.AuthMiddleware()`** (`middleware/auth.go`)
```
Request → Đọc header "Authorization: Bearer <token>"
        → SplitN(" ", 2) → lấy phần token
        → authService.ValidateToken(token)
          → jwt.Parse(token, func → kiểm tra SigningMethodHMAC)
          → Đọc claim "user_id" từ MapClaims
          → uuid.Parse(user_id_string)
        → c.Set("user_id", userID)  // gắn vào Gin context
        → c.Next()
```

**b) WebSocket — inline trong `GameHandler.HandleWebSocket()`** (`handler/game_handler.go`)
```
GET /api/v1/ws/game?token=<JWT>
  → Đọc query param "token" (hoặc fallback header "Authorization")
  → authService.ValidateToken(token) → userID
  → userRepo.FindByID(ctx, userID)   → user struct
  → upgrader.Upgrade(c.Writer, c.Request, nil)  // HTTP → WS upgrade
  → ws.NewClient(hub, conn, user.ID, user.Username)
  → hub.Register <- client           // gửi vào channel
  → go client.WritePump()            // goroutine ghi
  → go client.ReadPump()             // goroutine đọc
```

**Tại sao WS không dùng middleware?** Vì WebSocket upgrade cần xảy ra bên trong handler. Middleware Gin sẽ trả HTTP response trước khi upgrade, làm hỏng handshake.

---

## 5. Kết Nối WebSocket — Client Lifecycle

### Struct `Client` (`ws/client.go`)
```go
type Client struct {
    ID       uuid.UUID        // user ID từ DB
    Username string
    Hub      *Hub
    Conn     *websocket.Conn  // gorilla/websocket connection
    Send     chan []byte       // buffered channel, capacity=256
    Room     *Room
    mu       sync.Mutex       // bảo vệ trường Room
}
```

### ReadPump (goroutine 1)
```
Loop vô hạn:
  conn.ReadMessage()
    → Nếu error (close/timeout) → break khỏi loop → defer: hub.Unregister <- client
    → json.Unmarshal(message) → WSMessage{Type, Data}
    → hub.HandleMessage(client, msg) → switch msg.Type → gọi handler tương ứng
```

Cấu hình:
- `ReadLimit`: 4096 bytes — message lớn hơn sẽ bị drop
- `ReadDeadline`: 60s — nếu không nhận gì trong 60s → đóng
- `PongHandler`: mỗi pong nhận được → reset ReadDeadline về +60s

### WritePump (goroutine 2)
```
Loop vô hạn:
  select:
    case message <- client.Send:
      conn.SetWriteDeadline(now + 10s)
      conn.NextWriter(TextMessage) → write message → close writer
    case <- ticker.C:  // mỗi 54s
      conn.WriteMessage(PingMessage, nil)  // keepalive
```

### Hub Event Loop (`ws/hub.go`)
```go
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.Register:
            h.Clients[client] = true   // thêm vào map
        case client := <-h.Unregister:
            delete(h.Clients, client)   // xóa khỏi map
            close(client.Send)          // đóng channel → WritePump thoát
            matchmaker.Remove(client)   // xóa khỏi hàng đợi
            room.RemovePlayer(client)   // xóa khỏi phòng
        }
    }
}
```

### Message routing (`hub.HandleMessage`)
```go
switch msg.Type {
    "join_queue"  → handleJoinQueue(client, msg)
    "create_room" → handleCreateRoom(client, msg)
    "join_room"   → handleJoinRoom(client, msg)
    "start_game"  → handleStartGame(client)
    "ready"       → handleReady(client)
    "target_hit"  → handleTargetHit(client, msg)
    "leave_room"  → handleLeaveRoom(client)
    default       → send error "unknown message type"
}
```

Mỗi handler parse `msg.Data` bằng `json.Marshal` → `json.Unmarshal` vào struct cụ thể (ví dụ `JoinQueueData`, `TargetHitData`).

---

## 6. Luồng Solo — `handleJoinQueue` với mode="solo"

### Call chain
```
hub.handleJoinQueue(client, msg)
  → json.Unmarshal(msg.Data) → JoinQueueData{Mode:"solo", Language:"en", Level:"A1"}
  → data.Mode == "solo" → KHÔNG đi qua Matchmaker
  → hub.GetVocabs("en", "A1", 14)
    → vocabService.GetRandomSet(ctx, "en", "A1", 14)
      → vocabRepo.GetRandomSet(ctx, ...)
        → SQL: SELECT ... FROM vocabularies WHERE language=$1 AND level=$2 ORDER BY RANDOM() LIMIT $3
  → room := NewRoom("en", "A1", ModeSolo, vocabs)
    → Room{ID: uuid[:8], Code: random 6 chars, State: StateWaiting, TotalRounds: 10}
  → room.AddPlayer(client)
    → lock → check State==Waiting → Players[client] = &PlayerState{Score:0} → client.SetRoom(room)
  → hub.AddRoom(room)
    → lock → Rooms[room.ID]=room, RoomByCode[room.Code]=room
  → client.SendMessage({type:"match_found", data:{room_id, mode:"solo"}})
```

**Sau đó client gửi `"ready"`:**
```
hub.handleReady(client)
  → room.SetReady(client)
    → lock → ps.Ready = true
    → mode==Solo && allReady() → true (1 player, đã ready)
    → go room.startGame()
```

---

## 7. Luồng Duel — Matchmaker

### Call chain khi Player A join
```
hub.handleJoinQueue(client_A, msg)
  → data.Mode == "duel" → chuyển sang Matchmaker
  → matchmaker.Enqueue(client_A, "en", "B1")
    → lock
    → Duyệt queue: không tìm thấy ai match
    → append(queue, {client_A, "en", "B1"})
    → client_A.SendMessage({type:"queue_joined", data:{status:"waiting"}})
    → unlock
```

### Call chain khi Player B join → match
```
matchmaker.Enqueue(client_B, "en", "B1")
  → lock
  → Duyệt queue[0]: entry={client_A, "en", "B1"}
    → "en"=="en" && "B1"=="B1" && A.ID != B.ID → MATCH!
  → Xóa A khỏi queue: queue = append(queue[:0], queue[1:]...)
  → hub.GetVocabs("en", "B1", 14) → 14 từ random từ DB
  → room := NewRoom("en", "B1", ModeDuel, vocabs)
  → room.AddPlayer(client_A: opponent)
  → room.AddPlayer(client_B: client)
  → hub.AddRoom(room)
  → opponent.SendMessage({type:"match_found", data:{room_id, opponent: B.Username, mode:"duel"}})
  → client.SendMessage({type:"match_found", data:{room_id, opponent: A.Username, mode:"duel"}})
  → unlock
```

### Khi cả 2 gửi "ready"
```
room.SetReady(A) → A.Ready = true → allReady()? → false (B chưa ready)
room.SetReady(B) → B.Ready = true → allReady()? → true (cả 2 ready, len==2)
  → go room.startGame()
```

### Lưu ý kỹ thuật
- `matchmaker.queue` là `[]queueEntry`, **không phải channel**. Duyệt tuyến tính O(n).
- `sync.Mutex` bảo vệ toàn bộ thao tác Enqueue/Remove — một thời điểm chỉ 1 goroutine truy cập.
- Nếu client disconnect khi đang trong queue → `hub.Unregister` → `matchmaker.Remove(client)` → duyệt queue xóa theo `client.ID`.

---

## 8. Luồng Battle Royale — Tạo Phòng & Join

### Tạo phòng
```
hub.handleCreateRoom(client_host, msg)
  → parse CreateRoomData{Language:"en", Level:"A1"}
  → hub.GetVocabs("en", "A1", 14) → vocabs
  → room := NewRoom("en", "A1", ModeBattle, vocabs)
  → room.HostID = client_host.ID        // chỉ host mới start được
  → room.AddPlayer(client_host)
  → hub.AddRoom(room)                   // lưu vào Rooms + RoomByCode
  → client_host.SendMessage({type:"room_created", data:{room_code:"A8K3MN", room_id:"abc12345"}})
```

### Join phòng bằng code
```
hub.handleJoinRoom(client_X, msg)
  → parse JoinRoomData{RoomCode:"A8K3MN"}
  → lock.RLock → room = RoomByCode["A8K3MN"] → lock.RUnlock
  → Không tìm thấy? → send error "room not found"
  → room.AddPlayer(client_X)
    → lock → check State==Waiting → check len(Players) < 100 → Players[client_X] = &PlayerState{} → unlock
    → return false nếu đầy hoặc game đã bắt đầu
  → client_X.SendMessage({type:"match_found", data:{room_id, player_count, mode:"battle"}})
  → room.broadcast({type:"player_joined", data:{username:"X", player_count, players:[...]}})
```

### Host bấm Start
```
hub.handleStartGame(client_host)
  → room = client_host.GetRoom()
  → room.StartByHost(client_host)
    → lock
    → client_host.ID != room.HostID? → send error "only host can start"
    → len(Players) < 1? → send error "not enough players"
    → Auto-ready tất cả: for ps := range Players { ps.Ready = true }
    → unlock
    → room.startGame()  // KHÔNG phải go routine, chạy synchronous
```

### Sinh mã phòng
```go
const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"  // 31 ký tự, bỏ I/O/0/1
b := make([]byte, 6)
for i := range b { b[i] = chars[rand.Intn(len(chars))] }
// 31^6 = 887,503,681 tổ hợp
```

---

## 9. Vòng Lặp Game — startGame → nextRound → HandleHit → finishGame

### startGame()
```
room.startGame()
  → lock → State = StateCountdown, CurrentRound = 0 → unlock
  → broadcast({type:"countdown", data:{ms:3000}})        // tất cả client nhận
  → time.Sleep(3 * time.Second)                           // BLOCK goroutine 3 giây
  → room.nextRound()
```

### nextRound()
```
room.nextRound()
  → lock
  → CurrentRound++ (1, 2, ... 10)
  → Nếu CurrentRound > 10 → unlock → finishGame() → return
  → State = StatePlaying
  → Reset tất cả player: ps.Answered = false
  → vocabIdx = (CurrentRound - 1) % len(Vocabs)
  → correctVocab = Vocabs[vocabIdx]
  → targets = generateTargets(correctVocab)
  → unlock
  → broadcast({type:"round_start", data:{round, total:10, question:correctVocab.Meaning, targets, time_ms:5000}})
  → RoundTimer = time.AfterFunc(5s, func() {
      // Gọi sau 5 giây nếu chưa bị Stop
      lock → State=RoundEnd → unlock
      broadcastLeaderboard()
      broadcast({type:"round_end"})
      sleep(2s)
      nextRound()  // đệ quy
    })
```

### generateTargets(correctVocab)
```
1. Gọi generateSpreadPositions(4) → 4 cặp [x, y] từ 6 grid zones
   - 6 vùng không giao nhau: {minX, maxX, minY, maxY}
   - Y bắt đầu từ 30% để tránh đè lên HUD
   - Shuffle zones rồi mỗi target lấy 1 zone
   - Trong mỗi zone: x = minX + rand*(maxX-minX), y tương tự

2. targets[0] = Target{ID: correctVocab.ID[:8], Word, Meaning, X, Y, Correct:true}

3. Loop 3 lần (3 từ sai):
   - Random index trong Vocabs
   - Kiểm tra chưa dùng (map used)
   - targets[i] = Target{..., Correct:false}
   - Thử tối đa 20 lần nếu bị trùng

4. rand.Shuffle(targets) → xáo trộn thứ tự
```

### HandleHit(client, data)
```
room.HandleHit(client, TargetHitData{TargetID:"abc", ReactionMs:342})
  → lock (defer unlock)
  → Kiểm tra State == StatePlaying? → không thì return
  → ps = Players[client] → ps.Answered? → đã trả lời thì return
  → correctID = Vocabs[(CurrentRound-1)%len].ID.String()[:8]
  → isCorrect = (data.TargetID == correctID)
  → ps.Answered = true

  → Nếu đúng:
    points = calculateScore(342)
    // = (5000 - 342) * 1000 / 5000 = 931
    ps.Score += 931
    ps.Reactions = append(ps.Reactions, 342)

  → Nếu sai:
    ps.Score -= 50
    if ps.Score < 0 { ps.Score = 0 }

  → Gửi score_update:
    [Duel] → gửi cho TẤT CẢ player, mỗi người nhận điểm riêng + điểm đối thủ
    [Solo/Battle] → chỉ gửi cho người vừa hit

  → [Battle] go broadcastLeaderboard()  // async, gửi top 5 cho tất cả

  → Chuyển round:
    [Solo/Duel] Nếu đúng → RoundTimer.Stop() → go sleep(1s) → nextRound()
    [Battle] Nếu allAnswered() → RoundTimer.Stop() → go broadcastLeaderboard() + sleep(2s) + nextRound()
```

### calculateScore(reactionMs)
```go
func calculateScore(reactionMs int) int {
    base := 1000
    bonus := (5000 - reactionMs) * base / 5000
    if bonus < 100 { bonus = 100 }   // điểm tối thiểu = 100
    return bonus
}
// 0ms → 1000 điểm, 2500ms → 500, 4500ms → 100 (floor), sai → -50
```

### finishGame()
```
room.finishGame()
  → lock → State = StateFinished
  → getRanking():
    entries = [{username, score} for each player]
    sort.Slice(entries, score DESC)
    ranking = [{Rank:1, Username, Score}, ...]
  → winner logic:
    [Duel] rank[0].Score == rank[1].Score → "draw"
    [Duel] rank[0].Score == 0 → no winner
    [else] rank[0].Username
  → unlock

  → For each player:
    tính avgReaction = sum(Reactions) / len(Reactions)
    accuracy = len(Reactions) * 100 / TotalRounds  // Reactions chỉ lưu khi đúng
    send game_over → {winner, your_score, opponent_score (duel only), stats:{total_rounds, avg_reaction_ms, accuracy}, ranking}
```

---

## 10. Ngắt Kết Nối & Rời Phòng

### Rời chủ động — `leave_room`
```
hub.handleLeaveRoom(client)
  → room = client.GetRoom()
  → room.RemovePlayer(client)   // xử lý bên dưới
  → client.SetRoom(nil)
  → matchmaker.Remove(client)   // xóa khỏi queue nếu đang chờ
```

### Mất kết nối (WiFi ngắt, đóng tab)
```
client.ReadPump() → conn.ReadMessage() returns error
  → defer: hub.Unregister <- client
  → Hub.Run() nhận unregister:
    → delete(Clients, client)
    → close(client.Send) → WritePump nhận ok=false → conn.WriteCloseMessage → return
    → matchmaker.Remove(client)
    → room.RemovePlayer(client)
```

### room.RemovePlayer(client)
```
room.RemovePlayer(client)
  → lock (defer unlock)
  → delete(Players, client)
  → Nếu State == Finished → return (game đã kết thúc, không cần xử lý)

  → [Duel]:
    Gửi "opponent_left" cho player còn lại
    RoundTimer.Stop()
    State = Finished  // kết thúc game luôn

  → [Battle]:
    names = getPlayerNames()
    broadcastUnlocked({type:"player_left", data:{username, player_count, players}})
    // Game tiếp tục bình thường với số người còn lại
```

---

## 11. Lấy Từ Vựng — REST API

### Endpoint
```
GET /api/v1/vocab?lang=en&level=B1&limit=20
```

### Call chain
```
VocabHandler.GetVocabularies(c)
  → c.ShouldBindQuery(&q)      // VocabQuery{Language:"en", Level:"B1", Limit:20}
                                // binding: lang required, oneof=en|zh
  → vocabService.GetByLanguage(ctx, q)
    → clamp limit: <=0 || >100 → 20
    → vocabRepo.FindByLanguage(ctx, q)
      → SQL: SELECT id,word,meaning,language,level,difficulty,category
             FROM vocabularies
             WHERE language = $1 AND level = $2
             ORDER BY RANDOM()
             LIMIT $3
  → response.OK(c, vocabs)    // HTTP 200, JSON array
```

### Lấy từ cho game (internal, không qua HTTP)
```
hub.GetVocabs(language, level, count)
  → vocabService.GetRandomSet(context.Background(), language, level, count)
  → Nếu error hoặc rỗng → fallback: [{word:"hello", meaning:"xin chào"}, {word:"world", meaning:"thế giới"}]
```

---

## 12. Struct & Data Flow Tham Chiếu

### Room States (state machine)
```
StateWaiting → StateCountdown → StatePlaying → StateRoundEnd → StatePlaying → ... → StateFinished
                                     ↑              ↓
                                     └──────────────┘  (loop 10 rounds)
```

### Concurrency Model
```
1 Hub goroutine ─────── chạy mãi, xử lý Register/Unregister
N Client goroutines ──── mỗi client = 2 goroutine (Read + Write)
M Room goroutines ────── startGame → sleep → nextRound (đệ quy trên cùng goroutine)
                         RoundTimer → time.AfterFunc callback trên goroutine OS riêng
                         broadcastLeaderboard (Battle) → go routine riêng
```

### Locking strategy
- `Hub.mu` (`sync.RWMutex`): bảo vệ `Clients`, `Rooms`, `RoomByCode` maps. RLock cho đọc, Lock cho ghi.
- `Room.mu` (`sync.Mutex`): bảo vệ `Players` map, `State`, `CurrentRound`. Lock trước mọi thao tác, unlock khi xong.
- `Matchmaker.mu` (`sync.Mutex`): bảo vệ `queue` slice. Lock toàn bộ Enqueue/Remove.
- `Client.mu` (`sync.Mutex`): chỉ bảo vệ trường `Room` (get/set).

### Hằng số (`room.go`)
```go
maxRounds    = 10     // số round mỗi trận
roundTimeMs  = 5000   // 5 giây mỗi round
numTargets   = 4      // 1 đúng + 3 sai
countdownMs  = 3000   // 3 giây đếm ngược
maxPlayers   = 100    // giới hạn phòng Battle
```

---

## 13. Logging Architecture

### Thiết kế

Hệ thống sử dụng `log/slog` (Go stdlib 1.21+) — structured logging với output JSON (production) hoặc Text (development).

```
┌────────────────────────────────────────────────────┐
│  Logger Layer Architecture                         │
│                                                    │
│  Global Logger (slog.Default)                      │
│    ├── attrs: (configured by LOG_LEVEL, LOG_FORMAT) │
│    │                                               │
│    ├── BOOT Logger (main.go)                       │
│    │   └── component="BOOT"                        │
│    │                                               │
│    ├── HTTP Middleware Logger                       │
│    │   └── component="HTTP", request_id, method,   │
│    │       path, status, latency_ms, ip, user_id   │
│    │                                               │
│    ├── WS Logger (hub.go)                          │
│    │   └── component="WS", player, user_id,        │
│    │       room_id, room_code, player_count         │
│    │                                               │
│    ├── GAME Logger (room.go)                       │
│    │   └── component="GAME", room_id, player,      │
│    │       round, reaction_ms, points, total_score  │
│    │                                               │
│    └── REDIS Logger (redis_adapter.go)             │
│        └── component="REDIS", node_id, room_code,  │
│            from_node, channel                       │
└────────────────────────────────────────────────────┘
```

### Log Levels

| Level | Sử dụng | Ví dụ |
|---|---|---|
| `DEBUG` | Chi tiết game loop, grace period | `hit rejected: too late` |
| `INFO` | Events chính: connect, join, score, game finish | `correct hit`, `room created` |
| `WARN` | Lỗi không nghiêm trọng | `send buffer full`, `proxy not found` |
| `ERROR` | Lỗi cần điều tra | `failed to save player result`, `ws upgrade error` |

### Cấu hình

| Biến môi trường | Mặc định | Mô tả |
|---|---|---|
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` (production), `text` (development) |

### Request Tracing

- **HTTP**: Middleware `RequestID()` sinh UUID 8-char cho mỗi request → header `X-Request-ID` → log field `request_id`
- **WebSocket**: Mỗi log trong WS subsystem kèm `user_id` + `player` để trace theo user
- **Game**: Mỗi Room có contextual logger với `room_id` cố định → filter tất cả log của 1 trận game

### Output JSON ví dụ
```json
{"time":"2026-04-19T01:00:00Z","level":"INFO","msg":"correct hit","component":"GAME","room_id":"abc12345","player":"John","round":3,"reaction_ms":342,"points":931,"total_score":2456}
```

### Files liên quan
- `pkg/logger/logger.go` — Init slog, output toggle
- `internal/middleware/request_logger.go` — RequestID + HTTP request logging

---

## 14. Cross-Instance — Redis Pub/Sub

### Thiết kế

Khi chạy nhiều backend instance, Redis đóng vai trò:
1. **Room Registry** (`HSET lingo:room_registry`) — map room code → node ID
2. **Message Bus** (Pub/Sub channels `lingo:node:{nodeID}`) — chuyển tiếp WS messages giữa instances

```
Instance 1 (Room Owner)         Redis              Instance 2 (Joiner)
   ┌──────────────┐          ┌────────┐          ┌──────────────┐
   │ Room ABCDEF   │◄─Pub/Sub─│Registry│─Pub/Sub─►│ Real WS User │
   │ + ProxyClient │─────────►│Channels│◄─────────│              │
   └──────────────┘          └────────┘          └──────────────┘
```

### ProxyClient

`Client` với `RelayFunc` — khi Room broadcast, ProxyClient gửi qua Redis thay vì WebSocket.

### Message Protocol

| Message | Direction | Mô tả |
|---|---|---|
| `proxy_join` | Joiner → Owner | Player ở instance khác muốn join room |
| `proxy_action` | Joiner → Owner | Forward target_hit/ready/start/leave |
| `relay_ws` | Owner → Joiner | Chuyển WS broadcast cho real client |
| `proxy_left` | Joiner → Owner | Player disconnect |

### Chạy multi-instance

```bash
docker compose -f docker-compose.yml -f docker-compose.scale.yml up --build -d
```

### Files liên quan
- `internal/ws/redis_adapter.go` — RedisAdapter, room registry, Pub/Sub
- `docker-compose.scale.yml` — 2 backend instances
- `nginx/default.scale.conf` — upstream pool load balancing

