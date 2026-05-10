# じょぎBANK 設計ドキュメント

> Discordサークル「じょぎ」内専用通貨・銀行bot  
> 作成日: 2026-04-12

---

## 1. 概要・目的

じょぎ内では部員が作成した娯楽系bot（チンチロ・ガチャ・スロット）が稼働しており、それぞれが独自ポイントを使用している。これらを**共通通貨で統一**することで、サークル内の経済圏を一本化する。

じょぎBANKは通貨の発行・管理を担うbotで、銀行をテーマにした娯楽機能（借金・資産運用）も備える。

---

## 2. 技術スタック

| 項目 | 採用技術 |
|---|---|
| 言語 | TypeScript |
| Discordライブラリ | discord.js v14 |
| HTTPサーバー | Express（または Fastify） |
| DB | SQLite（小規模）または PostgreSQL |

---

## 3. アーキテクチャ：他botとの連携

**方式：じょぎBANKがHTTP APIサーバーを兼ねる**

他bot（チンチロ・ガチャ・スロット等）はHTTPリクエストでじょぎBANKの残高を操作する。

```
チンチロbot / ガチャbot / スロットbot
    │
    │  POST /api/balance/deduct
    │  POST /api/balance/grant
    │  GET  /api/balance/:userId
    ▼
じょぎBANK（Discord bot + REST APIサーバー）
    │
    ▼
Database（残高・履歴・ローン等）
```

### 認証

- 他botごとに**APIキーを発行**し、リクエストヘッダー `X-API-Key` で認証
- キーの発行・無効化は `/admin apikey generate` で管理者（開発者）が行う
- キー漏洩時は1件だけ無効化可能

---

## 4. 機能一覧

### 4-1. 通貨基本機能

| コマンド | 説明 |
|---|---|
| `/balance` | 自分の残高確認 |
| `/transfer @user 金額` | ユーザー間送金 |
| `/history` | 取引履歴（typeで種別フィルタ可） |
| `/daily` | デイリーボーナス受け取り |
| `/ranking` | サークル内資産ランキング |

- 通貨の単位名は未定（例：JC / じょぎコイン）
- 初回ウェルカムボーナスの有無は未定

### 4-2. 借金・貸付機能

| コマンド | 説明 |
|---|---|
| `/loan request @user 金額` | 借金申請（相手がDiscordボタンで承認/拒否） |
| `/loan list` | 自分の借入・貸付一覧 |
| `/loan repay @user 金額` | 返済（分割返済可） |
| `/loan status` | 利息込みの残債確認 |

#### 借金ルール（確定）

| 項目 | 内容 |
|---|---|
| 金利方式 | **単利・日利1%** |
| 利息計算式 | `利息 = 元本 × 0.01 × 経過日数` |
| 借金上限 | あり（`users.loan_limit` で個別管理） |
| 残高 | **マイナス残高を許容**（踏み倒し後） |
| 踏み倒し | 返済期限内に返済できなかった場合、一定期間**コマンド全制限**（BANライト） |
| 督促 | 返済期限超過でDiscord通知 |
| 保証人 | なし |

- `loans.status` は `pending` / `active` / `defaulted` / `repaid` の4種

### 4-3. 資産運用機能

#### 株・仮想通貨風相場

| コマンド | 説明 |
|---|---|
| `/invest buy 銘柄 枚数` | 株購入 |
| `/invest sell 銘柄 枚数` | 売却 |
| `/invest chart 銘柄` | 価格チャート表示 |

- じょぎ独自銘柄を複数用意（内輪ネタ系の名前推奨）
- 一定間隔（例：1時間ごとのcron）でランダム＋トレンドで価格変動
- `stock_history` テーブルに価格を蓄積し、チャートに使用

#### 定期預金

| コマンド | 説明 |
|---|---|
| `/deposit 金額 日数` | 定期預金（期間中引き出し不可） |

- 日利0.3%（借金金利より低め）
- 満期で元本＋利息が自動返還

#### ギャンブル

| コマンド | 説明 |
|---|---|
| `/gamble 金額` | ハイリスク即時決済 |

