# Deployment Guide

GPU監視ダッシュボードのKubernetesクラスターへのデプロイメントガイドです。

## 前提条件

### 必要なコンポーネント
- **Kubernetes Cluster**: v1.24以上
- **Helm**: v3.14以上
- **Ingress Controller**: nginx-ingress推奨
- **Prometheus**: NVIDIA GPU Exporterが設定済み

### 技術要件
- **Go**: 1.24（バックエンド）
- **Node.js**: 20.0（フロントエンド）
- **Docker**: 対応プラットフォーム linux/amd64

### 必要なPrometheusメトリクス
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

## クイックスタート

### 1. Helmリポジトリ追加
```bash
# Helmリポジトリを追加
helm repo add gpu-monitoring https://v01d42.github.io/k8s-gpu-monitoring-dev

# リポジトリ更新
helm repo update

# 利用可能なチャートバージョン確認
helm search repo gpu-monitoring
```

### 2. 基本インストール
```bash
# 最小構成でインストール
helm install gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --create-namespace \
  --set backend.env.PROMETHEUS_URL=http://prometheus-server:9090 \
  --set ingress.hosts[0].host=gpu-monitoring.local
```

### 3. アクセス確認
```bash
# Pod状態確認
kubectl get pods -n gpu-monitoring

# 出力例:
# NAME                                         READY   STATUS    RESTARTS   AGE
# gpu-monitoring-backend-xxx                  1/1     Running   0          2m
# gpu-monitoring-frontend-xxx                 1/1     Running   0          2m

# Service確認
kubectl get svc -n gpu-monitoring

# Ingress確認
kubectl get ingress -n gpu-monitoring

# Port-forwardでアクセス
kubectl port-forward -n gpu-monitoring svc/gpu-monitoring-frontend 3000:80
# http://localhost:3000 でアクセス
```

## 設定オプション

### 基本設定
```yaml
# values.yaml
global:
  imageRegistry: "ghcr.io/v01d42/k8s-gpu-monitoring-dev"
  
backend:
  enabled: true
  replicas: 1
  image:
    repository: "backend"
    tag: "1.0.0"
    pullPolicy: IfNotPresent
  env:
    PROMETHEUS_URL: "http://prometheus-server:9090"
    PORT: "8080"
    
frontend:
  enabled: true
  replicas: 1
  image:
    repository: "frontend"
    tag: "1.0.0"
    pullPolicy: IfNotPresent
  
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

### カスタムPrometheus設定
```bash
helm install gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --create-namespace \
  --set backend.env.PROMETHEUS_URL=http://prometheus.custom-namespace.svc.cluster.local:9090 \
  --set backend.resources.requests.cpu=500m \
  --set backend.resources.requests.memory=512Mi
```

### NodeSelectorでGPUノード限定
```yaml
backend:
  nodeSelector:
    nvidia.com/gpu: "true"
  tolerations:
    - key: nvidia.com/gpu
      operator: Equal
      value: "true"
      effect: NoSchedule
    
frontend:
  nodeSelector:
    node-type: gpu-node
```

### リソース調整
```yaml
backend:
  resources:
    requests:
      cpu: "250m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"

frontend:
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "200m"
      memory: "256Mi"
```

### セキュリティ設定（個人利用向け）
```yaml
backend:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    runAsGroup: 1001
    fsGroup: 1001
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
    fsGroup: 101
    capabilities:
      drop:
        - ALL
    readOnlyRootFilesystem: true
    allowPrivilegeEscalation: false
```

## ドメイン設定

### .localドメインの使用（開発・個人利用）
```bash
# インストール時にドメイン指定
helm install gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --create-namespace \
  --set backend.env.PROMETHEUS_URL=http://prometheus-server:9090 \
  --set ingress.hosts[0].host=gpu-monitoring.local

# Ingress Controller のIPを確認
kubectl get svc -n ingress-nginx ingress-nginx-controller

