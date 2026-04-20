# 📘 Learning Guide: Deployment (Docker + GKE + Ingress)

> Hướng dẫn deploy, quản lý Kubernetes / Deployment and K8s management guide

## Prerequisites
- Docker basics (image, container, Dockerfile)
- Linux command line

---

## 1. Docker — Containerization

### Multi-stage Build (Backend)
```dockerfile
# Stage 1: Build binary
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.19
COPY --from=builder /server .
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/health || exit 1
CMD ["./server"]
```

**Tại sao multi-stage? / Why multi-stage?**
- Build stage: ~1GB (Go SDK + dependencies)
- Runtime stage: ~15MB (binary + alpine)

### Next.js Standalone Build (Frontend)
```dockerfile
# next.config.js: output: 'standalone'
# → Produces self-contained server (no node_modules needed)
```

### Docker Compose (Local Development)

```yaml
# docker-compose.yml — 5 services
services:
  nginx:      # Reverse proxy (port 80)
  backend:    # Go API (port 8080)
  frontend:   # Next.js (port 3000)
  postgres:   # Database (port 5432)
  redis:      # Cache/PubSub (port 6379)
```

```bash
# Chạy local / Run locally
docker compose up -d
# → http://localhost (nginx routes /api/* → backend, /* → frontend)
```

---

## 2. Kubernetes (GKE) — Production

### Architecture
```
GKE Autopilot Cluster
├── Namespace: lingo-sniper
│   ├── Deployment: backend (2 replicas)
│   │   ├── backend container
│   │   └── cloud-sql-proxy sidecar
│   ├── Deployment: frontend (1 replica)
│   ├── Deployment: redis (1 replica)
│   ├── Service: backend (ClusterIP)
│   ├── Service: frontend (ClusterIP)
│   ├── Service: redis (ClusterIP)
│   ├── Ingress (Nginx) + TLS (cert-manager)
│   └── Secret: lingo-secrets (DB_URL, JWT_SECRET, etc.)
```

### Key K8s Files

| File | Chức năng / Purpose |
|------|---------------------|
| `k8s/namespace.yaml` | Tạo namespace `lingo-sniper` |
| `k8s/backend/deployment.yaml` | Backend pods + Cloud SQL Proxy sidecar |
| `k8s/backend/service.yaml` | ClusterIP service expose port 8080 |
| `k8s/frontend/deployment.yaml` | Frontend pod |
| `k8s/frontend/service.yaml` | ClusterIP service expose port 3000 |
| `k8s/redis/deployment.yaml` | Redis pod |
| `k8s/ingress.yaml` | Nginx Ingress + TLS config |
| `k8s/cluster-issuer.yaml` | Let's Encrypt TLS certificate issuer |
| `k8s/kustomization.yaml` | Kustomize overlay (image tag management) |

### Essential Commands / Lệnh cần biết

```bash
# Xem pods
kubectl get pods -n lingo-sniper

# Xem logs
kubectl logs deployment/backend -n lingo-sniper -f

# Xem logs của 1 pod cụ thể
kubectl logs backend-7f6578b987-zpjzg -n lingo-sniper -c backend

# Restart deployment (pull latest image)
kubectl rollout restart deployment/backend -n lingo-sniper

# Apply changes
kubectl apply -f k8s/backend/deployment.yaml
kubectl apply -k k8s/  # Kustomize

# Scale
kubectl scale deployment/backend --replicas=3 -n lingo-sniper

# Describe (debug)
kubectl describe pod <pod-name> -n lingo-sniper

# Secret management
kubectl create secret generic lingo-secrets -n lingo-sniper \
  --from-literal=db-url='postgresql://...' \
  --from-literal=jwt-secret='...' \
  --from-literal=redis-url='redis://...'
```

---

## 3. Ingress & TLS

### WebSocket qua Nginx Ingress
```yaml
# ingress.yaml — annotations quan trọng
metadata:
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"  # 1 hour WS timeout
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
    nginx.ingress.kubernetes.io/upstream-hash-by: "$remote_addr" # Session affinity
```

### TLS via cert-manager
```yaml
# cluster-issuer.yaml → Let's Encrypt
# ingress.yaml → tls: secretName: lingo-tls
# → Auto-provisions and renews TLS certificates
```

---

## 4. Cloud SQL Proxy (Sidecar Pattern)

```yaml
# Cloud SQL Auth Proxy chạy cùng pod với backend
containers:
  - name: backend
    image: .../backend:latest
  - name: cloud-sql-proxy         # Sidecar
    image: gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.14.1
    args:
      - "--structured-logs"
      - "--auto-iam-authn"
      - "lingo-sniper-prod:us-central1:lingo-sniper-db"
```

**Tại sao? / Why?**
- Secure connection to Cloud SQL without exposing public IP
- IAM-based auth (không cần manage database passwords)

---

## 5. Troubleshooting / Xử lý sự cố

| Issue | Cách kiểm tra / How to check | Cách fix / Fix |
|-------|------------------------------|----------------|
| Pod CrashLoopBackOff | `kubectl describe pod <name>` | Check logs, fix error, redeploy |
| Image pull error | `kubectl describe pod <name>` | Verify image tag exists in registry |
| WS connection fails | Check Ingress annotations | Add timeout + upstream-hash-by |
| TLS cert not working | `kubectl describe certificate` | Verify DNS + cluster-issuer |
| DB connection refused | Check Cloud SQL Proxy logs | Verify IAM permissions |

### Bài tập / Exercises
1. ✏️ `docker compose up` local → verify tất cả 5 services hoạt động
2. ✏️ Dùng `kubectl logs` xem log backend trên production
3. ✏️ Scale backend lên 3 replicas → xem khi nào pod ready
4. ✏️ Cố tình tạo lỗi (sai DB_URL) → quan sát CrashLoopBackOff → fix
