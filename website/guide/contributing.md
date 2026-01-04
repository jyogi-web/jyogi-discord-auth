# クイックスタート

じょぎメンバー認証システムの開発環境セットアップ手順です。

## 前提条件

1. Go 1.25以上がインストールされていること
2. Discord Developer Portalでアプリケーションを作成済み
3. じょぎDiscordサーバーのサーバーIDを取得済み

## 1. リポジトリのクローン

```bash
git clone https://github.com/jyogi-web/jyogi-discord-auth.git
cd jyogi-discord-auth
```

## 2. 依存関係のインストール

```bash
go mod download
```

## 3. 環境変数の設定

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

## 4. サーバーの起動

```bash
go run cmd/server/main.go
```

サーバーが起動したら、ブラウザで `http://localhost:8080` にアクセスしてください。

## Dockerでの開発

Dockerを使用することで、環境構築を簡単にし、チーム全体で統一された開発環境を利用できます。

### 前提条件

- Docker & Docker Compose がインストールされていること

### 起動

```bash
# 環境変数の設定
cp .env.example .env
# .envファイルを編集して必要な値を設定

# Docker Composeでビルド＆起動
docker-compose up -d

# ログを確認
docker-compose logs -f
```

### Makefile

makeコマンドで頻繁に行うタスクを実行できます。

```bash
make help        # ヘルプを表示
make run         # ローカル起動
make docker-up   # Docker起動
```
