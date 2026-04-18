# Câu Hỏi Phỏng Vấn — Lingo Sniper Backend

> Các câu hỏi kỹ thuật, system design, và tình huống thực tế liên quan đến project.
> Mỗi câu kèm phân tích vấn đề và giải pháp đề xuất.

---

## Phần 1: Kiến Trúc & System Design

### Q1. Nếu chạy 2 instance backend cùng lúc (horizontal scale), hệ thống sẽ gặp vấn đề gì?

**Vấn đề:**

- **WebSocket state mất đồng bộ**: Hub, Room, Matchmaker đều lưu in-memory. Nếu Player A kết nối tới Instance 1, Player B kết nối tới Instance 2 → Matchmaker ở mỗi instance không biết nhau → **không thể ghép cặp**.
- **Room bị chia đôi**: Player trong cùng 1 room nhưng ở 2 instance khác nhau → broadcast chỉ tới những player cùng instance.
- **RoomByCode không đồng bộ**: Tạo phòng ở Instance 1, join code ở Instance 2 → "room not found".

**Giải pháp:**

| Phương án | Mô tả | Ưu/Nhược |
|---|---|---|
| **Sticky Session** | Load balancer (Nginx/HAProxy) dùng IP hash hoặc cookie để đảm bảo 1 user luôn tới cùng 1 instance | ✅ Đơn giản. ❌ Không giải quyết matchmaking cross-instance |
| **Redis Pub/Sub** | Mỗi instance subscribe Redis channel. Khi match/broadcast → publish qua Redis → tất cả instance nhận | ✅ Giữ in-memory nhưng đồng bộ message. ❌ Cần refactor Hub |
| **Centralized State (Redis)** | Chuyển queue, room state, player map sang Redis. Instance chỉ là proxy WS | ✅ Stateless backend, scale thoải mái. ❌ Phức tạp, latency tăng |
| **Dedicated Game Server** | Tách game engine thành service riêng, không scale phần này. Scale REST API riêng | ✅ Đơn giản, phù hợp giai đoạn hiện tại. ❌ Game server = bottleneck |

**Đề xuất cho project này:** Giai đoạn đầu dùng **Dedicated Game Server** (1 instance xử lý WS) + scale REST API riêng. Khi vượt 1 game server → chuyển sang **Redis Pub/Sub**.

---

### Q2. Hệ thống hiện tại có đáp ứng được 1,000 CCU (concurrent users) không?

**Phân tích:**

- Mỗi WS client = 2 goroutine (ReadPump + WritePump) + 1 `chan []byte` (cap 256).
- 1,000 CCU = 2,000 goroutine → **Go xử lý tốt** (Go có thể chạy hàng triệu goroutine).
- Mỗi goroutine ~8KB stack → 2,000 × 8KB = ~16MB RAM → **không đáng kể**.
- Mỗi connection ~ 1 TCP socket → cần OS cho phép ≥1,000 file descriptors (mặc định macOS: 256, Linux: 1024). **Cần tune `ulimit -n`**.

**Bottleneck thật sự:**

| Component | Giới hạn | Vấn đề |
|---|---|---|
| **PostgreSQL** | `MaxOpenConns` (hiện config) | Mỗi lần tạo room → 1 query `SELECT ... ORDER BY RANDOM() LIMIT 14`. 100 rooms/phút → 100 queries/phút. Ổn. |
| **Matchmaker scan** | O(n) duyệt queue | 1,000 người cùng chờ → mỗi Enqueue duyệt tối đa 1,000 entries. Với Mutex lock → **block tất cả**. Nghẽn khi queue lớn. |
| **Room broadcast** | O(players) mỗi message | 100-player battle room, mỗi hit → broadcastLeaderboard = 100 lần `json.Marshal + channel send`. Nếu 100 người hit gần nhau → **100 × 100 = 10,000 operations** trong thời gian ngắn. |
| **`time.Sleep` blocking** | startGame() sleep 3s trên goroutine | 100 room start cùng lúc = 100 goroutine bị block 3s. Không ảnh hưởng vì Go scheduler. Nhưng nếu 10,000 room → cần review. |

