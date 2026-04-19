# Bài Học Ký Sự: Triển Khai Lingo Sniper Lên GKE

Tài liệu này lưu lại toàn bộ các kinh nghiệm "xương máu", các lỗi đã gặp và kiến thức học được trong quá trình đưa Lingo Sniper từ Docker Compose lên Google Kubernetes Engine (GKE).

---

## 1. Kiến Trúc Tổng Thể (The Big Picture)
Thay vì đóng gói cục bộ, hệ thống được cấu trúc lại theo chuẩn **Cloud-Native**:
- **Backend (Go) & Frontend (Next.js):** Đóng gói thành Docker Container, đẩy lên Google Artifact Registry và chạy dười dạng các K8s Pods có thể linh hoạt Scale (nhân bản).
- **Database (Cloud SQL for PostgreSQL):** Tách rời hoàn toàn khỏi K8s để Google Cloud tự động quản lý việc Backup và Đảm bảo tính khả dụng (High Availability - HA).
- **Redis State:** Đóng vai trò là Message Broker (Pub/Sub) giúp các Pod Backend khác nhau có thể đồng bộ trạng thái Game (WebSockets) theo thời gian thực.

## 2. Những Cú Trượt Chân Đáng Nhớ & Cách Giải Quyết

### A. Bài toán "Vắt Kiệt RAM/CPU" (K8s Resource Requests)
- **Lỗi gặp phải:** GKE báo lỗi `Insufficient cpu` và từ chối chạy Pod, dù máy chủ `e2-medium` (2 vCPU) vẫn còn khá trống.
- **Bài học:** Trong Kubernetes, nếu bạn set `limits` (ví dụ 500m CPU) mà **không set** `requests`, K8s sẽ tự động ngầm định `requests = limits = 500m`. Nó xí chỗ trước 500m CPU dù app chưa hề dùng tới!
- **Cách khắc phục:** Luôn khai báo rõ ràng một mức `requests` nhỏ (ví dụ `cpu: 10m`) và mức `limits` cao để Pod vừa dễ dàng được xếp lịch (schedule) vào các Node rẻ tiền, vừa có khả năng bung sức khi có nhiều người chơi.

### B. Mớ Bòng Bong Đăng Nhập K8s -> Cloud SQL
- **Lỗi gặp phải:** Backend K8s bị chối từ truy cập (Access Denied) khi cố connect vào Cloud SQL do không xác thực được bảo mật mạng.
- **Bài Học Bảo Mật (Workload Identity):** Tuyệt đối KHÔNG xuất file `.json` chứa key kết nối ném vào trong Code. Thay vào đó, chúng ta đã cấu hình **Workload Identity**. Bằng cách tạo ra một `Kubernetes Service Account` (KSA) và trói nó với `Google Service Account` (GSA), Backend tự động có "kim bài miễn tử" bảo mật nội bộ của Google mà không cần mật khẩu.
- **Mô hình Sidecar:** Chạy kèm một Container `cloud-sql-proxy` ngay bên trong cùng 1 Pod với Backend. Code Go chỉ cần connect ngây thơ vào `localhost:5432`, còn việc mã hoá SSL qua rào Google để cái Sidecar lo!

### C. Nginx Ingress & WebSocket Xung Đột
- **Lỗi gặp phải:** Định gắn thêm `configuration-snippet` vào Ingress để cấu hình WebSocket, nhưng Ingress Controller dỗi không thèm tải (bị Drop do vi phạm Regex).
- **Bài học:** Bản update Nginx Ingress mới nhất đã chặn tính năng `snippet` vì lý do bảo mật. Thật may mắn, Nginx hiện đại đã tự động bật tính năng Upgrade HTTP -> WebSockets mặc định. Chúng ta chỉ cần cấu hình `proxy-read-timeout` cực dài (86400s) để giữ cho kết nối Game không bị ngắt giữa chừng.

### D. Sự Cố Kiến Trúc CPU (ARM vs AMD)
- **Lỗi gặp phải:** Nếu bạn dùng Mac chip M (ARM64) để build Image và quăng lên GKE (vốn đa phần là chip Intel AMD64), K8s sẽ báo lỗi rớt mạng `exec format error`.
- **Bài học:** Luôn phải nhớ build chéo kiến trúc bằng lệnh `docker buildx build --platform linux/amd64` để đảm bảo Image chạy mượt mà trên môi trường máy chủ Linux truyền thống.

## 3. Quá Trình Xin Chữ Ký "Ổ Khóa Xanh" (HTTPS / SSL)
- Thay vì cấu hình lằng nhằng hoặc mua SSL tốn tiền, chúng ta đã tích hợp sức mạnh của **Cert-Manager**.
- Chỉ bằng việc ném thêm dòng `cert-manager.io/cluster-issuer: "letsencrypt-prod"` vào file cấu hình Ingress, K8s tự động liên lạc với Let's Encrypt, chứng minh bạn là chủ nhân đích thực của `lingosniper.lol`, lấy chứng chỉ về, và gắn hoàn toàn tự động!

---
## Nhìn Lại Sự Nghiệp
Qua dự án này, bạn không chỉ code xong 1 con Game thời gian thực, mà bạn đã đụng chạm tới toàn bộ các ngóc ngách "khét lẹt" nhất của DevOps hiện đại: Load Balancing, CI/CD, Containerization, Orchestration, Pub/Sub, Workload Identity, và DNS Management.

Chúc mừng bạn đã lên trình Master Kubernetes! 🎓
