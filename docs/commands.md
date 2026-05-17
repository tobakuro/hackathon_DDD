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

# ビルド（api・attacker 両方）
devbox run build

# テスト（カバレッジ付き）
devbox run test

# lint
devbox run lint

# lint 自動修正
devbox run lint:fix

# フォーマット
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
