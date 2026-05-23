<!-- システム全体のアーキテクチャ・構成図・使用技術スタックを記載 -->

# アーキテクチャ

## 現在の構成（移行中）

現在はtarget-serverのみKubernetesに移行済み。api・attackerはDockerで稼働中。

```
┌─────────────────────────────────────────────────────┐
│                Docker Compose                       │
│                                                     │
│  ┌──────────┐  /scale  ┌──────────────────────┐    │
│  │  Godot   │ ──────▶  │    api (port 9000)   │    │
│  │ (Client) │  /attack │                      │    │
│  │          │ ──────▶  │ ・streamingTarget()  │    │
│  │          │  /restart│ ・watchIsAlive()      │    │
│  └──────────┘ ──────▶  └──────────┬───────────┘    │
│                                   │ docker exec     │
│                         ┌─────────▼──────────┐      │
│                         │  attacker (×N)     │      │
│                         └─────────┬──────────┘      │
└───────────────────────────────────┼─────────────────┘
                                    │ HTTP攻撃
                    kubectl API     │
          ┌─────────────────────────┼──────────────────┐
          │         Kubernetes      │                  │
          │                ┌────────▼───────────────┐  │
          │                │  target-server (Pod)   │  │
          │                │  ・Liveness Probe      │  │
          │                │    /health を監視       │  │
          │                └────────────────────────┘  │
          └────────────────────────────────────────────┘
```

## 自動復旧の仕組み

再起動が発生するルートは3つある。

```
【ルート①】IsAliveトリガー（自動）
攻撃 → target落ちる → streamingTarget()が検知
    → IsAlive=false → watchIsAlive()が検知
    → K8s APIでRolling Restart → Pod再起動

【ルート②】手動（フロントから）
クライアントが任意タイミングで
POST /restart → K8s APIでRolling Restart → Pod再起動

【ルート③】Liveness Probe（K8s自律）
プロセスフリーズ → /healthに無応答
→ K8sが直接Pod再起動（apiを介さない）
```

## 今後の移行方針

他メンバーの「ターゲットダウンの仕組み」完成後、以下の順で全部K8sに統一する。

```
1. k8s/api/ と k8s/target/ を本家 api/ target/ に統合
2. attacker/ も含めて manifests/ に移行
3. docker-compose.yml を廃止
```

## 使用技術スタック

| レイヤー | 技術 |
|---|---|
| フロントエンド | Godot 4 + godot-vrm |
| APIサーバー | Go |
| 攻撃コンテナ | Go + Docker |
| 攻撃対象 | Go + Kubernetes (kind) |
| コンテナ管理 | Docker Compose（移行中） → Kubernetes |
| 監視 | cAdvisor + Prometheus |
| 開発環境 | Devbox (Nix) |
