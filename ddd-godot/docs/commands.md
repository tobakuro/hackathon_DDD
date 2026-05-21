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
docker compose logs -f game-server

# リソース使用状況をリアルタイム確認（攻撃の効果確認に使う）
docker stats

# ターゲットコンテナを手動で再起動（復旧テスト）
docker compose restart target-server
```

## Go サーバー

```bash
# devbox shell に入る
devbox shell

# 依存関係を整理
cd server && go mod tidy

# サーバーを直接起動（開発時）
cd server && go run main.go

# ビルド
cd server && go build -o bin/server .

# テスト
cd server && go test ./...
```

## よく使う確認コマンド

```bash
# ゲームサーバーの疎通確認
curl http://localhost:8080/health

# コンテナ一覧
docker compose ps

# ターゲットコンテナのリソース制限確認
docker inspect hackathon_ddd-target-server-1 | grep -A 10 "HostConfig"
```
