<!-- Devbox shellに入った後のフロント/バックエンドのサーバー起動手順を記載 -->

# 2. プロジェクトセットアップ手順

> WSL2 + Devbox がまだ入っていない場合は先に [docs/1_setup_devbox.md](./1_setup_devbox.md) を完了させてください。

---

## STEP 1. Docker Desktop のインストール（Windows）

### 1-1. インストーラーをダウンロード

[Docker Desktop 公式サイト](https://www.docker.com/products/docker-desktop/) から **Windows用インストーラー** をダウンロードして実行してください。

インストール時のオプションはデフォルトのままで問題ありません。

### 1-2. PCを再起動

インストール完了後、PCを再起動してください。

### 1-3. WSL Integration を有効化 ⭐ 重要

Docker Desktop を起動し、以下の手順で WSL2 との連携を有効にします。

1. 右上の歯車アイコン（Settings）をクリック
2. **Resources** → **WSL Integration** を開く
3. 使用している WSL2 ディストリビューション（例: `Ubuntu`）のトグルを **ON** にする
4. **Apply & Restart** をクリック

これにより WSL2 のターミナルから `docker` コマンドが使えるようになります。

### 1-4. 動作確認（WSL2ターミナル）

```bash
docker --version
docker compose version
```

バージョンが表示されれば完了です。

---

## STEP 2. Godot 4 のインストール（Windows）

### 2-1. Steam から Godot 4 をインストール

Steam を開き、**Godot Engine** を検索してインストールしてください。

> **.NET版（Godot Engine - .NET）は不要です。** 通常の **Godot Engine** を選んでください。

### 2-2. godot-vrm アドオンの導入

このプロジェクトでは VRM モデルの表示に [godot-vrm](https://github.com/V-Sekai/godot-vrm) アドオンを使用します。

1. Godot エディタを起動し、`front/` フォルダをプロジェクトとして開く
2. **AssetLib** タブを開き、`vrm` で検索
3. **VRM** アドオンをインストールする
4. **Project** → **Project Settings** → **Plugins** タブで VRM が有効になっていることを確認

---

## STEP 3. リポジトリのクローン（WSL2ターミナル）

```bash
cd ~
git clone <リポジトリURL>
cd hackathon_DDD
```

> `/mnt/c/` 配下への配置は避けてください。WSL2 のホームディレクトリ（`~/`）に置くのが必須です。詳細は [1_setup_devbox.md](./1_setup_devbox.md) を参照。

---

## STEP 4. devbox shell に入る（WSL2ターミナル）

```bash
devbox shell
# shell に入ると自動で go mod tidy が実行されます
```

---

## STEP 5. 環境変数の設定（WSL2ターミナル）

```bash
cp .env.example .env
```

`.env` を開いて必要な値を設定してください。各変数の説明は [docs/env_variables.md](./env_variables.md) を参照。

---

## STEP 6. Dockerコンテナを起動（WSL2ターミナル）

```bash
docker compose up --build
```

以下のコンテナが起動します：

| コンテナ | 役割 | ポート |
|---|---|---|
| `game-server` | Goゲームサーバー | 8080 |
| `target-server` | 攻撃対象（ゲームの敵） | 内部のみ |

---

## STEP 7. Godotプロジェクトを開く（Windows）

Godot エディタを起動し、`front/` フォルダをプロジェクトとして開いてください。

---

## セットアップ完了

ブラウザで `http://localhost:8080/health` にアクセスして `ok` と表示されれば、ゲームサーバーが正常に動いています。

---

## 開発時の起動手順（毎回）

```bash
# WSL2ターミナルで
devbox shell
docker compose up
```

Godot エディタは Windows 側で起動してください。

---

## トラブルシューティング

| 症状 | 確認箇所 |
|---|---|
| `docker: command not found` | Docker Desktop の WSL Integration が ON になっているか確認（STEP 1-3） |
| `docker compose up` でビルドエラー | `server/` や `target/` の Dockerfile が存在するか確認 |
| Godot で VRM が表示されない | godot-vrm アドオンが有効になっているか確認（STEP 2-2） |
| WSL2 関連のエラー | [docs/setup_error_handling.md](./setup_error_handling.md) を参照 |
