# 🚀 Deploy Lingo Sniper lên GKE + Cloud SQL

> Step-by-step guide deploy lên Google Cloud. Yêu cầu: GCP account với Free Trial ($300 credit).

---

## Mục lục

1. [Cài đặt CLI tools](#1-cài-đặt-cli-tools)
2. [Tạo GCP Project](#2-tạo-gcp-project)
3. [Tạo Artifact Registry](#3-tạo-artifact-registry)
4. [Build & Push Docker Images](#4-build--push-docker-images)
5. [Tạo Cloud SQL (PostgreSQL + HA)](#5-tạo-cloud-sql-postgresql--ha)
6. [Tạo GKE Cluster](#6-tạo-gke-cluster)
7. [Setup Cloud SQL Proxy IAM](#7-setup-cloud-sql-proxy-iam)
8. [Deploy lên K8s](#8-deploy-lên-k8s)
9. [Xác minh & Truy cập](#9-xác-minh--truy-cập)
10. [Quản lý & Vận hành](#10-quản-lý--vận-hành)
11. [Dọn dẹp tài nguyên](#11-dọn-dẹp-tài-nguyên)

---

## 1. Cài đặt CLI tools

```bash
# Google Cloud SDK
brew install --cask google-cloud-sdk

# kubectl (K8s CLI)
gcloud components install kubectl

# Login
gcloud auth login
```

---

## 2. Tạo GCP Project

```bash
# Đặt biến
export PROJECT_ID="lingo-sniper-prod"   # Thay bằng tên project của bạn
export REGION="asia-southeast1"          # Singapore — gần VN nhất
export ZONE="asia-southeast1-a"

# Tạo project (hoặc dùng project có sẵn)
gcloud projects create $PROJECT_ID --name="Lingo Sniper"
gcloud config set project $PROJECT_ID

# Enable billing (bắt buộc cho Free Trial)
# → Vào https://console.cloud.google.com/billing và link project

# Enable APIs
gcloud services enable \
  container.googleapis.com \
  sqladmin.googleapis.com \
  artifactregistry.googleapis.com \
  compute.googleapis.com \
  iam.googleapis.com
```

---

## 3. Tạo Artifact Registry

Nơi lưu Docker images (thay thế Docker Hub).

```bash
# Tạo repository
gcloud artifacts repositories create lingo-sniper \
  --repository-format=docker \
  --location=$REGION \
  --description="Lingo Sniper Docker images"

# Config Docker auth
gcloud auth configure-docker ${REGION}-docker.pkg.dev
```

---

## 4. Build & Push Docker Images

```bash
# Từ thư mục project root
cd /path/to/language-arena

# Đặt image prefix
export IMAGE_PREFIX="${REGION}-docker.pkg.dev/${PROJECT_ID}/lingo-sniper"

# Build & Push Backend
docker build -t ${IMAGE_PREFIX}/backend:latest ./backend
docker push ${IMAGE_PREFIX}/backend:latest

# Build & Push Frontend
docker build -t ${IMAGE_PREFIX}/frontend:latest ./frontend
docker push ${IMAGE_PREFIX}/frontend:latest
```

> **Tip**: Trên Mac M1/M2, thêm `--platform linux/amd64` vào `docker build` vì GKE chạy x86.
> ```bash
> docker build --platform linux/amd64 -t ${IMAGE_PREFIX}/backend:latest ./backend
> ```

---

## 5. Tạo Cloud SQL (PostgreSQL + HA)

```bash
# Tạo instance với High Availability
gcloud sql instances create lingo-sniper-db \
  --database-version=POSTGRES_16 \
  --tier=db-f1-micro \
  --region=$REGION \
  --availability-type=REGIONAL \
  --storage-size=10GB \
  --storage-auto-increase \
  --backup-start-time=02:00 \
  --enable-point-in-time-recovery

# ⏳ Mất 5-10 phút...

# Set password cho user postgres
gcloud sql users set-password postgres \
  --instance=lingo-sniper-db \
  --password="YOUR_STRONG_PASSWORD_HERE"

# Tạo database
gcloud sql databases create lingodb --instance=lingo-sniper-db

# Tạo user cho app
gcloud sql users create lingouser \
  --instance=lingo-sniper-db \
  --password="YOUR_APP_PASSWORD_HERE"

# Lấy connection name (cần cho Cloud SQL Proxy)
gcloud sql instances describe lingo-sniper-db --format='value(connectionName)'
# Output: lingo-sniper-prod:asia-southeast1:lingo-sniper-db
```

> **Ghi chú**: `--availability-type=REGIONAL` bật HA. Google sẽ tạo standby instance ở zone khác, tự động failover khi primary down.

---

## 6. Tạo GKE Cluster

```bash
# Tạo cluster (1 zonal, free management fee)
gcloud container clusters create lingo-sniper-cluster \
  --zone=$ZONE \
  --num-nodes=1 \
  --machine-type=e2-medium \
  --disk-size=30 \
  --enable-autorepair \
  --enable-autoupgrade \
  --workload-pool=${PROJECT_ID}.svc.id.goog

# ⏳ Mất 5-10 phút...

# Connect kubectl
gcloud container clusters get-credentials lingo-sniper-cluster --zone=$ZONE

# Verify
kubectl get nodes
# NAME                                                  STATUS   ROLES    AGE   VERSION
# gke-lingo-sniper-cluster-default-pool-xxx-xxx         Ready    <none>   1m    v1.xx
```

### Install Nginx Ingress Controller

```bash
# Helm (nếu chưa có)
brew install helm

# Add nginx-ingress repo
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Install
helm install nginx-ingress ingress-nginx/ingress-nginx \
  --namespace ingress-nginx --create-namespace \
  --set controller.service.type=LoadBalancer

# Lấy External IP (chờ 1-2 phút)
kubectl get svc -n ingress-nginx
# NAME                                 TYPE           EXTERNAL-IP     PORT(S)
# nginx-ingress-ingress-nginx-ctrl     LoadBalancer   34.xxx.xxx.xxx  80:xxx/TCP,443:xxx/TCP
```

---

## 7. Setup Cloud SQL Proxy IAM

Cloud SQL Proxy chạy như sidecar trong backend pod, cần IAM service account.

```bash
# Tạo GCP Service Account
gcloud iam service-accounts create cloud-sql-proxy \
  --display-name="Cloud SQL Proxy"

# Grant role
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:cloud-sql-proxy@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"

# Tạo K8s Service Account
kubectl create namespace lingo-sniper
kubectl create serviceaccount cloud-sql-proxy -n lingo-sniper

# Link K8s SA → GCP SA (Workload Identity)
gcloud iam service-accounts add-iam-policy-binding \
  cloud-sql-proxy@${PROJECT_ID}.iam.gserviceaccount.com \
  --role="roles/iam.workloadIdentityUser" \
  --member="serviceAccount:${PROJECT_ID}.svc.id.goog[lingo-sniper/cloud-sql-proxy]"

# Annotate K8s SA
kubectl annotate serviceaccount cloud-sql-proxy \
  --namespace=lingo-sniper \
  iam.gke.io/gcp-service-account=cloud-sql-proxy@${PROJECT_ID}.iam.gserviceaccount.com
```

---

## 8. Deploy lên K8s

### 8.1 Cập nhật manifests

Trước khi apply, thay placeholder trong các file K8s:

```bash
# Thay PROJECT_ID và REGION trong manifests
cd k8s

# macOS sed
sed -i '' "s|PROJECT_ID|${PROJECT_ID}|g" backend/deployment.yaml frontend/deployment.yaml
sed -i '' "s|REGION|${REGION}|g" backend/deployment.yaml frontend/deployment.yaml
```

### 8.2 Tạo Secrets

```bash
# Tạo secret (thay YOUR_APP_PASSWORD bằng password đã set ở bước 5)
kubectl create secret generic lingo-secrets \
  --namespace=lingo-sniper \
  --from-literal=db-url="postgres://lingouser:YOUR_APP_PASSWORD@localhost:5432/lingodb?sslmode=disable" \
  --from-literal=jwt-secret="$(openssl rand -base64 32)" \
  --from-literal=redis-url="redis://redis:6379"
```

### 8.3 Apply manifests

```bash
# Deploy theo thứ tự
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/redis/deployment.yaml
kubectl apply -f k8s/backend/deployment.yaml
kubectl apply -f k8s/backend/service.yaml
kubectl apply -f k8s/frontend/deployment.yaml
kubectl apply -f k8s/frontend/service.yaml
kubectl apply -f k8s/ingress.yaml

# Hoặc apply tất cả cùng lúc
kubectl apply -f k8s/ --recursive
```

---

## 9. Xác minh & Truy cập

```bash
# Check pods
kubectl get pods -n lingo-sniper
# NAME                        READY   STATUS    RESTARTS   AGE
# backend-xxx                 2/2     Running   0          1m    ← 2/2 = backend + sql-proxy
# backend-yyy                 2/2     Running   0          1m
# frontend-zzz                1/1     Running   0          1m
# redis-aaa                   1/1     Running   0          1m

# Check logs (structured JSON!)
kubectl logs deploy/backend -n lingo-sniper -c backend | head -5

# Check health
export LB_IP=$(kubectl get svc -n ingress-nginx \
  -o jsonpath='{.items[0].status.loadBalancer.ingress[0].ip}')
curl http://$LB_IP/health
# {"status":"ok","timestamp":1713524400}

# Mở browser
echo "🚀 Open: http://$LB_IP"
```

---

## 10. Quản lý & Vận hành

### Scale backend

```bash
# Thay đổi replicas
kubectl scale deploy/backend -n lingo-sniper --replicas=3

# Hoặc setup auto-scaling
kubectl autoscale deploy/backend -n lingo-sniper \
  --min=2 --max=5 --cpu-percent=70
```

### Xem logs (structured)

```bash
# Xem logs real-time
kubectl logs -f deploy/backend -n lingo-sniper -c backend

# Filter bằng jq
kubectl logs deploy/backend -n lingo-sniper -c backend | jq 'select(.component=="GAME")'

# Xem log game cụ thể
kubectl logs deploy/backend -n lingo-sniper -c backend | jq 'select(.room_id=="abc123")'
```

### Rolling update

```bash
# Build image mới
docker build --platform linux/amd64 -t ${IMAGE_PREFIX}/backend:v2 ./backend
docker push ${IMAGE_PREFIX}/backend:v2

# Update deployment
kubectl set image deploy/backend backend=${IMAGE_PREFIX}/backend:v2 -n lingo-sniper

# Theo dõi rollout
kubectl rollout status deploy/backend -n lingo-sniper

# Rollback nếu cần
kubectl rollout undo deploy/backend -n lingo-sniper
```

### Test Cloud SQL HA Failover

```bash
# Trigger failover từ GCP Console:
# Cloud SQL → lingo-sniper-db → Overview → Failover
# Hoặc CLI:
gcloud sql instances failover lingo-sniper-db

# Quan sát backend logs — connection sẽ drop rồi tự recover
kubectl logs -f deploy/backend -n lingo-sniper -c backend
```

---

## 11. Dọn dẹp tài nguyên

> ⚠️ **QUAN TRỌNG**: Xóa khi không dùng để tránh tốn credit!

```bash
# Xóa K8s resources
kubectl delete -f k8s/ --recursive

# Xóa GKE cluster
gcloud container clusters delete lingo-sniper-cluster --zone=$ZONE --quiet

# Xóa Cloud SQL (⚠️ xóa data!)
gcloud sql instances delete lingo-sniper-db --quiet

# Xóa Docker images
gcloud artifacts repositories delete lingo-sniper \
  --location=$REGION --quiet

# Xóa service account
gcloud iam service-accounts delete \
  cloud-sql-proxy@${PROJECT_ID}.iam.gserviceaccount.com --quiet
```

---

## Tham khảo nhanh

| Lệnh | Mô tả |
|-------|--------|
| `kubectl get pods -n lingo-sniper` | Xem pods |
| `kubectl logs -f deploy/backend -n lingo-sniper -c backend` | Logs backend |
| `kubectl describe pod <name> -n lingo-sniper` | Debug pod |
| `kubectl exec -it deploy/redis -n lingo-sniper -- redis-cli` | Redis CLI |
| `kubectl top pods -n lingo-sniper` | Resource usage |
| `gcloud sql instances describe lingo-sniper-db` | Cloud SQL info |
