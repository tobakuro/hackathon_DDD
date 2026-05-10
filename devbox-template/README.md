# TemplateDevboxProject

Devboxを用いた開発プロジェクトのテンプレートリポジトリです。  
このテンプレからプロジェクトを始めることで、環境構築の手間を最小化し、すぐに開発に着手できます。

> 📂 リポジトリ: https://github.com/tobakuro/TemplateDevboxProject

---

## 🚀 はじめに

### 1. リポジトリの取得

GitHub上で「**Use this template**」ボタンから新規リポジトリを作成するか、直接cloneします。

```bash
git clone https://github.com/tobakuro/TemplateDevboxProject.git <あなたのプロジェクト名>
cd <あなたのプロジェクト名>
```

### 2. 環境構築 (WSL → Devbox)

WSLのインストールからDevbox shellに入るまでの手順は、以下のドキュメントを参照してください。

📖 [docs/1_setup_devbox.md](./docs/1_setup_devbox.md)

詰まったら → 📖 [docs/setup_error_handling.md](./docs/setup_error_handling.md)

### 3. プロジェクト固有のセットアップ

Devbox shellに入った後、フロント・バックエンドの起動手順は以下を参照してください。

📖 [docs/2_setup.md](./docs/2_setup.md)

### 4. よく使うコマンド

開発中によく使うコマンドは以下にまとまっています。

📖 [docs/commands.md](./docs/commands.md)

---

## 📚 ドキュメント一覧

| ファイル | 内容 |
|---|---|
| [1_setup_devbox.md](./docs/1_setup_devbox.md) | WSL〜Devbox shellまでの環境構築手順 |
| [2_setup.md](./docs/2_setup.md) | フロント/バックエンドのサーバー起動手順 (プロジェクト固有) |
| [commands.md](./docs/commands.md) | よく使うコマンド集 |
| [setup_error_handling.md](./docs/setup_error_handling.md) | WSL/Devbox/Nix関連のエラー対処集 |
| [architecture.md](./docs/architecture.md) | システム構成・技術スタック |
| [directory_structure.md](./docs/directory_structure.md) | ディレクトリ構成と各フォルダの役割 |
| [git_workflow.md](./docs/git_workflow.md) | ブランチ戦略・コミット規約・PR運用 |
| [coding_rules.md](./docs/coding_rules.md) | コーディング規約 |
| [env_variables.md](./docs/env_variables.md) | 環境変数の管理ルール |

---

## 📁 リポジトリ構成

```
.
├── .github/
│   ├── ISSUE_TEMPLATE/         # Issueテンプレート (bug/feature/task)
│   ├── PULL_REQUEST_TEMPLATE.md
│   └── workflows/              # GitHub Actions (CI/CD)
│       ├── ci.yml
│       └── cd.yml
├── docs/                       # プロジェクトドキュメント (全mdはここに集約)
│   ├── 1_setup_devbox.md
│   ├── 2_setup.md
│   ├── commands.md
│   ├── setup_error_handling.md
│   ├── architecture.md
│   ├── directory_structure.md
│   ├── git_workflow.md
│   ├── coding_rules.md
│   └── env_variables.md
├── .env.example                # 環境変数のサンプル
├── .gitignore
├── devbox.json                 # Devboxの設定ファイル (使用技術/パッケージ定義)
└── README.md
```

---

## 🛠️ 使用技術の追加方法

`devbox.json` の `packages` 配列に必要なパッケージを追記してください。

```json
{
  "packages": [
    "nodejs@20",
    "python@3.12",
    "go@latest",
    "postgresql@17",
    "redis@latest"
  ]
}
```

利用可能なパッケージは [Nixhub Search](https://www.nixhub.io/) から検索できます。

---

## ✨ 含まれているもの

- ✅ **Devbox設定** - プロジェクトごとに独立した開発環境
- ✅ **GitHub Actions CI/CD** - push/PR時の自動ビルド・テスト + main/タグでのデプロイ枠
- ✅ **Issueテンプレート** - bug報告 / 機能要望 / タスク
- ✅ **PRテンプレート**
- ✅ **充実したドキュメント** - セットアップから運用ルールまで
- ✅ **エラーハンドリング集** - WSL/Devbox/Nixで実際に踏んだエラーと解決策
