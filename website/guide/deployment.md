# デプロイメント

じょぎメンバー認証システムのデプロイ方法について説明します。

推奨される構成は以下の通りです：

- **認証サーバー**: Google Cloud Run
- **プロフィール同期**: Google Cloud Functions

## Google Cloud Run (認証サーバー)

### 概要

Cloud Runは、コンテナ化されたアプリケーションを実行するためのサーバーレスプラットフォームです。
認証サーバーはHTTPリクエストを処理するため、Cloud Runに適しています。

### セットアップ

#### 1. エントリーポイント

`cmd/server/main.go` が実行されるようにDockerfileを構成します（リポジトリのルートにある `Dockerfile` を使用）。

#### 2. デプロイスクリプト

`scripts/deploy-cloud-run.sh` を使用してデプロイできます。

```bash
export GCP_PROJECT_ID="your-project-id"
./scripts/deploy-cloud-run.sh
```

### 環境変数

Cloud Runサービスには以下の環境変数を設定してください：

- `DISCORD_CLIENT_ID`
- `DISCORD_CLIENT_SECRET`
- `DISCORD_REDIRECT_URI`
- `DISCORD_GUILD_ID`
- `JWT_SECRET`
- `CORS_ALLOWED_ORIGINS`

詳細な手順については、リポジトリ内の `docs/deployment-cloud-run.md` を参照してください。

---

## Google Cloud Functions (プロフィール同期)

### 概要

プロフィール同期は定期的に実行されるバッチ処理であるため、Cloud FunctionsとCloud Schedulerの組み合わせが最適です。

### デプロイ手順

`deployments/cloud-functions` ディレクトリにあるスクリプトを使用します。

```bash
cd deployments/cloud-functions
cp .env.yaml.example .env.yaml
./deploy.sh
```

### スケジューリング

`setup-scheduler.sh` を実行して、定期実行（cron）を設定します。

```bash
./setup-scheduler.sh
```

詳細については、リポジトリ内の `docs/deployment-functions.md` を参照してください。

## その他のデプロイオプション

- **Docker**: `docker-compose.yml` を使用して任意のサーバーで実行可能です。
- **AWS Lambda**: `deployments/aws-lambda` に設定例があります。
