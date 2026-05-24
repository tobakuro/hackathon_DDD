# API リファレンス

司令塔APIサーバー（ポート: `9000`）が提供するエンドポイントの一覧です。
フロントエンド（Godot）はここに HTTPリクエストを送ってゲームを制御します。

---

## エンドポイント一覧

### `POST /attack`

稼働中のattacker全台に攻撃命令を送ります。

**リクエスト**

```json
{
  "target": "all",
  "count": 5
}
```

| フィールド | 型 | 説明 |
|---|---|---|
| `target` | string | 現在は `"all"` のみ対応（全attacker一斉攻撃） |
| `count` | int | 各attackerが target-server を叩く回数 |

**レスポンス**

```
3 体のattackerに攻撃命令を下しました。
```

**何ができるか**

attackerコンテナ全台に指定回数の攻撃を並列で実行させます。
attackerが target-server の `/` を叩くと、target-server 側でllama.cppが悲鳴を生成し、WebSocket経由でGodotに配信されます。

---

### `PUT /scale`

attackerコンテナの台数を変更します。

**リクエスト**

```json
{
  "count": 3
}
```

| フィールド | 型 | 説明 |
|---|---|---|
| `count` | int | attackerの台数 |

**レスポンス**

```
コンテナを 3 個に変更しましたよ
```

**何ができるか**

攻撃の強度をゲーム中に動的に変更できます。台数を増やすほど target-server への負荷が上がります。

---

### `POST /restart`

target-server Pod を再起動します。

**リクエスト**

ボディなし。

**レスポンス**

```json
{
  "message": "Podの再起動を開始しました"
}
```

**何ができるか**

ゲームオーバー後に target-server をリセットします。Kubernetes の Rolling Restart として発行されるため、新しい Pod が起動してから古い Pod が終了します。

> target-server の死亡検知（`IsAlive=false`）は APIサーバーが自動で監視しており、検知時に自動でこの再起動が発行されます。フロントから手動で呼ぶことも可能です。

---

### `GET /now-target`

target-server の現在のリソース使用率を取得します。

**レスポンス**

```json
{
  "memory": 72.4,
  "cpu": 48.2,
  "is_alive": true
}
```

| フィールド | 型 | 説明 |
|---|---|---|
| `memory` | float64 | メモリ使用率（%） |
| `cpu` | float64 | CPU使用率（%） |
| `is_alive` | bool | target-server が生存中かどうか |

**何ができるか**

HPゲージなどのUI表示に使います。攻撃を受けるほど CPU/メモリが上昇し、`is_alive` が `false` になるとゲームオーバーです。

---

### `GET /attack-history`

target-server のリソース使用率の時系列データを取得します。

> ⚠️ 現在はスタブ実装のため固定値を返します。

**レスポンス**

```json
[
  { "timestamp": "2024-06-01T12:00:00Z", "memory": 70.2, "cpu": 55.1 },
  { "timestamp": "2024-06-01T12:01:00Z", "memory": 80.5, "cpu": 65.3 },
  { "timestamp": "2024-06-01T12:02:00Z", "memory": 90.1, "cpu": 75.0 }
]
```

**何ができるか**

リザルト画面などで攻撃の推移をグラフ表示することを想定しています。

---

## WebSocket `/ws`（target-server 直接）

司令塔APIではなく、**target-server（ポート: `8080`）** が提供するエンドポイントです。

```
ws://target-server:8080/ws
```

**何ができるか**

攻撃を受けるたびに target-server からリアルタイムでイベントが配信されます。Godotはここに接続して悲鳴テキストや音声データを受け取ります。

**受信するイベントの形式**

```json
{ "id": "1", "type": "start", "hit": 3 }
{ "id": "1", "type": "char",  "char": "ギ" }
{ "id": "1", "type": "char",  "char": "ャ" }
{ "id": "1", "type": "end" }
```

| `type` | タイミング | 内容 |
|---|---|---|
| `start` | 攻撃受信時 | 悲鳴生成開始。累計ヒット数（`hit`）を含む |
| `char` | 生成中（逐次） | llama.cppが1文字生成するたびに配信 |
| `end` | 生成完了時 | 悲鳴テキストの終端。音声データ生成完了を示す |
| `error` | エラー時 | エラーメッセージ（`message`）を含む |