# 出力例:
# NAME                       TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)
# ingress-nginx-controller   LoadBalancer   10.96.123.45    192.168.1.100  80:30080/TCP,443:30443/TCP

# /etc/hostsに追加（Linux/Mac）
echo "192.168.1.100 gpu-monitoring.local" | sudo tee -a /etc/hosts

# Windows の場合
# C:\Windows\System32\drivers\etc\hosts に以下を追加:
# 192.168.1.100 gpu-monitoring.local
```

### 外部ドメインの使用
```bash
# 外部ドメインでインストール
helm install gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --create-namespace \
  --set backend.env.PROMETHEUS_URL=http://prometheus-server:9090 \
  --set ingress.hosts[0].host=gpu-monitoring.yourdomain.com

# TLS証明書を使用する場合
helm install gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --create-namespace \
  --set backend.env.PROMETHEUS_URL=http://prometheus-server:9090 \
  --set ingress.hosts[0].host=gpu-monitoring.yourdomain.com \
  --set ingress.tls[0].secretName=gpu-monitoring-tls \
  --set ingress.tls[0].hosts[0]=gpu-monitoring.yourdomain.com
```

## 監視・運用

### ヘルスチェック
```bash
# Backend API ヘルスチェック
kubectl exec -n gpu-monitoring deployment/gpu-monitoring-backend \
  -- wget -qO- http://localhost:8080/api/health

# 成功時のレスポンス例:
# {"success":true,"message":"Service is healthy","data":{"status":"healthy","timestamp":"2024-01-01T12:00:00Z","version":"1.0.0"}}

# Frontend ヘルスチェック
kubectl exec -n gpu-monitoring deployment/gpu-monitoring-frontend \
  -- wget -qO- http://localhost:80/health

# 成功時のレスポンス: "healthy"

# 外部からのヘルスチェック
curl http://gpu-monitoring.local/api/health
curl http://gpu-monitoring.local/health
```

### ログ確認
```bash
# Backend ログ（構造化ログ）
kubectl logs -n gpu-monitoring deployment/gpu-monitoring-backend -f

# ログ例:
# 2024/01/01 12:00:00 Starting GPU Monitoring API Server...
# 2024/01/01 12:00:00 Prometheus URL: http://prometheus-server:9090
# 2024/01/01 12:00:00 Server Port: 8080
# 2024/01/01 12:00:00 Server starting on port 8080
# GET /api/health 200 123µs 192.168.1.1

# Frontend ログ（Nginxアクセスログ）
kubectl logs -n gpu-monitoring deployment/gpu-monitoring-frontend -f

# 全体のログ
kubectl logs -n gpu-monitoring -l app.kubernetes.io/name=k8s-gpu-monitoring-dev -f

# 特定時間範囲のログ
kubectl logs -n gpu-monitoring deployment/gpu-monitoring-backend --since=1h
```

### リソース使用量確認
```bash
# CPU/メモリ使用量
kubectl top pods -n gpu-monitoring

# 出力例:
# NAME                              CPU(cores)   MEMORY(bytes)
# gpu-monitoring-backend-xxx        15m          89Mi
# gpu-monitoring-frontend-xxx       5m           25Mi

# 詳細情報
kubectl describe pods -n gpu-monitoring

# リソース制限の確認
kubectl get pods -n gpu-monitoring -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[0].resources}{"\n"}{end}'
```

### メトリクス確認
```bash
# API経由でGPUメトリクス確認
kubectl exec -n gpu-monitoring deployment/gpu-monitoring-backend \
  -- wget -qO- http://localhost:8080/api/v1/gpu/metrics | jq '.'

# Prometheus接続確認
kubectl exec -n gpu-monitoring deployment/gpu-monitoring-backend \
  -- wget -qO- http://prometheus-server:9090/api/v1/query?query=nvidia_gpu_utilization_percent
```

## アップグレード

### チャート更新
```bash
# リポジトリ更新
helm repo update

