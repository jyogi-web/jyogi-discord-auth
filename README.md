# じょぎメンバー認証システム

Discord OAuth2を使用したじょぎメンバー専用の認証システム。Discordアカウントでログインし、じょぎサーバーのメンバーシップを確認後、JWTを発行する。他の内製ツールがSSOとして利用できる認証基盤を提供します。

## 概要

じょぎ内製ツール作成にあたり基盤となる認証システムです。Discordがメインのチャットツールであるため、DiscordアカウントをIdPとして活用し、SSOを実現します。

### アーキテクチャ

1. **Identity Provider (IdP)**: Discord（ユーザー情報、所属サーバーの管理）
2. **Auth Server（じょぎ認証）**: Discord OAuth2を実行し、ユーザーが「じょぎメンバーであるか」を判定。独自のアクセストークン（JWT）を発行する認証基盤。
3. **Client Apps（各内製ツール）**: 「じょぎ認証」にログインを委ねるアプリケーション群。

## 機能

- **Discord OAuth2ログイン**: Discordアカウントでログイン
- **メンバーシップ確認**: じょぎサーバーメンバーのみアクセス許可
- **JWT発行**: 認証成功後、アクセストークン（JWT）を発行
- **SSO対応**: 他の内製ツールがこの認証サーバーを利用可能
- **セッション管理**: ユーザーのログイン状態を管理
- **プロフィール同期**: Discord自己紹介チャンネルから自動的にメンバープロフィールを取得・更新

## 技術スタック

- **言語**: Go 1.23+
- **データベース**: SQLite（将来PostgreSQLへの移行可能な設計）
- **認証**: Discord OAuth2, JWT
- **デプロイ**: Google Cloud Functions（推奨）、Railway

## 前提条件

1. Go 1.23以上がインストールされていること
2. Discord Developer Portalでアプリケーションを作成済み
3. じょぎDiscordサーバーのサーバーIDを取得済み

## クイックスタート

### 1. リポジトリのクローン

```bash
git clone https://github.com/jyogi-web/jyogi-discord-auth.git
cd jyogi-discord-auth
```

### 2. 依存関係のインストール

```bash
go mod download
```

### 3. 環境変数の設定

`.env.example`をコピーして`.env`ファイルを作成し、必要な環境変数を設定してください：

```bash
cp .env.example .env
```

`.env`ファイルを編集して以下の値を設定：

- `DISCORD_CLIENT_ID`: Discord Developer Portalで取得したClient ID
- `DISCORD_CLIENT_SECRET`: Discord Developer Portalで取得したClient Secret
- `DISCORD_REDIRECT_URI`: OAuth2リダイレクトURI（開発環境では`http://localhost:8080/auth/callback`）
- `DISCORD_GUILD_ID`: じょぎDiscordサーバーのサーバーID
- `DISCORD_BOT_TOKEN`: Discord Botトークン（プロフィール同期用）
- `DISCORD_PROFILE_CHANNEL`: 自己紹介チャンネルID
- `JWT_SECRET`: JWT署名用の秘密鍵（最低32文字）

### 4. データベースのマイグレーション

```bash
./scripts/migrate.sh
```

### 5. サーバーの起動

```bash
go run cmd/server/main.go
```

サーバーが起動したら、ブラウザで `http://localhost:8080` にアクセスしてください。

## Discord Developer Portal設定