**Kết luận:** 1,000 CCU chơi Solo/Duel (2 người/room) → **hoàn toàn OK**. 1,000 CCU trong 10 Battle rooms (100 người/room) → **có thể gặp contention trên Room.mu**.

**Giải pháp nếu cần scale:**
1. **Matchmaker**: Chuyển từ `[]queueEntry` sang map phân theo `language:level` key → O(1) lookup.
2. **Room broadcast**: Dùng fan-out goroutine pool thay vì loop tuần tự.
3. **Vocab query cache**: Cache `GetRandomSet` result trong Redis (TTL 5 phút) → giảm DB load.

---

### Q3. Tại sao chọn gorilla/websocket thay vì Server-Sent Events (SSE)?

**Trả lời:**
- Game cần **bi-directional realtime** — client gửi `target_hit`, server gửi `score_update`. SSE chỉ hỗ trợ server → client.
- Với SSE, client phải gửi `target_hit` qua HTTP POST → thêm latency (HTTP overhead, TCP handshake nếu không keep-alive).
- WebSocket giữ 1 TCP connection duy nhất → latency thấp hơn (~1ms vs ~50-100ms cho HTTP roundtrip).
- Trade-off: WS phức tạp hơn (connection management, reconnect logic). SSE đơn giản hơn, tự reconnect, hoạt động qua CDN/proxy dễ hơn.

---

### Q4. Giải thích cơ chế concurrency của Room. Có race condition nào không?

**Phân tích hiện tại:**

```
Room.mu (sync.Mutex) bảo vệ:
  - Players map
  - State
  - CurrentRound
  - PlayerState (Score, Answered, Ready)
```

**Potential race condition đã được xử lý:**
- `HandleHit` lock toàn bộ Room → 2 player hit cùng lúc → serialized, an toàn.
- `broadcastLeaderboard` trong Battle mode gọi bằng `go` → nhưng bên trong có lock riêng → OK.

**Potential race condition CHƯA xử lý:**

1. **RoundTimer callback vs nextRound goroutine:**
   ```
   Player đúng → go sleep(1s) → nextRound()     ← goroutine A
   Timer 5s fires → State=RoundEnd → nextRound() ← goroutine B
   ```
   Nếu player trả lời đúng ở giây thứ 4.5 → `RoundTimer.Stop()` **có thể** race với callback đã được schedule. `time.AfterFunc` document: Stop returns false nếu timer đã fire. Nhưng code hiện tại **không kiểm tra return value** → có thể 2 goroutine cùng gọi `nextRound()` → **double round advance**.

   **Fix:** Thêm guard trong `nextRound()`:
   ```go
   if r.State != StatePlaying && r.State != StateRoundEnd {
       return // already advanced
   }
   ```

2. **Duel: cả 2 player hit đúng gần nhau:**
   ```
   Player A hit đúng → isCorrect → go sleep(1s) → nextRound()
   Player B hit đúng 10ms sau → isCorrect → go sleep(1s) → nextRound()
   ```
   Cả 2 đều trigger `nextRound` → **double advance**. Vì `ps.Answered` chỉ ngăn 1 player hit 2 lần, không ngăn 2 player cùng trigger advance.

   **Fix:** Dùng flag `roundAdvancing bool` trong Room, set true khi bắt đầu advance, check trước khi advance.

---

## Phần 2: Database & Performance

### Q5. `ORDER BY RANDOM() LIMIT N` có vấn đề gì với bảng lớn?

**Vấn đề:**
- PostgreSQL thực hiện `ORDER BY RANDOM()` bằng cách: scan toàn bộ bảng → gán random value cho mỗi row → sort → lấy N dòng đầu.
- Complexity: **O(n log n)** với n = số row trong bảng.
- 1,000 từ → nhanh (~1ms). 1,000,000 từ → **chậm (~500ms-1s)**.

**Giải pháp khi bảng lớn:**

