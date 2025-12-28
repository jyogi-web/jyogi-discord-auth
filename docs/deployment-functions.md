# プロフィール同期Function デプロイガイド

プロフィール同期機能をサーバーレスFunctionとしてデプロイし、cronで定期実行する方法を説明します。

## 概要

このドキュメントでは、以下のクラウドプロバイダーへのデプロイ方法を説明します:

1. **Google Cloud Functions** (推奨)
2. **AWS Lambda**
3. **Docker (汎用)**

## デプロイ先の選択

| プロバイダー | 無料枠 | 設定の容易さ | データベース | 推奨度 |
|------------|-------|------------|------------|--------|
| **Google Cloud Functions** | 月200万回呼び出し | ⭐⭐⭐⭐⭐ | Cloud SQL | ⭐⭐⭐⭐⭐ |
| **AWS Lambda** | 月100万回呼び出し | ⭐⭐⭐⭐ | RDS | ⭐⭐⭐⭐ |
| **Docker** | - | ⭐⭐⭐ | 任意 | ⭐⭐⭐ |

## 1. Google Cloud Functions

### メリット
- 無料枠が大きい
- デプロイが簡単
- Cloud Schedulerとの統合が容易
- ログ管理が優れている

### デプロイ手順

詳細は [`deployments/cloud-functions/README.md`](../deployments/cloud-functions/README.md) を参照。

```bash
cd deployments/cloud-functions

# 環境変数を設定
cp .env.yaml.example .env.yaml
# .env.yamlを編集

# デプロイ
./deploy.sh

# Cloud Schedulerをセットアップ（cron設定）
./setup-scheduler.sh
```

### cron設定例

```bash
# 毎時0分
SCHEDULE="0 * * * *"

# 毎日午前3時（JST）
SCHEDULE="0 18 * * *"  # UTC 18:00 = JST 03:00

# 30分ごと
SCHEDULE="*/30 * * * *"
```

### コスト（月間）

毎時1回実行（720回/月）の場合:
- 呼び出し: 720回（無料枠内）
- コンピューティング: 約0.1GB秒（無料枠内）
- **月額: $0**

## 2. AWS Lambda

### メリット
- AWSエコシステムとの統合
- EventBridge (CloudWatch Events) で柔軟なcron設定
- 無料枠がある

### デプロイ手順

詳細は [`deployments/aws-lambda/README.md`](../deployments/aws-lambda/README.md) を参照。

```bash
# ビルド
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap cmd/sync-profiles-fn/main.go
zip function.zip bootstrap

# Lambda関数を作成
aws lambda create-function \
  --function-name sync-profiles \
  --runtime provided.al2023 \
  --role arn:aws:iam::YOUR_ACCOUNT_ID:role/jyogi-profile-sync-role \
  --handler bootstrap \
  --zip-file fileb://function.zip \
  --environment "Variables={DISCORD_BOT_TOKEN=...,DISCORD_PROFILE_CHANNEL=...}"

# EventBridgeでcron設定
aws events put-rule \
  --name sync-profiles-hourly \
  --schedule-expression "cron(0 * * * ? *)"
```

### コスト（月間）

毎時1回実行（720回/月）の場合:
- リクエスト: 720回（無料枠内）
- コンピューティング: 約0.1GB秒（無料枠内）
- **月額: $0**

## 3. Docker デプロイ

### メリット
- どのクラウドプロバイダーでも動作
- ローカルテストが容易
- 完全なコントロール

### ビルド

```bash
docker build -f Dockerfile.sync-profiles -t jyogi-profile-sync .
```

### ローカル実行

```bash
docker run -p 8080:8080 \
  -e DISCORD_BOT_TOKEN="your_token" \
  -e DISCORD_PROFILE_CHANNEL="your_channel_id" \
  -e DISCORD_GUILD_ID="your_guild_id" \
  -e JWT_SECRET="your_secret" \
  jyogi-profile-sync
```

### テスト

```bash
curl -X POST http://localhost:8080/sync
```

### Cloud Runへのデプロイ（Google Cloud）

```bash
# Container Registryにpush
docker tag jyogi-profile-sync gcr.io/YOUR_PROJECT_ID/jyogi-profile-sync
docker push gcr.io/YOUR_PROJECT_ID/jyogi-profile-sync

# Cloud Runにデプロイ
gcloud run deploy jyogi-profile-sync \
  --image gcr.io/YOUR_PROJECT_ID/jyogi-profile-sync \
  --platform managed \
  --region asia-northeast1 \
  --no-allow-unauthenticated \
  --set-env-vars DISCORD_BOT_TOKEN=...,DISCORD_PROFILE_CHANNEL=...

# Cloud Schedulerで定期実行
gcloud scheduler jobs create http sync-profiles-job \
  --schedule "0 * * * *" \
  --uri "$(gcloud run services describe jyogi-profile-sync --format='value(status.url)')/sync" \
  --http-method POST \
  --oidc-service-account-email your-service-account@your-project.iam.gserviceaccount.com
```

