<!-- 環境変数の管理ルール・必要キー一覧・取得方法を記載 -->

# 環境変数管理

> このドキュメントでは、本プロジェクトで使用する環境変数の管理ルールと一覧を記載してください。

## Godot フロントエンド

- `BACKEND_HOST`: `area_3d.gd` と `node_3d.gd` が API を叩く先のホスト名または IP アドレス。既定値は `localhost`
- `BACKEND_PORT`: `area_3d.gd` と `node_3d.gd` が API を叩く先のポート。既定値は `9000`、Quest 2 で `adb reverse tcp:9001 tcp:9000` を使う場合は `9001`
- `SCREAM_WEBSOCKET_URL`: `scream_display.gd` が接続する WebSocket URL。既定値は `ws://localhost:8082/ws`

どちらも未設定の場合は `localhost` 系の既定値を使います。Meta Quest 2 の実機ビルドでは、端末内の `localhost` から PC 側の開発用 API に到達させるため、USB/ADB 経由でポート転送を行います。これにより、Quest 2 側のアプリは同一 Wi-Fi 上の PC に直接接続せず、`adb reverse` 経由でローカルバックエンドを叩けます。

```bash
adb reverse tcp:9001 tcp:9000
adb reverse tcp:8082 tcp:8082
```

ADB リバースを使わない場合は、PC の到達可能な LAN IP を `BACKEND_HOST` と `SCREAM_WEBSOCKET_URL` で明示し、`BACKEND_PORT` は接続先の待ち受けポートに合わせてください。
