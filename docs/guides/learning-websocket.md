# 📘 Learning Guide: WebSocket & Real-time Architecture

> Hướng dẫn học WebSocket, pub/sub, và kiến trúc real-time / WebSocket fundamentals and scaling

## Prerequisites
- HTTP basics (request-response model)
- Basic networking (TCP, ports)

---

## 1. WebSocket Protocol Fundamentals

### HTTP vs WebSocket

| Aspect | HTTP | WebSocket |
|--------|------|-----------|
| Connection | Request → Response → Close | Persistent, bidirectional |
| Initiation | Client only | Both client and server |
| Overhead | Headers every request (~800 bytes) | 2-14 bytes per frame |
| Use case | CRUD, page loads | Real-time: chat, games, live data |

### Handshake (Upgrade)
```
Client → Server:
GET /ws/game HTTP/1.1
Upgrade: websocket
Connection: Upgrade

Server → Client:
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
```

### Trong dự án / In this project
- Library: `gorilla/websocket`
- Endpoint: `GET /api/v1/ws/game?token=<jwt>`
- File: `handler/game_handler.go` → `upgrader.Upgrade()`

---

## 2. Hub-Client Architecture (Fan-out pattern)

### Concept
```
                    ┌──── Client A (goroutines: read + write)
Hub ────────────────┼──── Client B
   (event loop)     ├──── Client C
                    └──── Client D
```

### File mapping
| Concept | File | Chức năng / Role |
|---------|------|------------------|
| **Hub** | `ws/hub.go` | Central event loop, message routing |
| **Client** | `ws/client.go` | 1 WS connection = 2 goroutines |
| **Room** | `ws/room.go` | Game instance, state machine |
| **Matchmaker** | `ws/matchmaker.go` | Auto-pair players by language/level |

### Hub Event Loop
```go
// hub.go — infinite select loop
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.Register:
            h.Clients[client] = true
        case client := <-h.Unregister:
            delete(h.Clients, client)
            close(client.Send)
        }
    }
}
```

### Client Goroutines
```
Connection established
├── go ReadPump()   → Read from browser → Hub.HandleMessage()
└── go WritePump()  → Read from Send channel → Write to browser

Heartbeat: Ping every 30s, Pong deadline 60s
```

---

## 3. Message Protocol / Giao thức message

### Message Format
```json
{
    "type": "target_hit",
    "data": {
        "target_id": "abc-123",
        "reaction_ms": 342
    }
}
```

### Message Types

**Client → Server:**
```
join_queue    → Tham gia hàng đợi matchmaking
create_room  → Tạo phòng chơi mới
join_room    → Tham gia phòng bằng code
ready        → Đánh dấu sẵn sàng
start_game   → Bắt đầu game (chỉ host)
target_hit   → Click vào target
leave_room   → Rời phòng
```

**Server → Client:**
```
queue_joined      → Đã vào hàng đợi
match_found       → Đã ghép cặp thành công
room_created      → Phòng đã tạo (kèm room code)
player_joined     → Có người mới vào phòng
countdown         → 3...2...1...
round_start       → Câu hỏi + targets mới
score_update      → Cập nhật điểm
live_leaderboard  → Bảng xếp hạng live
round_end         → Kết thúc round
game_over         → Kết quả cuối cùng
host_changed      → Host mới khi host cũ rời
error             → Thông báo lỗi
```

---

## 4. Scaling WebSocket — Multi-Pod with Redis Pub/Sub

### Vấn đề / Problem
- Kubernetes có 2 backend pods
- Player A connects to Pod 1, Player B connects to Pod 2
- Cả 2 cùng phòng → nhưng khác pod!

### Giải pháp / Solution: Redis Pub/Sub Proxy Pattern

```
Pod 1 (Room owner)          Redis          Pod 2 (Player B's pod)
┌─────────────────┐    ┌────────────┐    ┌─────────────────┐
│ Room "ABC123"    │    │            │    │                 │
│ - Player A      │    │  Pub/Sub   │    │ - Player B      │
│ - ProxyClient B │◄──►│  Channels  │◄──►│ - proxyClients  │
│   (virtual)     │    │            │    │   map           │
└─────────────────┘    └────────────┘    └─────────────────┘
```

### Flow in detail
1. Room "ABC" created on Pod 1 → registered in Redis
2. Player B on Pod 2 sends `join_room "ABC"`
3. Pod 2 → Redis: `LookupRoom("ABC")` → returns "Pod 1"
4. Pod 2 → Redis: `PublishToNode(pod1, proxy_join)`
5. Pod 1 creates `ProxyClient` (virtual client whose `SendMessage` relays via Redis)
6. Game events: Pod 1 → Redis → Pod 2 → Player B's real WebSocket

### Files involved
- `ws/redis_adapter.go` — Redis connection, pub/sub, lookup
- `ws/hub.go` — `handleProxyJoin()`, `handleProxyAction()`, `handleRelayWS()`

---

## 5. Error Handling & Resilience

### Server-side
- **Ping/Pong heartbeat**: Detect dead connections (timeout: 60s)
- **Buffered send channel**: `make(chan []byte, 256)` — prevent blocking
- **Buffer full → disconnect**: Don't let slow clients block the server

### Client-side (Frontend)
- **Auto-reconnect**: Exponential backoff on disconnect
- **Connection timeout**: Show error UI if can't connect in 10s
- **Message queue**: Buffer messages during reconnection

### Bài tập / Exercises
1. ✏️ Nếu Redis down, hệ thống hoạt động thế nào? (Hint: graceful degradation)
2. ✏️ Tại sao dùng Pub/Sub mà không dùng Redis Streams?
3. ✏️ ProxyClient khác gì với Client thật?
4. ✏️ Nếu 1 pod restart giữa game, player bị ảnh hưởng thế nào?
