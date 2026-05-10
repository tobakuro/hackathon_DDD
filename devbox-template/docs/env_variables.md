<!-- 環境変数の管理ルール・必要キー一覧・取得方法を記載 -->

# 環境変数管理

このドキュメントでは、本プロジェクトで使用する環境変数の管理ルールと一覧を記載します。

---

## 📁 ファイル構成

| ファイル | 用途 | Gitコミット |
|---|---|---|
| `.env` | ローカル開発用 | ❌ commitしない |
| `.env.local` | ローカル個別設定 | ❌ commitしない |
| `.env.example` | サンプル(キーのみ・値は空) | ✅ commitする |
| `.env.production` | 本番用 (秘匿情報はSecretsに) | ❌ commitしない |

> 🔒 **秘匿情報 (APIキー・パスワードなど) は絶対にcommitしない。**  
> 万が一commitしてしまった場合は即座にキーをローテートする。

---

## 🔑 必要な環境変数一覧

> プロジェクトに合わせて以下の表を埋めてください。

| キー | 説明 | 例 | 取得元 |
|---|---|---|---|
| `EXAMPLE_API_KEY` | サンプルAPIキー | `xxxxx` | サービス管理画面から発行 |
| `DATABASE_URL` | DB接続文字列 | `postgres://user:pass@localhost:5432/db` | ローカル/本番DB |

---

## 🚀 セットアップ手順

1. `.env.example` をコピーして `.env` を作成

```bash
cp .env.example .env
```

2. 各キーの値を埋める (チームで共有されているSecretsを参照)

3. アプリを起動して動作確認

---

## ☁️ 本番環境

本番用の環境変数は、デプロイ先サービスのSecrets/環境変数機能で管理:

- **GitHub Actions**: Settings → Secrets and variables → Actions
- **Vercel / Cloudflare**: ダッシュボードのEnvironment Variables
- **AWS / GCP**: Secrets Manager / Secret Manager

---

## 🆘 トラブル時

- 環境変数が反映されない → アプリを再起動 (Vite/Next.jsはホットリロードで読み直さない場合あり)
- キーが漏洩したかも → 即座にキーをローテート + Git履歴を確認 (`git log --all -p | grep <キー>`)
