# Hướng Dẫn Cài Đặt CI/CD Bằng GitHub Actions & WIF

Tài liệu này hướng dẫn cách cấu hình **Workload Identity Federation (WIF)** để GitHub Actions có thể lấy quyền can thiệp vào Google Cloud (GKE + Artifact Registry) mà KHÔNG cần xuất file JSON chứa key bảo mật.

## Bước 1: Khởi tạo các biến môi trường
Sử dụng Google Cloud Shell hoặc Terminal máy bạn, thay đổi các giá trị sau cho đúng với dự án của bạn rồi chạy (Copy & Paste):

```bash
# SỬA LẠI 3 DÒNG NÀY
export PROJECT_ID="lingo-sniper-prod"
export GITHUB_REPO="michael/language-arena" # VD: "octocat/my-repo"
export SERVICE_ACCOUNT_NAME="github-actions"

# Các dòng dưới tự sinh
export PROJECT_NUMBER=$(gcloud projects describe ${PROJECT_ID} --format="value(projectNumber)")
export SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
```

## Bước 2: Tạo Service Account cho Github
Tạo tài khoản dịch vụ chuyên dùng cho quá trình CI/CD:

```bash
gcloud iam service-accounts create ${SERVICE_ACCOUNT_NAME} \
  --project="${PROJECT_ID}" \
  --display-name="GitHub Actions Service Account"
```

Cấp quyền đẩy Image lên Artifact Registry và Deploy lên GKE:
```bash
# Quyền ghi Artifact Registry
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
  --role="roles/artifactregistry.writer"

# Quyền được Deploy container vào GKE
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
  --role="roles/container.developer"
```

## Bước 3: Tạo WIF Pool & Provider
Tạo một "Pool" chứa danh tính:
```bash
gcloud iam workload-identity-pools create "github-pool" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --display-name="GitHub Actions Pool"
```

Tạo một Provider để chứng thực Github thông qua OIDC:
```bash
gcloud iam workload-identity-pools providers create-oidc "github-provider" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --workload-identity-pool="github-pool" \
  --display-name="GitHub Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
  --issuer-uri="https://token.actions.githubusercontent.com"
```

## Bước 4: Chấp nhận kết nối từ Github Repo của bạn
Trói Github Repo của bạn với cái Service Account vừa tạo ở Bước 2:
```bash
gcloud iam service-accounts add-iam-policy-binding "${SERVICE_ACCOUNT_EMAIL}" \
  --project="${PROJECT_ID}" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/github-pool/attribute.repository/${GITHUB_REPO}"
```

## Bước 5: Cập nhật file `.github/workflows/deploy.yml`
Chạy lệnh này để lấy chuỗi cấu hình **Provider Name**:
```bash
echo "projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/github-pool/providers/github-provider"
```
*Copy kết quả dòng trên, mở file `.github/workflows/deploy.yml`, thay vào tham số `WORKLOAD_IDENTITY_PROVIDER`.*

*Tương tự, copy Email của Service Account thay vào biến `SERVICE_ACCOUNT`.*

---

💥 **HOÀN TẤT!** 
Bây giờ, chỉ cần bạn gõ `git push origin main` thì điều kì diệu sẽ tự động xảy ra trong Tab **Actions** trên GitHub!
