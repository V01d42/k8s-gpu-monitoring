# GPU Monitoring API

K8s上でPrometheusからGPUメトリクスを取得・表示するためのGo 1.24で実装されたREST APIサーバーです。

## 特徴

- **Go 1.24の新しいServeMux**: 最新のHTTPルーティング機能とメソッド指定ルーティングを使用
- **標準ライブラリ中心**: 外部依存を最小限に抑えたシンプルな設計
- **並行処理**: Goroutineを使用した効率的なPrometheusクエリの並列実行
- **RESTful API**: 標準的なHTTP APIデザインパターン
- **包括的エラーハンドリング**: カスタムエラータイプと適切なHTTPステータスコード
- **CORS対応**: プリフライトリクエストを含む完全なクロスオリジン対応
- **Graceful Shutdown**: シグナルハンドリングによる安全なサーバー停止処理
- **構造化ログ**: 運用に適したログ出力

## API エンドポイント

### ヘルスチェック
```
GET /api/healthz
```
サーバーとPrometheus接続の健全性をチェック

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
```
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
      "memory_utilization": 50.0,
      "temperature": 65.0,
      "power_draw": 250.0,
      "power_limit": 300.0,
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ],
  "message": "GPU metrics retrieved successfully"
}
```

## プロジェクト構造

```
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
│   └── prometheus/
│       └── client.go            # Prometheusクライアント
├── go.mod                       # Go 1.24モジュール定義
└── Dockerfile                   # マルチステージDockerビルド
```

## 設定

環境変数で設定できます：

- `PROMETHEUS_URL`: PrometheusサーバーのURL（デフォルト: `http://localhost:9090`）
- `PORT`: APIサーバーのポート（デフォルト: `8080`）

## レスポンス形式

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

## 実行方法

### 前提条件
- Go 1.24以上
- アクセス可能なPrometheusサーバー

### 開発環境
```bash
# 依存関係を取得
go mod download

# 開発モードでサーバーを起動
go run cmd/server/main.go

# 環境変数を指定して起動
PROMETHEUS_URL=http://prometheus:9090 PORT=8080 go run cmd/server/main.go
```

### 本番環境
```bash
# 最適化されたビルド
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

このAPIは以下のNVIDIA GPUメトリクス（nvidia-gpu-exporter対応）を期待します：

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
```

各メトリクスには以下のラベルが必要：
- `node`: Kubernetesノード名
- `gpu`: GPU インデックス番号

## アーキテクチャ

```
┌─────────────────┐
│   HTTP Client   │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│   ServeMux      │ ← Go 1.24 新機能
│  (Method-based  │   メソッド指定ルーティング
│   Routing)      │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│   Middleware    │ ← CORS, Logging, Recovery
│   (Chain)       │   パニック回復
└─────────────────┘
         │
         ▼
┌─────────────────┐
│   GPUHandler    │ ← HTTP Request Handlers
│                 │   構造化レスポンス
└─────────────────┘
         │
         ▼
┌─────────────────┐
│ PrometheusClient│ ← 並行クエリ実行
│  (Concurrent    │   効率的なHTTPクライアント
│   Queries)      │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  Prometheus     │
│    Server       │
└─────────────────┘
```

## パフォーマンス

### 最適化機能
- **並行クエリ**: 複数のPrometheusクエリをgoroutineで並列実行
- **HTTPクライアント再利用**: コネクションプールの活用
- **適切なタイムアウト**: リクエスト・レスポンスタイムアウト設定
- **効率的なJSONパース**: 標準ライブラリの最適化されたencoding/json

### ベンチマーク結果
```bash
# 典型的なパフォーマンス（開発環境）
GET /api/v1/gpu/metrics: ~200ms (6個のGPU、並行クエリ)
GET /api/healthz: ~50ms
```

## セキュリティ

### 実装済みセキュリティ機能
- **入力検証**: 適切なHTTPメソッドとパスの検証
- **エラー情報制限**: 機密情報を含まないエラーメッセージ
- **リソース制限**: タイムアウトとリクエストサイズ制限
- **CORS設定**: セキュアなクロスオリジン設定
- **パニック回復**: Recovery ミドルウェアによるパニック処理

### Dockerセキュリティ
- **非rootユーザー**: コンテナ内で非特権ユーザーで実行
- **読み取り専用ルートファイルシステム**: セキュリティ強化
- **最小限のイメージ**: Alpine Linuxベース

### セキュリティ推奨事項
- HTTPS終端をIngress/LoadBalancerで実装
- Prometheusアクセスを内部ネットワークに制限
- 適切なKubernetes NetworkPolicyの設定

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