```sql
-- Cách 1: TABLESAMPLE (PostgreSQL, gần ngẫu nhiên, O(1))
SELECT * FROM vocabularies TABLESAMPLE BERNOULLI(1)
WHERE language = $1 AND level = $2 LIMIT $3;

-- Cách 2: Random offset (exact random, O(1))
SELECT * FROM vocabularies
WHERE language = $1 AND level = $2
OFFSET floor(random() * (SELECT count(*) FROM vocabularies WHERE language = $1 AND level = $2))
LIMIT $3;

-- Cách 3: Pre-computed random column
ALTER TABLE vocabularies ADD COLUMN rand_order DOUBLE PRECISION DEFAULT random();
CREATE INDEX idx_vocab_rand ON vocabularies(language, level, rand_order);
-- Query: WHERE rand_order >= random() ORDER BY rand_order LIMIT N
```

---

### Q6. Tại sao không lưu game state vào database trong khi chơi?

**Trả lời:**
- Game round kéo dài 5 giây, 10 rounds = 50 giây. Nếu mỗi hit ghi DB → 4 player × 10 round = 40 DB writes trong 50s. Với 100 rooms → **4,000 writes/phút.** PostgreSQL xử lý được nhưng thêm latency (~5-10ms/write).
- In-memory state → latency ~0ms. Tối ưu cho realtime game.
- **Trade-off**: Nếu server crash giữa game → mất hết state, game kết thúc không lưu.

**Giải pháp nếu cần persistence:**
1. **Write-behind**: Ghi vào Redis sau mỗi round, batch flush vào PostgreSQL sau game kết thúc.
2. **Event sourcing**: Log tất cả events (`target_hit`, `score_update`) vào Redis Stream/Kafka → replay nếu crash.
3. **Checkpoint**: Sau mỗi round, async ghi snapshot vào Redis với TTL 5 phút.

---

### Q7. Connection pool PostgreSQL nên config thế nào?

**Phân tích:**
- Backend hiện tại dùng `database/sql` built-in pooling.
- Mỗi `db.QueryContext` → lấy 1 connection từ pool → query → trả lại.
- Nếu tất cả connection đang busy → `QueryContext` block cho tới khi có connection free hoặc context timeout.

**Khuyến nghị cho 1,000 CCU:**
```go
db.SetMaxOpenConns(25)           // PostgreSQL mặc định max_connections = 100
db.SetMaxIdleConns(10)           // giữ 10 connection rảnh sẵn
db.SetConnMaxLifetime(5 * time.Minute)  // tránh connection stale
db.SetConnMaxIdleTime(1 * time.Minute)  // đóng connection rảnh sau 1 phút
```

**Tại sao 25 chứ không phải 100?**
- PostgreSQL performance giảm khi >100 concurrent connections (context switching, shared buffer contention).
- 25 connections đủ cho: REST API queries + vocab random set queries. WS game logic **không dùng DB** trong khi chơi (chỉ lúc tạo room).

---

## Phần 3: WebSocket & Realtime

### Q8. Nếu client gửi `target_hit` nhưng WebSocket bị delay 2 giây, reactionMs sẽ sai không?

**Phân tích:**

`reactionMs` được tính **ở client-side**:
```javascript
const reactionMs = Date.now() - roundStartTimeRef.current;
```

→ `reactionMs` phản ánh thời gian thực tế từ khi client nhận `round_start` đến khi click. **Network delay không ảnh hưởng** vì cả 2 timestamp đều ở client.

**Nhưng có vấn đề khác:**
- Client có thể **cheat** bằng cách gửi `reactionMs: 1` → luôn được 999 điểm.
- **Không có server-side validation** cho reactionMs.

**Fix:**
```go
func (r *Room) HandleHit(client *Client, data TargetHitData) {
    // Server tính thời gian thực tế
    serverElapsed := time.Since(r.roundStartTime).Milliseconds()
    // Dùng max(client, server-tolerance) để chống cheat
    reactionMs := max(data.ReactionMs, int(serverElapsed) - 500) // cho phép 500ms network jitter
}
```

---

### Q9. `client.Send` channel (cap 256) đầy thì sao?

**Code hiện tại:**
```go
func (c *Client) SendMessage(msg WSMessage) {
    select {
    case c.Send <- data:
    default:
        log.Printf("client %s send buffer full, dropping message")
    }
}
```

**Vấn đề:**
- Nếu client nhận chậm (mạng yếu, tab bị suspend) → channel đầy → **message bị drop silently**.
- Trong game: drop `round_start` → player không thấy câu hỏi mới. Drop `score_update` → điểm hiển thị sai.

