<!-- プロジェクトのディレクトリ構成と各フォルダの役割を記載 -->

# ディレクトリ構成

```
.
├── front/                   # Godot 4 プロジェクト（ゲームフロントエンド）
│
├── api/                     # Go APIサーバー（現在はDocker動作）
│   ├── main.go
│   ├── handler.go
│   ├── Nowtarget.go
│   ├── prometheus.go
│   └── types.go
│
├── attacker/                # 攻撃コンテナ（ゲームの攻撃役）
│   ├── main.go
│   └── Dockerfile
│
├── target/                  # 攻撃対象コンテナ（ゲームの敵サーバー）
│   ├── main.go
│   └── Dockerfile
│
├── k8s/                     # K8s移行用ソースコード（Docker版と並行管理中）
│   ├── api/                 # K8s対応版 api（watchIsAlive・restartTargetServer追加済み）
│   └── target/              # K8s対応版 target（/healthエンドポイント追加済み）
│
├── manifests/               # Kubernetes Manifest
│   ├── game/                # ゲーム関連Pod
│   │   ├── attacker-deployment.yaml
│   │   └── target-server-deployment.yaml  # Liveness Probe設定済み
│   └── monitoring/          # 監視スタック（cAdvisor + Prometheus）
│       ├── cadvisor-deployment.yaml
│       ├── cadvisor-service.yaml
│       ├── cadvisor-claim3-persistentvolumeclaim.yaml
│       ├── prometheus-deployment.yaml
│       ├── prometheus-service.yaml
│       └── prometheus-cm0-configmap.yaml
│
├── docs/                    # ドキュメント
├── docker-compose.yml       # コンテナ構成（K8s完全移行まで使用）
├── go.mod                   # Goモジュール（ルートで一元管理）
├── devbox.json              # 開発環境（Go・kubectl・kind・kompose・k9s・helm管理）
└── .env.example             # 環境変数サンプル
```

## 各コンポーネントの役割

| ディレクトリ | 担当 | 技術 | 状態 |
|---|---|---|---|
| `front/` | ゲームUI・3D・VRM・当たり判定 | Godot 4 + godot-vrm | 開発中 |
| `api/` | APIサーバー・ゲームロジック | Go | Docker動作中 |
| `attacker/` | 攻撃コンテナ（target-server を攻撃） | Go + Docker | Docker動作中 |
| `target/` | 攻撃対象（リソース制限付き） | Go + Docker | Docker動作中 |
| `k8s/` | K8s移行用ソースコード | Go | 移行待ち |
| `manifests/` | Kubernetes Manifest一式 | YAML | target-serverのみ適用中 |
