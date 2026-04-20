# 🎯 Interview Questions: DevOps & Infrastructure

> Câu hỏi phỏng vấn DevOps qua dự án / DevOps interview prep

---

## Docker

### Q1: Multi-stage Docker build hoạt động thế nào?

**Answer:**
```dockerfile
# Stage 1: Build (Go SDK ~1GB)
FROM golang:1.25-alpine AS builder
RUN go build -o /server

# Stage 2: Runtime (~15MB)
FROM alpine:3.19
COPY --from=builder /server .
```
- Stage 1 chỉ dùng để build → thrown away
- Final image chỉ chứa binary + minimal OS
- Benefit: Image size 1GB → 15MB, attack surface nhỏ hơn

### Q2: Docker HEALTHCHECK dùng để làm gì?

**Answer:**
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1
```
- Docker daemon periodically check container health
- Status: `healthy`, `unhealthy`, `starting`
- Orchestrator (K8s, Docker Compose) có thể restart unhealthy containers
- `--start-period`: Grace period cho app khởi động

---

## Kubernetes

### Q3: Readiness Probe vs Liveness Probe?

**Answer:**
| | Readiness | Liveness |
|---|-----------|----------|
| **Mục đích** | "Can receive traffic?" | "Is still alive?" |
| **Fail → consequence** | Remove from Service (no traffic) | Kill + restart pod |
| **Use case** | DB chưa ready, still warming up | Deadlock, memory leak, hung |
| **Endpoint** | `/ready` (check DB + Redis) | `/health` (lightweight) |

```yaml
# Dự án dùng:
readinessProbe:
  httpGet:
    path: /health     # Sẽ đổi thành /ready khi build mới
    port: 8080
livenessProbe:
  httpGet:
    path: /health
    port: 8080
```

### Q4: Sidecar pattern — Cloud SQL Proxy?

**Answer:**
- **Pattern**: 2 containers trong 1 pod, share network namespace
- **Cloud SQL Proxy**: Handles secure connection to Cloud SQL
  - IAM-based auth (không cần database password)
  - Encrypted connection (không cần SSL config)
  - Auto-reconnect on transient errors
- Backend connects to `localhost:5432` → Proxy forwards to Cloud SQL
- **Why sidecar**: Tách network concern ra khỏi application code

### Q5: Service vs Ingress vs LoadBalancer?

**Answer:**
```
Internet → Ingress (Nginx) → Service (ClusterIP) → Pods
```
| Type | Scope | Use |
|------|-------|-----|
| **ClusterIP** | Internal only | Pod-to-pod communication |
| **NodePort** | External on node IP | Dev/testing |
| **LoadBalancer** | External cloud LB | Simple external access (costly) |
| **Ingress** | HTTP routing + TLS | Production (path-based routing, TLS termination) |

Dự án dùng: Ingress (nginx) → routes `/api/*` to backend, `/*` to frontend

---

## CI/CD

### Q6: Workload Identity Federation vs Service Account Keys?

**Answer:**
| | WIF (dự án dùng) | SA Keys (traditional) |
|---|---|---|
| **Secret management** | ❌ No secrets | ✅ JSON key file stored |
| **Rotation** | Auto (OIDC tokens short-lived) | Manual rotation needed |
| **Leak risk** | None (no key exists) | High (key file can leak) |
| **Setup** | More complex (IAM config) | Simple (download key) |
| **Best practice** | ✅ Google recommended | ⚠️ Legacy |

### Q7: Tại sao tag Docker images bằng git SHA thay vì `latest`?

**Answer:**
- **Traceability**: Biết chính xác commit nào đang chạy trên production
- **Rollback**: `kubectl set image deploy/backend backend=.../backend:abc123` → instant rollback
- **Cache consistency**: `:latest` có thể bị cache → pod không pull image mới
- **Debugging**: Log error → check git SHA → git show → exact code
- **CI/CD flow**: Build tag `${{ github.sha }}` → Kustomize update → deploy

### Q8: Nếu deployment fail, rollback thế nào?

**Answer:**
```bash
# Xem revision history
kubectl rollout history deployment/backend -n lingo-sniper

# Rollback to previous
kubectl rollout undo deployment/backend -n lingo-sniper

# Rollback to specific revision
kubectl rollout undo deployment/backend --to-revision=3 -n lingo-sniper
```
- K8s giữ lại revision history (mặc định 10)
- Rollback = revert to previous ReplicaSet

---

## Monitoring & Troubleshooting

### Q9: Khi production bị lỗi, debug thế nào?

**Answer (step by step):**
1. **Check logs**: `kubectl logs deployment/backend -n lingo-sniper -f`
2. **Search by request_id**: GCP Cloud Logging → filter `request_id=xyz`
3. **Check pod status**: `kubectl get pods -n lingo-sniper` → look for CrashLoopBackOff, Warning
4. **Describe pod**: `kubectl describe pod <name>` → check Events section
5. **Check resource usage**: GKE dashboard → CPU/memory metrics
6. **Check recent deployments**: `kubectl rollout history deployment/backend`
7. **If needed**: Rollback to last known good revision

### Q10: Giải thích structured logging và tại sao quan trọng?

**Answer:**
```json
// Unstructured (bad)
"Error: failed to fetch leaderboard"

// Structured (good)
{"component":"REPO.User","op":"GetLeaderboard","limit":20,"err":"connection refused","duration_ms":102,"request_id":"a1b2c3"}
```
- **Searchable**: Filter by `component=REPO.User` hoặc `request_id=a1b2c3`
- **Alertable**: Set alert khi `duration_ms > 500` hoặc `level=ERROR` count spike
- **Traceable**: 1 request_id → xem toàn bộ call chain
- **Machine parseable**: Log aggregation tools (GCP Logging, ELK, Datadog) parse automatically