# 現在のバージョン確認
helm list -n gpu-monitoring

# アップグレード
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring

# 特定バージョンにアップグレード
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --version 1.0.1

# 設定を保持してアップグレード
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --reuse-values
```

### 設定変更
```bash
# レプリカ数変更
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set backend.replicas=2 \
  --set frontend.replicas=2

# リソース設定変更
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set backend.resources.requests.cpu=500m \
  --set backend.resources.requests.memory=512Mi

# Prometheus URL変更
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set backend.env.PROMETHEUS_URL=http://new-prometheus:9090
```

### イメージタグ更新
```bash
# 新しいイメージタグに更新
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set backend.image.tag=1.0.1 \
  --set frontend.image.tag=1.0.1

# GitHub Container Registryから最新イメージ取得
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set global.imageRegistry=ghcr.io/v01d42/k8s-gpu-monitoring-dev \
  --set backend.image.tag=latest \
  --set frontend.image.tag=latest
```

### ロールバック
```bash
# リリース履歴確認
helm history gpu-monitoring -n gpu-monitoring

# 前のバージョンにロールバック
helm rollback gpu-monitoring 1 -n gpu-monitoring
```

## トラブルシューティング

### よくある問題

#### 1. Prometheus接続エラー
```bash
# Pod内からPrometheus疎通確認
kubectl exec -n gpu-monitoring deployment/gpu-monitoring-backend \
  -- wget -qO- http://prometheus-server:9090/api/v1/query?query=up

# DNS確認
kubectl exec -n gpu-monitoring deployment/gpu-monitoring-backend \
  -- nslookup prometheus-server

# Prometheusサービス確認
kubectl get svc -A | grep prometheus

# Prometheusの名前空間確認
kubectl get svc -n monitoring | grep prometheus
kubectl get svc -n prometheus | grep prometheus

# 正しいPrometheus URLに更新
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set backend.env.PROMETHEUS_URL=http://prometheus-server.monitoring.svc.cluster.local:9090
```

#### 2. Ingress設定問題
```bash
# Ingress確認
kubectl describe ingress -n gpu-monitoring

# Ingress Controller ログ
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller

# Ingress Controller がインストールされているか確認
kubectl get pods -n ingress-nginx

# /etc/hosts確認
cat /etc/hosts | grep gpu-monitoring

# Ingress Controller のIPを再確認
kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

#### 3. Pod起動問題
```bash
# Pod状態詳細確認
kubectl describe pod -n gpu-monitoring -l app.kubernetes.io/name=k8s-gpu-monitoring-dev

# イベント確認
kubectl get events -n gpu-monitoring --sort-by='.lastTimestamp'

# リソース不足確認
kubectl top nodes

# ImagePullBackOff の場合
kubectl describe pod -n gpu-monitoring -l app.kubernetes.io/name=k8s-gpu-monitoring-dev | grep -A10 "Events:"
```

#### 4. イメージPull問題
```bash
# イメージPull詳細確認
kubectl describe pod -n gpu-monitoring -l app.kubernetes.io/name=k8s-gpu-monitoring-dev | grep -A5 "Failed"

# 現在のイメージタグ確認
helm get values gpu-monitoring -n gpu-monitoring | grep tag

# イメージが存在するか確認（GitHub Container Registry）
# https://github.com/V01d42/k8s-gpu-monitoring-dev/pkgs/container/k8s-gpu-monitoring-dev%2Fbackend

# プライベートレジストリの場合はImagePullSecrets設定
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_USERNAME \
  --docker-password=YOUR_TOKEN \
  --namespace gpu-monitoring

helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set global.imagePullSecrets[0].name=ghcr-secret
```

