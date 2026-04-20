# 📚 Lingo Sniper — Technical Documentation

> Tài liệu kỹ thuật dự án Lingo Sniper / Technical documentation suite

---

## 🏗️ Architecture / Kiến trúc

| Document | Description |
|----------|-------------|
| [System Design](architecture/system-design.md) | Overall architecture, tech stack, DB schema, deployment topology, design decisions |
| [Backend Flow](architecture/backend-flow.md) | Request lifecycle, middleware chain, WebSocket game flow, cross-pod sync, logging |
| [Frontend Flow](architecture/frontend-flow.md) | Next.js routing, core hooks, component architecture, API layer, styling |

---

## 📘 Learning Guides / Hướng dẫn học

| Guide | Technologies Covered |
|-------|---------------------|
| [Go Backend](guides/learning-go-backend.md) | Go, Gin, slog, goroutines, channels, database/sql |
| [Next.js Frontend](guides/learning-nextjs.md) | Next.js App Router, React hooks, Tailwind CSS, state management |
| [WebSocket & Real-time](guides/learning-websocket.md) | WS protocol, Hub-Client pattern, Redis Pub/Sub, scaling |
| [Deployment](guides/learning-deployment.md) | Docker, GKE, Kubernetes, Ingress, TLS, Cloud SQL Proxy |
| [CI/CD](guides/learning-cicd.md) | GitHub Actions, Workload Identity Federation, Kustomize |
| [Security](guides/learning-security.md) | JWT, bcrypt, CORS, rate limiting, origin validation |

---

## 🎯 Interview Prep / Chuẩn bị phỏng vấn

| Category | Topics |
|----------|--------|
| [Backend Questions](interview/backend-questions.md) | Go fundamentals, architecture patterns, DB design, observability |
| [Frontend Questions](interview/frontend-questions.md) | React/Next.js, state management, performance, accessibility |
| [System Design](interview/system-design-qa.md) | Real-time architecture, WebSocket scaling, caching, bottleneck analysis |
| [DevOps Questions](interview/devops-questions.md) | Docker, Kubernetes, CI/CD, monitoring, troubleshooting |

---

## Quick Start / Bắt đầu nhanh

```bash
# Clone
git clone https://github.com/tinhuynh1/language-arena.git
cd language-arena

# Run locally with Docker Compose
docker compose up -d

# → Open http://localhost
```

## Tech Stack Summary

```
Frontend:  Next.js 15 + React 19 + Tailwind CSS v4
Backend:   Go 1.25 + Gin + gorilla/websocket
Database:  PostgreSQL 16 + Redis 7
Infra:     Docker + GKE Autopilot + Nginx Ingress
CI/CD:     GitHub Actions + Workload Identity Federation
Domain:    lingosniper.lol (TLS via cert-manager)
```
