# K8s GPU Monitoring Dashboard

Kubernetes上でPrometheusからGPUメトリクスを取得・表示するための統合監視ダッシュボード

![GPU Dashboard](https://img.shields.io/badge/Status-Production%20Ready-green)
![Go](https://img.shields.io/badge/Go-1.24-blue)
![React](https://img.shields.io/badge/React-19-blue)
![TypeScript](https://img.shields.io/badge/TypeScript-5.7-blue)
![Vite](https://img.shields.io/badge/Vite-7.0-purple)

## Quick Start

### Helm install

```bash
# Helmリポジトリ追加
helm repo add gpu-monitoring https://v01d42.github.io/k8s-gpu-monitoring-dev
helm repo update

# 基本インストール
helm install gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --create-namespace \
  --set backend.env.PROMETHEUS_URL=http://prometheus-server:9090 \
  --set ingress.hosts[0].host=gpu-monitoring.local

# Ingress IPを確認してhostsファイルに追加
kubectl get ingress -n gpu-monitoring
echo "192.168.1.100 gpu-monitoring.local" | sudo tee -a /etc/hosts
```

### Access

#### Web UI
```bash
# Ingressアクセス
http://gpu-monitoring.local

# Port-forwardでローカルアクセス
kubectl port-forward -n gpu-monitoring svc/gpu-monitoring-frontend 3000:80
# http://localhost:3000 でアクセス
```

#### API
```bash
# ヘルスチェック
curl http://gpu-monitoring.local/api/health

# GPUメトリクス取得
curl http://gpu-monitoring.local/api/v1/gpu/metrics

# GPU利用率のみ取得（軽量）
curl http://gpu-monitoring.local/api/v1/gpu/utilization

# GPUノード一覧
curl http://gpu-monitoring.local/api/v1/gpu/nodes
```

## プロジェクト構成

```
k8s-gpu-monitoring-dev/
├── backend/                      # Go 1.24 API サーバー
│   ├── cmd/server/main.go       # メインエントリーポイント
│   ├── internal/
│   │   ├── handlers/           # HTTP hander（GPU related API）
│   │   │   ├── gpu.go          # Main hander
│   │   │   └── gpu_test.go     # Unit test
│   │   ├── middleware/         # HTTP middleware
│   │   │   └── middleware.go   # CORS, log, recovery
│   │   ├── models/             # Data model definition
│   │   │   └── gpu.go          # GPU related struct
│   │   └── prometheus/         # Prometheus client
│   │       └── client.go       # HTTP API client
│   ├── go.mod                  # Go 1.24 molude setting
│   └── Dockerfile
├── frontend/                   # React 19
│   └── Dockerfile
├── charts/                     # Helm Charts
│   └── k8s-gpu-monitoring-dev/ # Helm Chart
│       ├── Chart.yaml          # Chart difinition
│       ├── values.yaml         # Default setting value
│       └── templates/          # Kubernetes manifest
│           ├── _helpers.tpl    # Helm helper
│           ├── backend/        # Backend resources
│           ├── frontend/       # Frontend resources
│           └── ingress.yaml    # Ingress setting
├── scripts/                    # Operation script
│   └── release.sh              # Release process automation
├── .github/workflows/          # CI/CD
│   └── release.yml             # GitHub Actions workflows
└── docs/                       # Document
    └── DEPLOYMENT.md           # Deployment guide
```

## 開発環境

### 前提条件
- **Go 1.24以上**
- **Node.js 24以上**
- **Docker & Docker Compose**

### ローカル開発環境

#### 1. バックエンドAPI
```bash
cd backend
go mod download
go run cmd/server/main.go
# サーバーが http://localhost:8080 で起動
```

#### 2. フロントエンド
```bash
cd frontend
npm ci
npm run dev
```

#### 3. 開発時のアクセス
- **フロントエンド**: http://localhost:3000
- **バックエンドAPI**: http://localhost:8080
- **APIドキュメント**: http://localhost:8080/api/health

## 設定

### 環境変数

#### バックエンド
```bash
PROMETHEUS_URL=http://prometheus-server:9090  # PrometheusサーバーURL
PORT=8080                                     # APIサーバーポート
```

### 必要なPrometheusメトリクス

以下のNVIDIA GPUメトリクス（nvidia-gpu-exporter対応）が必要：

```promql
# GPU利用率
nvidia_gpu_utilization_percent

# メモリ関連
nvidia_gpu_used_memory_bytes
nvidia_gpu_total_memory_bytes
nvidia_gpu_free_memory_bytes
nvidia_gpu_memory_utilization_percent

# 温度
nvidia_gpu_temperature_celsius
```

## カスタマイズ

### Helm設定

#### 基本設定
```yaml
# values.yaml
global:
  imageRegistry: "ghcr.io/v01d42/k8s-gpu-monitoring-dev"

backend:
  enabled: true
  replicas: 1
  resources:
    requests:
      cpu: "250m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"
  env:
    PROMETHEUS_URL: "http://prometheus-server:9090"
    
frontend:
  enabled: true
  replicas: 1
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "200m"
      memory: "256Mi"
  
ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: gpu-monitoring.local
      paths:
        - path: /api
          pathType: Prefix
          backend:
            service: backend
            port: 8080
        - path: /
          pathType: Prefix
          backend:
            service: frontend
            port: 80
```

#### セキュリティ設定
```yaml
backend:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    runAsGroup: 1001
    capabilities:
      drop:
        - ALL
    readOnlyRootFilesystem: true
    allowPrivilegeEscalation: false

frontend:
  securityContext:
    runAsNonRoot: true
    runAsUser: 101
    runAsGroup: 101
    capabilities:
      drop:
        - ALL
    readOnlyRootFilesystem: true
    allowPrivilegeEscalation: false
```

## テスト

### バックエンドテスト
```bash
cd backend
go test ./internal/handlers/...
```

### フロントエンドテスト
```bash
```

### Helmチャートテスト
```bash
helm test gpu-monitoring --namespace gpu-monitoring
```

## API仕様

### エンドポイント

| Method | Path | 説明 | レスポンス |
|--------|------|------|-----------|
| GET | `/api/health` | ヘルスチェック・Prometheus接続確認 | `APIResponse` |
| GET | `/api/v1/gpu/metrics` | 全GPUの詳細メトリクス | `APIResponse<GPUMetrics[]>` |
| GET | `/api/v1/gpu/nodes` | GPU搭載ノード一覧 | `APIResponse<GPUNode[]>` |
| GET | `/api/v1/gpu/utilization` | GPU利用率のみ（軽量） | `APIResponse<GPUUtilization[]>` |

## 監視・運用

### ヘルスチェック
```bash
# Backend API
kubectl exec -n gpu-monitoring deployment/gpu-monitoring-backend -- \
  wget -qO- http://localhost:8080/api/health

# Frontend
kubectl exec -n gpu-monitoring deployment/gpu-monitoring-frontend -- \
  wget -qO- http://localhost:80/health
```

### ログ確認
```bash
# Backend ログ
kubectl logs -n gpu-monitoring deployment/gpu-monitoring-backend -f

# Frontend ログ
kubectl logs -n gpu-monitoring deployment/gpu-monitoring-frontend -f

# 全体ログ
kubectl logs -n gpu-monitoring -l app.kubernetes.io/name=k8s-gpu-monitoring-dev -f
```

### リソース確認
```bash
# Pod状態確認
kubectl get pods -n gpu-monitoring

# リソース使用量確認
kubectl top pods -n gpu-monitoring

# 詳細情報
kubectl describe pods -n gpu-monitoring
```

## アップグレード

### Helmでのアップグレード
```bash
# リポジトリ更新
helm repo update

# アップグレード
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring

# 特定バージョンにアップグレード
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --version 1.0.1
```

### 設定変更アップグレード
```bash
# リソース設定変更
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set backend.resources.requests.cpu=500m \
  --set backend.resources.requests.memory=512Mi
```

## 開発・コントリビューション

### リリースプロセス

```bash
# 新バージョンリリース
./scripts/release.sh 1.0.1
# 1. Chart.yamlとvalues.yamlのバージョン更新
# 2. Git コミット・タグ作成
# 3. GitHub Actions による自動ビルド・デプロイ
```

## アンインストール

```bash
# アプリケーション削除
helm uninstall gpu-monitoring --namespace gpu-monitoring

# Namespace削除
kubectl delete namespace gpu-monitoring

# /etc/hostsエントリ削除
sudo sed -i '/gpu-monitoring.local/d' /etc/hosts
```