**Giải pháp:**

| Phương án | Mô tả |
|---|---|
| **Disconnect client** | Channel đầy = client lag quá nhiều → đóng connection, buộc reconnect |
| **Ring buffer** | Thay channel bằng ring buffer, ghi đè message cũ → luôn có message mới nhất |
| **Priority queue** | `round_start` và `game_over` = high priority, không bao giờ drop. `score_update` = low priority, có thể drop |
| **Backpressure** | Giảm tốc độ game nếu có player lag → ảnh hưởng trải nghiệm player khác |

---

### Q10. Reconnect flow hiện tại có hỗ trợ không? Nếu player mất mạng 3 giây rồi có lại thì sao?

**Hiện tại: KHÔNG hỗ trợ reconnect.**

- Mất WS → ReadPump error → Unregister → RemovePlayer → Game coi như player rời.
- Player reconnect = tạo Client mới, UUID giống nhưng Room đã bị xóa.

**Giải pháp reconnect:**
```
1. Khi player disconnect → KHÔNG xóa khỏi Room ngay
   → Đặt PlayerState.Disconnected = true
   → Bắt đầu grace timer (30s)

2. Nếu player reconnect trong 30s:
   → Tìm Room bằng userID
   → Gán Client mới vào PlayerState cũ
   → Gửi lại state hiện tại: round, score, targets

3. Nếu hết 30s → xóa thật sự
```

**Cần thêm:**
- `Hub.ClientsByUserID map[uuid.UUID]*Client` — tra cứu nhanh client theo userID
- `Room.DisconnectedPlayers map[uuid.UUID]*PlayerState` — giữ state tạm

---

## Phần 4: Security

### Q11. Hệ thống có chống cheat được không?

**Các vector cheat hiện tại:**

| Cheat | Cách thực hiện | Mức độ |
|---|---|---|
| **Fake reactionMs** | Gửi `{reaction_ms: 1}` → luôn max điểm | 🔴 Critical |
| **Auto-answer** | Đọc `targets[].correct` từ `round_start` message → tự động hit target đúng | 🔴 Critical |
| **Multi-account** | Mở 2 tab, duel với chính mình → farm điểm | 🟡 Medium |
| **Flood hit** | Gửi nhiều `target_hit` cùng lúc | 🟢 Low (đã có `ps.Answered` check) |

**Fix cho từng vector:**

1. **Fake reactionMs** → Server-side timing (xem Q8)
2. **Auto-answer** → **KHÔNG gửi `correct: true/false` trong `targets`**. Chỉ gửi `targets[].id` và `targets[].word`. Server so sánh target_id với correctID. **Hiện tại đang gửi `correct` field → client biết đáp án.**
   ```go
   // Fix: xóa field Correct khi gửi cho client
   type TargetClient struct {
       ID   string  `json:"id"`
       Word string  `json:"word"`
       X    float64 `json:"x"`
       Y    float64 `json:"y"`
       // KHÔNG có Correct
   }
   ```
3. **Multi-account** → Track IP + fingerprint, giới hạn 1 active game/IP.
4. **Flood hit** → Đã có `ps.Answered` check, an toàn.

---

### Q12. JWT secret bị lộ thì ảnh hưởng gì?

**Ảnh hưởng:**
- Attacker tạo được JWT token tùy ý với bất kỳ `user_id` → **impersonate bất kỳ user**.
- Truy cập WS endpoint → chơi game, submit điểm giả dưới tên người khác.
- Truy cập protected API → đọc stats, history của bất kỳ user.

**Giải pháp:**
1. **Rotate secret**: Thay JWT secret ngay, tất cả token cũ invalid.
2. **Token versioning**: Thêm `token_version` vào user table. Khi rotate → increment version → token cũ với version cũ bị reject.
3. **Short-lived token + Refresh token**: Access token 15 phút, Refresh token 7 ngày. Lộ access token → ảnh hưởng giới hạn.

---

## Phần 5: Tình Huống Thực Tế

### Q13. 100 người chơi Battle room, tất cả hit cùng lúc — server xử lý thế nào?

**Phân tích:**