1. [Discord Developer Portal](https://discord.com/developers/applications)でアプリケーションを作成
2. OAuth2設定:
   - Redirect URIs: `http://localhost:8080/auth/callback`（開発環境）
   - Scopes: `identify`, `guilds.members.read`
3. Client IDとClient Secretを取得して`.env`に設定

## Docker開発環境

Dockerを使用することで、環境構築を簡単にし、チーム全体で統一された開発環境を利用できます。

### 前提条件

- Docker & Docker Compose がインストールされていること

### Docker Composeで起動

```bash
# 環境変数の設定
cp .env.example .env
# .envファイルを編集して必要な値を設定

# Docker Composeでビルド＆起動
docker-compose up -d

# ログを確認
docker-compose logs -f

# 停止
docker-compose down
```

### Makefileを使った開発

開発用のコマンドを簡単に実行できるMakefileを用意しています。

```bash
# ヘルプを表示
make help

# ローカル環境でビルド
make build

# ローカル環境でサーバー起動
make run

# テスト実行
make test

# コードフォーマット
make fmt

# 静的解析
make vet

# Docker環境でビルド
make docker-build

# Docker環境で起動
make docker-up

# Docker環境で停止
make docker-down

# Dockerログ表示
make docker-logs

# マイグレーション実行
make migrate-up

# マイグレーションロールバック
make migrate-down

# 新しいマイグレーション作成
make migrate-create NAME=add_users_table

# プロフィール同期（1回のみ）
make sync-profiles

# プロフィール同期（定期実行）
make sync-profiles-daemon
```

### golang-migrateを使ったマイグレーション管理

このプロジェクトでは、[golang-migrate](https://github.com/golang-migrate/migrate)を使用してデータベースマイグレーションを管理しています。

#### インストール

**macOS:**

```bash
brew install golang-migrate
```

**Linux:**

```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

**Go install:**

```bash
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

#### マイグレーションコマンド

```bash
# マイグレーション実行
migrate -path migrations -database "sqlite3://./jyogi_auth.db" up

# 1つロールバック
migrate -path migrations -database "sqlite3://./jyogi_auth.db" down 1

# 全てロールバック
migrate -path migrations -database "sqlite3://./jyogi_auth.db" down -all

# マイグレーションバージョン確認
migrate -path migrations -database "sqlite3://./jyogi_auth.db" version

# 新しいマイグレーション作成
migrate create -ext sql -dir migrations -seq add_users_table
```

## 開発

### コードフォーマット

```bash
gofmt -w .
```

### 静的解析

```bash
go vet ./...
```

### テスト実行

```bash
go test ./...
```

### テストカバレッジ

```bash
go test -cover ./...
```

## プロフィール同期機能

Discord自己紹介チャンネルからメンバーのプロフィール情報を自動的に取得・保存する機能を提供します。

### セットアップ

1. Discord Developer Portalでボットを作成
2. 以下の権限を付与:
   - `Read Messages/View Channels`
   - `Read Message History`
3. ボットトークンを`.env`の`DISCORD_BOT_TOKEN`に設定
4. 自己紹介チャンネルIDを`.env`の`DISCORD_PROFILE_CHANNEL`に設定

### 使用方法

#### 1回だけ実行（手動同期）

```bash
make sync-profiles
```

または:

```bash
go run ./cmd/sync-profiles -once
```

#### 定期実行（デーモンモード）

```bash
# デフォルト: 60分間隔
make sync-profiles-daemon
```

または:

```bash
# カスタム間隔（例: 30分間隔）
go run ./cmd/sync-profiles -interval 30
```

### サポートするプロフィールフォーマット

以下のフォーマットのメッセージを自動的にパースします:

```
⭕本名: じょぎ太郎
⭕学籍番号: 2XA1234
⭕趣味: ゲーム、アニメ鑑賞
⭕じょぎでやりたいこと: ゲーム作成
⭕ひとこと: よろしくお願いします！
```

記号の有無、全角・半角コロン、スペースなど、様々なフォーマットに対応しています。

### サーバーレスFunctionとしてデプロイ（推奨）

プロフィール同期をGoogle Cloud FunctionsやAWS Lambdaなどのサーバーレス環境にデプロイして、cronで定期実行できます。

詳細は [docs/deployment-functions.md](docs/deployment-functions.md) を参照してください。

#### クイックスタート（Google Cloud Functions）

```bash
cd deployments/cloud-functions

# 環境変数を設定
cp .env.yaml.example .env.yaml
# .env.yamlを編集

# デプロイ
./deploy.sh

# Cloud Schedulerをセットアップ（cronで毎時実行）
./setup-scheduler.sh
```

#### 対応プラットフォーム

- **Google Cloud Functions** - 推奨、無料枠が大きい
- **AWS Lambda** - EventBridgeでcron実行
- **Docker** - 任意のクラウドプロバイダーで実行可能

## プロジェクト構造

```
jyogi-discord-auth/
├── cmd/
│   ├── server/          # メインサーバーエントリーポイント
│   ├── sync-profiles/   # プロフィール同期ツール（CLI）
│   └── sync-profiles-fn/ # プロフィール同期Function（HTTP）
├── deployments/
│   ├── cloud-functions/ # Google Cloud Functions設定
│   └── aws-lambda/      # AWS Lambda設定
├── internal/
│   ├── domain/          # ドメインモデル（User, Profile, Session, etc.）
│   ├── repository/      # データアクセス層
│   ├── service/         # ビジネスロジック
│   ├── handler/         # HTTPハンドラー
│   ├── middleware/      # HTTPミドルウェア
│   └── config/          # 設定管理
├── pkg/
│   ├── discord/         # Discord APIクライアント、プロフィールパーサー
│   ├── auth/            # クライアント認証
│   └── jwt/             # JWTユーティリティ
├── web/
│   ├── templates/       # HTMLテンプレート
│   └── static/          # 静的ファイル
├── migrations/          # データベースマイグレーション
├── tests/               # テスト
└── scripts/             # 開発・運用スクリプト
```

## デプロイ

### Google Cloud Functions へのデプロイ（推奨）

**メインサーバー（認証API）**:

- Cloud Run等のコンテナサービスを推奨（詳細は今後追加予定）

**プロフィール同期のサーバーレスデプロイ**:

```bash
cd deployments/cloud-functions

# 環境変数を設定
cp .env.yaml.example .env.yaml
# .env.yamlを編集

# デプロイ
./deploy.sh

# Cloud Schedulerをセットアップ（cronで毎時実行）
./setup-scheduler.sh
```

詳細は [deployments/cloud-functions/README.md](deployments/cloud-functions/README.md) を参照してください。

### その他のデプロイ先

**Fly.io へのデプロイ**:

```bash
fly launch
fly deploy
```

**Railway へのデプロイ**:

```bash
railway init
railway up
```

詳細なデプロイ手順は [docs/deployment.md](docs/deployment.md) を参照してください。

## API エンドポイント

### 認証

- `GET /auth/login` - Discordログイン
- `GET /auth/callback` - Discordコールバック
- `POST /auth/logout` - ログアウト

### トークン

- `POST /token` - JWT発行
- `POST /token/refresh` - トークン更新

### OAuth2 (SSO)

- `GET /oauth/authorize` - 認可リクエスト
- `POST /oauth/token` - トークン取得

### API (Protected)

- `GET /api/verify` - JWT検証
- `GET /api/user` - ユーザー情報取得

詳細なAPI仕様は [specs/001-jyogi-member-auth/contracts/api.md](specs/001-jyogi-member-auth/contracts/api.md) を参照してください。

## ライセンス

MIT License

## 貢献

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。

## サポート

問題が発生した場合は、[GitHub Issues](https://github.com/jyogi-web/jyogi-discord-auth/issues)で報告してください。
