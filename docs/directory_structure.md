<!-- プロジェクトのディレクトリ構成と各フォルダの役割を記載 -->

# ディレクトリ構成

```
.
├── client/                  # Godot 4 プロジェクト（ゲームフロントエンド）
│   ├── assets/              # VRMモデル・テクスチャ・サウンド
│   └── scenes/              # Godotシーンファイル
│
├── server/                  # Go ゲームサーバー
│   ├── main.go              # エントリポイント
│   ├── go.mod
│   └── Dockerfile
│
├── target/                  # 攻撃対象コンテナ（ゲームの敵サーバー）
│   └── Dockerfile
│
├── docs/                    # ドキュメント
├── docker-compose.yml       # コンテナ構成（game-server + target-server）
├── devbox.json              # 開発環境（Go管理）
└── .env.example             # 環境変数サンプル
```

## 各コンポーネントの役割

| ディレクトリ | 担当 | 技術 |
|---|---|---|
| `client/` | ゲームUI・3D・VRM・当たり判定 | Godot 4 + godot-vrm |
| `server/` | WebSocket・Docker操作・ゲームロジック | Go |
| `target/` | 攻撃対象（リソース制限付き） | Docker |