```
100 players hit trong ~100ms → 100 goroutine gọi HandleHit()
HandleHit() có r.mu.Lock() → serialized → chỉ 1 goroutine chạy tại 1 thời điểm
→ 99 goroutine đợi
```

**Thời gian xử lý mỗi HandleHit:**
- Lock acquire: ~50ns
- Score calculation: ~10ns
- SendMessage (push vào channel): ~100ns
- broadcastLeaderboard (`go`): spawn goroutine ~1μs
- Unlock: ~50ns
- **Total: ~1-2μs per hit**

**100 hits serialized: ~100-200μs (0.1-0.2ms)** → **không đáng kể**, player không cảm nhận được.

**Nhưng broadcastLeaderboard:**
- 100 hits → 100 lần `go broadcastLeaderboard()`
- Mỗi lần: lock → getRanking (sort 100 entries ~5μs) → broadcast (100 SendMessage ~10μs) → unlock
- 100 lần: **~1.5ms total** → vẫn OK
- **Vấn đề**: 100 leaderboard messages gửi liên tục → client nhận 100 updates liên tiếp → flicker UI. **Fix**: debounce leaderboard broadcast, chỉ gửi tối đa 1 lần / 500ms.

---

### Q14. Server restart giữa lúc có 50 game đang chơi — hậu quả gì?

**Hậu quả:**
- Tất cả WS connection đóng → 50 rooms mất hết state (in-memory).
- Frontend nhận WS close event → hiển thị "Connection lost".
- Không có game nào được ghi vào `game_sessions` (vì `finishGame()` chưa chạy).

**Giải pháp:**
1. **Graceful shutdown**: Trước khi stop, gửi `game_over` cho tất cả rooms → client biết game kết thúc.
2. **State persistence**: Trước shutdown, serialize tất cả Room state vào Redis → sau restart, load lại.
3. **Client reconnect**: Frontend tự reconnect sau 3s, gửi `rejoin_room` với room_id → server restore state từ Redis.

---

### Q15. Nếu muốn thêm ngôn ngữ mới (ví dụ: Nhật, Hàn) thì cần sửa gì?

**Cần sửa:**
1. **Database**: INSERT từ vựng mới vào bảng `vocabularies` với `language = 'ja'` hoặc `'ko'`.
2. **Backend**: **Không cần sửa code.** Matchmaker match theo `language` field → tự động match player cùng ngôn ngữ. Room.GetVocabs query theo `language` param → tự lấy đúng.
3. **Frontend**: Thêm button ngôn ngữ mới + level system (JLPT N5-N1 cho Nhật, TOPIK 1-6 cho Hàn).
4. **Validation**: Hiện tại `VocabQuery.Language` validate `oneof=en zh` → cần thêm `ja ko`.

**Kết luận**: Backend được design **language-agnostic** — ngôn ngữ chỉ là 1 field filter, không hardcode logic riêng. Chỉ cần seed data và mở rộng validation.

---

### Q16. Muốn thêm chế độ chơi mới (ví dụ: Time Attack — ai trả lời nhiều nhất trong 60 giây) thì kiến trúc hiện tại có hỗ trợ không?

**Phân tích kiến trúc hiện tại:**
- `Room.Mode` là `model.GameMode` (string: "solo", "duel", "battle").
- Logic game chia nhánh theo mode trong: `SetReady()`, `HandleHit()`, `finishGame()`, `RemovePlayer()`.
- Round-based design: cố định 10 rounds, mỗi round 1 câu hỏi.

**Khó khăn cho Time Attack:**
- Time Attack = 1 round dài 60s, nhiều câu liên tục → phá vỡ mô hình `nextRound()` đệ quy hiện tại.
- Cần sinh câu hỏi mới ngay khi player trả lời đúng (không đợi round end).

**Giải pháp:**
```
Option A: Hack trong kiến trúc hiện tại
  - Mode "timeattack", TotalRounds = 999, roundTimeMs = 60000
  - HandleHit đúng → nextRound() ngay (không sleep)
  - Timer 60s → finishGame() bất kể round nào
  → Quick & dirty, nhưng hoạt động

Option B: Refactor sang Strategy pattern
  - Interface: GameMode { StartGame(), HandleHit(), OnTimeout(), IsFinished() }
  - RoundBasedMode implements GameMode (Solo, Duel, Battle hiện tại)
  - TimeAttackMode implements GameMode (logic riêng)
  - Room gọi r.GameMode.HandleHit() thay vì switch/case
  → Clean, extensible, nhưng cần refactor lớn
```

