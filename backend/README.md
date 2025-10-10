# K8s GPU Monitoring Dashboard Backend

K8s上でPrometheusからGPUメトリクスを取得・表示するためのREST APIサーバーです。

![Go](https://img.shields.io/badge/Go-1.24-blue)

## API Endpoint

### Health Check

```http
GET /api/healthz
```

サーバーとPrometheus接続可能性の確認

**レスポンス例:**

```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T12:00:00Z",
    "version": "1.0.0"
  }
}
```

### GPUメトリクス取得

```http
GET /api/v1/gpu/metrics
```

全GPUの詳細なメトリクス情報を取得（並行クエリで高速化）

**レスポンス例:**

```json
{
  "success": true,
  "data": [
    {
      "node_name": "gpu-node-1",
      "gpu_index": 0,
      "gpu_name": "NVIDIA Tesla V100",
      "utilization": 75.5,
      "memory_used": 8.0,
      "memory_total": 16.0,
      "memory_free": 8.0,
      "memory_utilization": 50.5,
      "temperature": 65.0,
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ],
  "message": "GPU metrics retrieved successfully"
}
```

### GPUプロセス取得

```http
GET /api/v1/gpu/processes
```

GPU上で稼働中のプロセス情報を取得（Prometheusのプロセスメトリクスを並行クエリ）

**レスポンス例:**

```json
{
  "success": true,
  "data": [
    {
      "node_name": "gpu-node-1",
      "gpu_index": 0,
      "pid": 1234,
      "process_name": "python",
      "user": "alice",
      "command": "python train.py",
      "gpu_memory": 1024,
      "cpu": 8.5,
      "memory": 15.2,
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ],
  "message": "GPU processes retrieved successfully"
}
```

## プロジェクト構造

```plaintext
backend/
├── cmd/
│   └── server/
│       └── main.go              # アプリケーションエントリーポイント
├── internal/
│   ├── handlers/
│   │   ├── gpu.go               # GPUメトリクス関連ハンドラー
│   │   └── gpu_test.go          # ハンドラーのテスト
│   ├── middleware/
│   │   └── middleware.go        # CORS・ログ・リカバリミドルウェア
│   ├── models/
│   │   └── gpu.go               # データモデル定義
│   ├── prometheus/
│   │   └── client.go            # Prometheusクライアント
│   └── timeutil/
│       └── timeutil.go          # 時刻関連のモジュール
├── go.mod                       # Go 1.24モジュール定義
└── Dockerfile                   # マルチステージDockerビルド
```

## Configuration

### 環境変数

| Variable | Description | Default |
|----------|-------------|---------|
| `PROMETHEUS_URL` | Prometheus Server URL | `http://localhost:9090` |
| `PORT` | APIサーバーのポート | `8080` |

## Responce Format

すべてのAPIレスポンスは以下の統一形式です：

```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully",
  "error": null
}
```

エラー時：

```json
{
  "success": false,
  "error": "Error description",
  "message": null,
  "data": null
}
```

## Development

### Requirements

- Go 1.24+
- Accessable Prometheus Server

### Dev Environment

```bash
# 依存関係を取得
go mod download

# 開発モードでサーバーを起動
go run cmd/server/main.go

# 環境変数を指定して起動
PROMETHEUS_URL=http://prometheus:9090 PORT=8080 go run cmd/server/main.go
```

### Prod Environment

```bash
# ビルドの最適化
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gpu-monitoring-api cmd/server/main.go

# 実行
./gpu-monitoring-api
```

### Docker

```bash
# マルチステージビルドでイメージを構築
docker build -t gpu-monitoring-api .

# コンテナを実行
docker run -p 8080:8080 \
  -e PROMETHEUS_URL=http://prometheus:9090 \
  gpu-monitoring-api

# ヘルスチェック付きで実行
docker run -p 8080:8080 \
  --health-cmd="wget --no-verbose --tries=1 --spider http://localhost:8080/api/healthz || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  -e PROMETHEUS_URL=http://prometheus:9090 \
  gpu-monitoring-api
```

## テスト

### 単体テスト

```bash
# すべてのテストを実行
go test ./...

# カバレッジ付きでテスト
go test -cover ./...

# 詳細出力
go test -v ./...

# 特定のパッケージのテスト
go test -v ./internal/handlers/
```

### テストの特徴

- **モックPromtheusクライアント**: 外部依存なしでテスト実行
- **HTTPテスト**: httptest.Recorderを使用したHTTPハンドラーテスト
- **エラーケース**: 正常系・異常系の包括的テスト

### ベンチマークテスト

```bash
# ベンチマークテスト実行
go test -bench=. ./...

# メモリプロファイル付き
go test -bench=. -benchmem ./...
```

## 必要なPrometheusメトリクス

このAPIは以下のNVIDIA GPUメトリクス（nvidia-gpu-exporter対応）を前提とする：

```promql
# GPU利用率（パーセンテージ）
nvidia_gpu_utilization_percent

# メモリ関連（バイト単位）
nvidia_gpu_used_memory_bytes
nvidia_gpu_total_memory_bytes  
nvidia_gpu_free_memory_bytes
nvidia_gpu_memory_utilization_percent

# 温度（摂氏）
nvidia_gpu_temperature_celsius

# プロセス関連
nvidia_gpu_process_gpu_memory_bytes
nvidia_gpu_process_cpu_percent
nvidia_gpu_process_memory_percent
```

各メトリクスには以下のラベルが必要：

- `node`: Kubernetesノード名
- `gpu`: GPU インデックス番号

## トラブルシューティング

### よくある問題

1. **Prometheus接続エラー**

   ```bash
   # Prometheus疎通確認
   curl http://prometheus-server:9090/api/v1/query?query=up
   
   # DNS確認
   nslookup prometheus-server
   ```

2. **メトリクス取得エラー**

   ```bash
   # nvidia-gpu-exporterの状態確認
   kubectl get pods -l app=nvidia-gpu-exporter
   
   # メトリクス確認
   curl http://prometheus:9090/api/v1/query?query=nvidia_gpu_utilization_percent
   ```

3. **メモリ不足**

   ```bash
   # リソース使用量確認
   docker stats gpu-monitoring-api
   
   # メモリ制限を調整
   docker run --memory=512m gpu-monitoring-api
   ```

### デバッグ情報

ログレベルの調整：

```bash
# 詳細ログで起動
LOG_LEVEL=debug go run cmd/server/main.go
```

デバッグエンドポイント（開発時のみ）：

```bash
# メモリ使用量
curl http://localhost:8080/debug/vars

# pprof（開発ビルド時）
go tool pprof http://localhost:8080/debug/pprof/profile
```