#### 5. フロントエンド表示問題
```bash
# フロントエンドのログ確認
kubectl logs -n gpu-monitoring deployment/gpu-monitoring-frontend

# ブラウザの開発者ツールでネットワークタブを確認
# F12 -> Network -> XHR を確認してAPI通信をチェック

# API通信の確認
curl http://gpu-monitoring.local/api/health
curl http://gpu-monitoring.local/api/v1/gpu/metrics

# フロントエンドからバックエンドへの通信確認
kubectl exec -n gpu-monitoring deployment/gpu-monitoring-frontend \
  -- wget -qO- http://gpu-monitoring-backend:8080/api/health
```

### デバッグコマンド
```bash
# 現在の設定値確認
helm get values gpu-monitoring -n gpu-monitoring

# リリース履歴
helm history gpu-monitoring -n gpu-monitoring

# 生成されたマニフェスト確認
helm get manifest gpu-monitoring -n gpu-monitoring

# ドライラン（実際にはデプロイせずに検証）
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --dry-run --debug

# 全リソース確認
kubectl get all -n gpu-monitoring

# ConfigMapやSecret確認
kubectl get configmap,secret -n gpu-monitoring

# PersistentVolume確認（もし使用している場合）
kubectl get pv,pvc -n gpu-monitoring
```

### パフォーマンス最適化
```bash
# リソース使用量の監視
watch kubectl top pods -n gpu-monitoring

# メモリ不足の場合
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set backend.resources.requests.memory=512Mi \
  --set backend.resources.limits.memory=1Gi

# CPU使用率が高い場合
helm upgrade gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set backend.resources.requests.cpu=500m \
  --set backend.resources.limits.cpu=1000m
```

## カスタマイズ例

### GPU専用ノードに配置
```yaml
# values.yaml
backend:
  nodeSelector:
    nvidia.com/gpu: "true"
  tolerations:
    - key: nvidia.com/gpu
      operator: Equal
      value: "true"
      effect: NoSchedule

frontend:
  # フロントエンドは通常のノードに配置
  nodeSelector: {}
  tolerations: []
```

### 複数環境対応
```yaml
# production.yaml
backend:
  replicas: 2
  resources:
    requests:
      cpu: "500m"
      memory: "512Mi"
    limits:
      cpu: "1000m"
      memory: "1Gi"
  env:
    PROMETHEUS_URL: "http://prometheus-server.monitoring.svc.cluster.local:9090"

frontend:
  replicas: 2
  resources:
    requests:
      cpu: "200m"
      memory: "256Mi"
    limits:
      cpu: "400m"
      memory: "512Mi"

ingress:
  hosts:
    - host: gpu-monitoring.production.com
  tls:
    - secretName: gpu-monitoring-tls
      hosts:
        - gpu-monitoring.production.com
```

### 個人利用向け最適化
```yaml
# personal.yaml（推奨設定）
global:
  imageRegistry: "ghcr.io/v01d42/k8s-gpu-monitoring-dev"

backend:
  replicas: 1
  resources:
    requests:
      cpu: "250m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"

frontend:
  replicas: 1
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "200m"
      memory: "256Mi"

# 個人利用では不要
serviceAccount:
  create: false
rbac:
  create: false

ingress:
  enabled: true
  hosts:
    - host: gpu-monitoring.local
```

## アンインストール

### 完全削除
```bash
# アプリケーション削除
helm uninstall gpu-monitoring --namespace gpu-monitoring

# Namespace削除（他にリソースがない場合）
kubectl delete namespace gpu-monitoring

# /etc/hostsエントリ削除
sudo sed -i '/gpu-monitoring.local/d' /etc/hosts

# 設定ファイルのバックアップ削除（必要に応じて）
rm -f gpu-monitoring-values.yaml
```

### データ保持して削除
```bash
# アプリケーションのみ削除（NamespaceとPVCを保持）
helm uninstall gpu-monitoring --namespace gpu-monitoring

# 後で再インストールする場合
helm install gpu-monitoring gpu-monitoring/k8s-gpu-monitoring-dev \
  --namespace gpu-monitoring \
  --set backend.env.PROMETHEUS_URL=http://prometheus-server:9090
```