## データベース永続化

Functionは一時ストレージを使用するため、SQLiteはリクエスト間で永続化されません。

### 本番環境での選択肢

#### 1. PostgreSQL（推奨）

**Google Cloud SQL:**
```bash
# Cloud SQL PostgreSQLインスタンスを作成
gcloud sql instances create jyogi-auth-db \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=asia-northeast1

# データベースを作成
gcloud sql databases create jyogi_auth --instance=jyogi-auth-db

# Cloud Functionsから接続
gcloud functions deploy sync-profiles \
  --add-cloudsql-instances YOUR_PROJECT_ID:asia-northeast1:jyogi-auth-db
```

**AWS RDS:**
```bash
# RDS PostgreSQLインスタンスを作成
aws rds create-db-instance \
  --db-instance-identifier jyogi-auth-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --master-username admin \
  --master-user-password YourPassword \
  --allocated-storage 20
```

#### 2. NoSQLデータベース

- **Firestore (Google Cloud)**
- **DynamoDB (AWS)**

#### 3. S3/Cloud Storage + SQLite

ステートレス関数でSQLiteを使用:

1. 開始時にS3/Cloud StorageからSQLiteファイルをダウンロード
2. 処理実行
3. 終了時にS3/Cloud Storageにアップロード

⚠️ **注意**: 同時実行時の競合に注意

## cronスケジュール参考

### Google Cloud Functions / Cloud Run

```bash
# 毎時0分
"0 * * * *"

# 毎日午前3時（JST）
"0 18 * * *"  # UTC 18:00 = JST 03:00

# 毎週月曜日午前9時（JST）
"0 0 ? * MON"

# 30分ごと
"*/30 * * * *"

# 毎日正午（JST）
"0 3 * * *"  # UTC 03:00 = JST 12:00
```

### AWS EventBridge

```bash
# 毎時0分
"cron(0 * * * ? *)"

# 毎日午前3時（JST）
"cron(0 18 * * ? *)"  # UTC 18:00 = JST 03:00

# 毎週月曜日午前9時（JST）
"cron(0 0 ? * MON *)"

# 30分ごと
"cron(0/30 * * * ? *)"
```

⚠️ **タイムゾーンに注意**: クラウドプロバイダーのcronは通常UTC時間です

## モニタリングとログ

### Google Cloud

```bash
# ログを確認
gcloud functions logs read sync-profiles --gen2 --region asia-northeast1 --limit 50

# エラーログのみ
gcloud functions logs read sync-profiles --gen2 --region asia-northeast1 --filter="severity=ERROR"
```

### AWS

```bash
# ログを確認
aws logs tail /aws/lambda/sync-profiles --follow

# エラーログのみ
aws logs tail /aws/lambda/sync-profiles --filter-pattern "ERROR"
```

## トラブルシューティング

### タイムアウト

Functionのタイムアウトを延長:

**Google Cloud:**
```bash
--timeout 540s  # 最大9分
```

**AWS:**
```bash
--timeout 900  # 最大15分
```

### メモリ不足

メモリを増やす:

**Google Cloud:**
```bash
--memory 512MB
```

**AWS:**
```bash
--memory-size 512
```

### 権限エラー

- Botがサーバーに招待されているか確認
- MESSAGE CONTENT INTENTが有効か確認
- チャンネル権限が付与されているか確認

## セキュリティベストプラクティス

1. **環境変数で秘密情報を管理**
   - Bot Tokenは環境変数で設定
   - `.env.yaml`をGitに含めない

2. **認証を有効化**
   - Cloud Functions: `--no-allow-unauthenticated`
   - AWS Lambda: IAM認証

3. **最小権限の原則**
   - 必要な権限のみを付与

4. **ログに秘密情報を含めない**
   - Tokenなどをログ出力しない

## まとめ

推奨構成:

1. **小規模・テスト**: Google Cloud Functions（無料枠が大きい）
2. **中規模**: Cloud Run + Cloud SQL（柔軟性とスケーラビリティ）
3. **AWSユーザー**: AWS Lambda + RDS

すべての構成で、月間720回（毎時1回）の実行は無料枠内で収まります。