---

### Q17. WebSocket message format hiện tại có tối ưu không? Có nên dùng Protobuf/MessagePack?

**Hiện tại:** JSON plain text.

```json
{"type":"round_start","data":{"round":1,"total":10,"question":"Xin chào","targets":[{"id":"abc","word":"Hello","x":25.3,"y":45.7,"correct":true},...],"time_ms":5000}}
```

**Kích thước:** ~300-500 bytes per message.

**So sánh:**

| Format | Kích thước | Parse speed | Debug |
|---|---|---|---|
| JSON | 100% (baseline) | Chậm nhất | ✅ Dễ đọc |
| MessagePack | ~60-70% | Nhanh 2-3x | ❌ Binary |
| Protobuf | ~40-50% | Nhanh 5-10x | ❌ Cần schema |

**Có cần optimize không?**
- 1,000 CCU × 10 msg/s × 400 bytes = **4 MB/s** bandwidth → **không đáng kể với server hiện đại**.
- JSON parse 400 bytes: ~1-5μs → **không phải bottleneck**.
- **Kết luận**: JSON là đủ cho quy mô này. Chuyển sang binary format khi >10,000 CCU hoặc khi bandwidth là vấn đề (mobile 3G).

---

## Phần 6: Câu Hỏi Mở Rộng

### Q18. So sánh kiến trúc monolith hiện tại với microservice. Khi nào nên tách?

**Monolith hiện tại:**
```
1 Go binary = REST API + WebSocket Engine + Game Logic + Auth
```

**Ưu điểm:** Deploy đơn giản, không network overhead giữa service, debug dễ, phù hợp team nhỏ.

**Khi nào tách:**
- **Auth Service**: Khi có nhiều client (mobile app, admin panel) cùng dùng auth → tách ra, dùng chung.
- **Game Engine**: Khi REST API cần scale nhưng game engine không → tách để scale độc lập.
- **Vocab Service**: Khi bộ từ vựng cần CMS riêng → tách thành Content Service.

**Threshold gợi ý**: >5 developer, >10,000 CCU, hoặc khi deploy frequency khác nhau giữa các phần.

---

### Q19. Nếu thêm hệ thống ranking Elo cho Duel, cần thay đổi gì?

**Cần thêm:**
1. `users.elo_rating INT DEFAULT 1200` — Elo khởi điểm.
2. **Matchmaker**: Thay vì match theo `language + level`, thêm điều kiện `abs(elo_A - elo_B) < 200`.
3. **finishGame()**: Sau game Duel, tính Elo mới:
   ```
   Expected_A = 1 / (1 + 10^((elo_B - elo_A) / 400))
   New_elo_A = elo_A + K * (result - Expected_A)
   // K = 32 cho player mới, 16 cho player cũ
   ```
4. **Queue timeout**: Nếu không tìm thấy match trong 30s → mở rộng khoảng Elo (±200 → ±400 → any).

---

### Q20. Giải thích tại sao dùng `sync.Mutex` thay vì channel cho Room state?

**Trả lời:**

Room state cần **random access** — bất kỳ player nào cũng có thể hit target bất cứ lúc nào, và cần đọc/ghi `Players` map, `State`, `Score` ngay lập tức.

- **Channel**: Phù hợp cho **sequential event processing** (Hub.Run() dùng channel cho Register/Unregister). Nhưng nếu dùng channel cho Room → tất cả operations (hit, ready, leave) phải serialize qua 1 channel → **bottleneck**, và code phức tạp hơn (phải marshal mọi thứ thành event struct).
- **Mutex**: Phù hợp cho **shared mutable state** với nhiều goroutine truy cập. Lock → đọc/ghi trực tiếp → unlock. Code đơn giản, performance tốt (uncontended mutex lock = ~25ns).

**Go idiom**: *"Share memory by communicating"* (channel) khi data flow theo 1 hướng. *"Communicate by sharing memory"* (mutex) khi data cần random access từ nhiều nơi.
