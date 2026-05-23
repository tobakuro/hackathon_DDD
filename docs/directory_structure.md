<!-- プロジェクトのディレクトリ構成と各フォルダの役割を記載 -->

# ディレクトリ構成

```
.
├── front/                   # Godot 4 プロジェクト（ゲームフロントエンド）
│
├── api/                     # Go APIサーバー
│   └── main.go
│
├── attacker/                # 攻撃コンテナ（ゲームの攻撃役）
│   ├── main.go
│   └── Dockerfile
│
├── target/                  # 攻撃対象コンテナ（ゲームの敵サーバー）
│   ├── main.go
│   └── Dockerfile
│
├── docs/                    # ドキュメント
├── docker-compose.yml       # コンテナ構成（attacker + target-server）
├── go.mod                   # Goモジュール（ルートで一元管理）
├── devbox.json              # 開発環境（Go・golangci-lint管理）
└── .env.example             # 環境変数サンプル
```

## 各コンポーネントの役割

| ディレクトリ | 担当 | 技術 |
|---|---|---|
| `front/` | ゲームUI・3D・VRM・当たり判定 | Godot 4 + godot-vrm |
| `api/` | APIサーバー・ゲームロジック | Go |
| `attacker/` | 攻撃コンテナ（target-server を攻撃） | Go + Docker |
| `target/` | 攻撃対象（リソース制限付き） | Go + Docker |
