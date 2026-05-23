<!-- 開発でよく使うコマンドをまとめて記載 -->

# よく使うコマンド集

> すべてのコマンドは WSL2 の devbox shell 内で実行してください。
> Godot エディタは Windows 側で起動します。

## Docker

```bash
# コンテナをビルドして起動
docker compose up --build

# バックグラウンドで起動
docker compose up -d

# コンテナを停止
docker compose down

# コンテナのログを確認
docker compose logs -f target-server
docker compose logs -f attacker

# リソース使用状況をリアルタイム確認（攻撃の効果確認に使う）
docker stats

# ターゲットコンテナを手動で再起動（復旧テスト）
docker compose restart target-server
```

## Go (api / attacker)

```bash
# devbox shell に入る
devbox shell

# 依存関係を整理
go mod tidy
```

### ビルド・テスト

```bash
# ビルド（api・attacker 両方）
devbox run build

# テスト（カバレッジ付き）
devbox run test
```

### Lint・フォーマット・型チェック

```bash
# lint（golangci-lint）
devbox run lint

# lint 自動修正
devbox run lint:fix

# フォーマット（gofmt で上書き）
devbox run format

# フォーマットチェックのみ（CI と同じ）
devbox run format:check

# 型チェック（go vet）
devbox run typecheck
```

## よく使う確認コマンド

```bash
# コンテナ一覧
docker compose ps

# ターゲットコンテナのリソース制限確認
docker inspect hackathon_ddd-target-server-1 | grep -A 10 "HostConfig"
```

## Kubernetes

```bash
# クラスター作成（初回のみ）
kind create cluster

# クラスター情報確認
kubectl cluster-info --context kind-kind
```

### イメージのビルドとクラスターへの読み込み（初回 or コード変更時）

kindはローカルのDockerイメージを直接参照できないため、明示的に読み込む必要がある。

```bash
# イメージをビルド
docker build -t target-server ./target
docker build -t attacker ./attacker

# kindクラスターにイメージを読み込む
kind load docker-image target-server
kind load docker-image attacker
```

### Manifestの適用・削除

```bash
# Manifestを適用（ゲーム関連）
kubectl apply -f manifests/game/

# Manifestを適用（監視スタック）
kubectl apply -f manifests/monitoring/

# 削除
kubectl delete -f manifests/game/
```

### 状態確認・デバッグ

```bash
# Pod一覧と状態確認
kubectl get pods

# Pod のログを確認
kubectl logs -f <pod-name>

# Pod の詳細確認（Liveness Probe状態など）
kubectl describe pod <pod-name>

# TUIで全リソースを確認
k9s
```

### コード変更後の反映

```bash
docker build -t target-server ./target
kind load docker-image target-server
kubectl rollout restart deployment/target-server
```

### クラスター削除

```bash
kind delete cluster
```
