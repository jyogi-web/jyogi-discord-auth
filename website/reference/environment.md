# 環境変数リファレンス

システムの動作を設定するための環境変数一覧です。
開発環境では `.env` ファイルに、本番環境（Cloud Runなど）ではサービスの環境変数として設定します。

## 必須設定

| 変数名 | 説明 | 例 |
| :--- | :--- | :--- |
| `DISCORD_CLIENT_ID` | Discord Developer Portalで取得したClient ID | `123456789012345678` |
| `DISCORD_CLIENT_SECRET` | Discord Developer Portalで取得したClient Secret | `abcdefghijklmnopqrstuvwxyz` |
| `DISCORD_REDIRECT_URI` | OAuth2コールバックURL | `http://localhost:8080/auth/callback` |
| `DISCORD_GUILD_ID` | 対象のDiscordサーバーID | `987654321098765432` |
| `JWT_SECRET` | JWT署名用シークレット（32文字以上推奨） | `your-secure-random-string-minimum-32-chars` |

## プロフィール同期設定

プロフィール同期機能を使用する場合に必要です。

| 変数名 | 説明 | 例 |
| :--- | :--- | :--- |
| `DISCORD_BOT_TOKEN` | Discord Bot Token | `MTA...` |
| `DISCORD_PROFILE_CHANNEL` | 自己紹介チャンネルのID | `123456789012345678` |

## サーバー・DB設定

| 変数名 | 説明 | デフォルト値 |
| :--- | :--- | :--- |
| `SERVER_PORT` | サーバーがリッスンするポート | `8080` |
| `ENV` | 実行環境 (`development` / `production`) | `development` |
| `DATABASE_PATH` | SQLiteデータベースファイルのパス（開発用） | `./jyogi_auth.db` |
| `HTTPS_ONLY` | HTTPSを強制するか (`true` / `false`) | `false` |
| `CORS_ALLOWED_ORIGINS` | CORSを許可するオリジン（カンマ区切り） | `http://localhost:3000` |

## Cloud Run / TiDB設定 (本番用)

| 変数名 | 説明 |
| :--- | :--- |
| `GCP_PROJECT_ID` | Google Cloud プロジェクトID |
| `GCP_REGION` | デプロイリージョン (例: `asia-northeast1`) |
| `TIDB_DB_HOST` | TiDBホスト名 |
| `TIDB_DB_PORT` | TiDBポート番号 (デフォルト: `4000`) |
| `TIDB_DB_USERNAME` | TiDBユーザー名 |
| `TIDB_DB_PASSWORD` | TiDBパスワード |
| `TIDB_DB_DATABASE` | データベース名 |
| `TIDB_DISABLE_TLS` | TLS接続を無効にするか (`true` / `false`) |

## セキュリティのベストプラクティス

### シークレット管理

- **開発環境**: `.env` ファイルを使用し、必ず `.gitignore` に追加してください。
- **本番環境**: 環境変数として直接設定するか、GCP Secret Manager などのシークレット管理サービスを使用してください。
- **シークレットの生成**: `JWT_SECRET` などは以下のコマンドで生成できます：
  ```bash
  openssl rand -base64 32
  ```

### 本番環境の必須設定

本番環境（`ENV=production`）では以下の設定を必ず有効にしてください：

- `HTTPS_ONLY=true`: HTTPS 接続を強制
- `CORS_ALLOWED_ORIGINS`: 本番ドメインのみを許可
