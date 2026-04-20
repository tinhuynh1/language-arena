# 🎯 Interview Questions: System Design

> Câu hỏi phỏng vấn system design qua dự án / System design interview prep

---

## Real-time Architecture

### Q1: Design một hệ thống multiplayer game real-time?

**Answer using this project:**

**Requirements:**
- Real-time interaction (<100ms latency)
- Multiple concurrent game rooms
- Score tracking + leaderboard
- Scalable to thousands of concurrent players

**Architecture:**
```
Client → Nginx (LB + TLS) → Go Backend (N pods) → PostgreSQL
                                   ↕
                              Redis Pub/Sub
```

**Key decisions:**
1. **WebSocket** cho bidirectional real-time (không polling, không SSE)
2. **Hub-Room pattern**: 1 Hub manages N Rooms, mỗi Room là state machine
3. **Stateful rooms**: Room state in memory (không lưu DB mỗi action — quá chậm)
4. **Persist only results**: Kết quả game lưu DB khi game kết thúc
5. **Redis Pub/Sub**: Sync giữa pods khi 2 players ở 2 pods khác nhau

### Q2: Làm sao scale WebSocket server?

**Answer:**
**Problem**: WS connections are stateful — player A ở Pod 1, room ở Pod 2

**Solution: Proxy Pattern with Redis Pub/Sub**
1. Room lookup: Redis stores `room_code → pod_id` mapping
2. Cross-pod join: Player on Pod 2 → proxy_join via Redis → Pod 1 creates ProxyClient
3. ProxyClient: Virtual client whose `SendMessage()` publishes back via Redis
4. Result: Player B sees real-time updates even though Room is on different pod

**Session Affinity**: Nginx `upstream-hash-by: $remote_addr` → same user hits same pod usually (reduces cross-pod traffic)

**Alternative approaches:**
- **Sticky sessions only**: Simpler but fails if pod restarts
- **Shared state (Redis)**: Store all room state in Redis — more consistent but higher latency
- **Dedicated game servers**: Separate WS servers from API servers — better for large scale

### Q3: Tại sao không dùng Server-Sent Events (SSE) thay WebSocket?

**Answer:**
| Aspect | SSE | WebSocket |
|--------|-----|-----------|
| Direction | Server → Client only | Bidirectional |
| Client actions | Need separate HTTP requests | Send via same connection |
| Protocol | HTTP/2 compatible | Separate protocol (upgrade) |
| Use case | Live feeds, notifications | Games, chat, collaborative editing |

- Game cần bidirectional: player gửi `target_hit` + server gửi `score_update`
- SSE chỉ 1 chiều → phải dùng request HTTP riêng cho player actions → thêm latency

---

## Database Design

### Q4: Giải thích schema design cho game results?

**Answer:**
```
users → game_session_players (N:M) → game_sessions
```
- **Separated tables**: `game_sessions` (session-level data) + `game_session_players` (per-player data)
- **Tại sao**: 1 game session có N players, mỗi player có score/rank riêng
- **Leaderboard query**: `SELECT FROM users ORDER BY total_score DESC` — denormalized cho fast reads
- **Denormalization trade-off**: `users.total_score` updated after each game (duplicated data) — acceptable vì read >> write

### Q5: Caching strategy?

**Answer (current + ideal):**

**Current**: No caching layer (queries hit DB directly)

**Improvement plan:**
1. **Leaderboard cache**: Redis ZSET — `ZADD leaderboard <score> <user_id>` — O(log N) insert/query
2. **Vocabulary cache**: Redis with TTL — vocab data rarely changes
3. **Cache invalidation**: Update leaderboard cache when game finishes
4. **NOT cached**: Auth tokens (stateless JWT — no server-side cache needed)

---

## Scalability

### Q6: Bottleneck analysis — dự án scale thế nào nếu 10,000 concurrent users?

**Answer:**

| Component | Current | Bottleneck at 10K? | Solution |
|-----------|---------|-------|----------|
| **Backend pods** | 2 | ⚠️ Maybe | Auto-scale to 10+ pods |
| **DB connections** | 25/pod = 50 total | ⚠️ Yes | PgBouncer connection pooler |
| **Redis** | Single instance | ✅ Fine | Redis handles 100K+ ops/sec |
| **Rate limiter** | In-memory per pod | ⚠️ Inconsistent | Redis-based rate limiting |
| **WS connections** | ~500/pod | ✅ Fine | Go handles 10K+ connections per process |
| **Room state** | In-memory | ⚠️ Pod restart = lost | Checkpoint to Redis periodically |

### Q7: Nếu PostgreSQL chậm, làm gì?

**Answer (theo thứ tự):**
1. **Measure first**: Repo layer logs slow queries (>100ms) — identify actual bottleneck
2. **Indexing**: Add indexes on frequently queried columns (`users.email`, `users.total_score`)
3. **Connection pooling**: PgBouncer in front of Cloud SQL
4. **Read replicas**: Leaderboard queries → read replica, writes → primary
5. **Caching**: Redis cache for leaderboard (invalidate on game end)
6. **Last resort**: Shard by language/region

---

## Infrastructure

### Q8: Tại sao GKE Autopilot thay vì self-managed K8s?

**Answer:**
- **Autopilot**: GKE manages node pools, auto-repair, auto-upgrade
- **Trade-off**: Less control over node config, slightly higher cost
- **Benefit for small team**: Zero infrastructure management
- **Alternative**: GKE Standard (more control) or plain VMs (more work)

### Q9: Deployment strategy — Rolling Update vs Blue-Green?

**Answer:**
```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 0          # Don't create extra pods
    maxUnavailable: 1    # Take down 1 at a time
```
- **Rolling Update**: Gradual replacement, zero-downtime
- **maxSurge: 0**: Tiết kiệm resources (không tạo pod thừa)
- **Trade-off**: Nếu new version bị lỗi → 1 pod chạy old, 1 pod chạy new → inconsistent
- **Improvement**: Blue-Green (full switch) hoặc Canary (traffic splitting)