- 勝率・倍率は実装時に設計（チンチロ等との差別化として超高倍率低確率も検討）

### 4-4. 管理者機能（開発者限定）

| コマンド | 説明 |
|---|---|
| `/admin grant @user 金額` | 通貨付与（イベント報酬・初期配布等） |
| `/admin take @user 金額` | 通貨没収（不正対応） |
| `/admin forgive @user` | 借金チャラ（特別措置） |
| `/admin ban @user` | コマンド利用停止 |
| `/admin apikey generate bot名` | 他bot用APIキー発行 |

---

## 5. DBスキーマ

### `users`
```sql
discord_id   TEXT PRIMARY KEY   -- DiscordユーザーID
balance      INTEGER            -- 残高（マイナス許容）
loan_limit   INTEGER            -- 借金上限（デフォルト値はアプリ側で定義）
banned_until TIMESTAMP          -- NULL以外ならコマンド全制限中
created_at   TIMESTAMP
```

### `transactions`
```sql
id           INTEGER PRIMARY KEY
from_user    TEXT REFERENCES users  -- 送金元（システム付与ならNULL可）
to_user      TEXT REFERENCES users  -- 受取先
amount       INTEGER
type         TEXT  -- transfer / daily / loan_repay / invest / gamble / admin 等
created_at   TIMESTAMP
```

### `loans`
```sql
id            INTEGER PRIMARY KEY
lender_id     TEXT REFERENCES users  -- 貸す側
borrower_id   TEXT REFERENCES users  -- 借りる側
principal     INTEGER                -- 元本
interest_rate REAL                   -- 0.01 固定（単利）
days_elapsed  INTEGER                -- 経過日数（バッチで更新）
status        TEXT                   -- pending / active / defaulted / repaid
due_date      TIMESTAMP
created_at    TIMESTAMP
```

### `stocks`
```sql
id            INTEGER PRIMARY KEY
symbol        TEXT    -- 銘柄コード（例: JGI）
name          TEXT    -- 銘柄名（例: じょぎコイン先物）
current_price INTEGER
updated_at    TIMESTAMP
```

### `stock_holdings`
```sql
id            INTEGER PRIMARY KEY
user_id       TEXT REFERENCES users
stock_id      INTEGER REFERENCES stocks
quantity      INTEGER
avg_buy_price INTEGER
```

### `stock_history`
```sql
id          INTEGER PRIMARY KEY
stock_id    INTEGER REFERENCES stocks
price       INTEGER
recorded_at TIMESTAMP
```

### `deposits`
```sql
id         INTEGER PRIMARY KEY
user_id    TEXT REFERENCES users
amount     INTEGER
rate       REAL      -- 0.003 固定
matured_at TIMESTAMP
status     TEXT      -- active / matured
```

### `api_keys`
```sql
id         INTEGER PRIMARY KEY
key        TEXT UNIQUE  -- 発行済みAPIキー
bot_name   TEXT         -- 対象bot名（例: chinchiro-bot）
created_at TIMESTAMP
```

---

## 6. 実装ロードマップ（推奨順）

1. **DB設計・初期化スクリプト**（`schema.sql` / Prismaスキーマ等）
2. **通貨基本コマンド**（`/balance` `/transfer` `/daily` `/history` `/ranking`）
3. **HTTP APIサーバー**（Express + APIキー認証ミドルウェア）
4. **他botとの連携テスト**（チンチロ等と繋いで最小動作確認）
5. **借金・貸付機能**（申請フロー・単利計算・督促バッチ）
6. **資産運用機能**（株相場・定期預金・ギャンブル）

> ステップ4まで早めに動かすと、サークルメンバーのモチベーションが上がりやすい。

---

## 7. 未決定事項

- 通貨の名称・単位
- 初回ウェルカムボーナスの有無・金額
- デイリーボーナスの金額
- 借金上限のデフォルト値
- 株銘柄の名称・種類・価格変動ロジックの詳細
- ギャンブルの勝率・倍率設定
- 踏み倒し時のコマンド制限期間
