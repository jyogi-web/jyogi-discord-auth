# AWS Lambda デプロイ

プロフィール同期機能をAWS Lambdaとしてデプロイし、EventBridgeでcron実行します。

## 前提条件

1. AWSアカウントが作成済み
2. AWS CLIがインストール済み
3. IAMユーザーまたはロールに必要な権限がある

## セットアップ

### 1. AWS CLIの設定

```bash
aws configure
```

### 2. ビルド

Goバイナリをビルド:

```bash
cd ../../
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap cmd/sync-profiles-fn/main.go
zip function.zip bootstrap
```

### 3. Lambda関数の作成

```bash
# IAMロールを作成
aws iam create-role \
  --role-name jyogi-profile-sync-role \
  --assume-role-policy-document file://deployments/aws-lambda/trust-policy.json

# 基本実行権限を付与
aws iam attach-role-policy \
  --role-name jyogi-profile-sync-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

# Lambda関数を作成
aws lambda create-function \
  --function-name sync-profiles \
  --runtime provided.al2023 \
  --role arn:aws:iam::YOUR_ACCOUNT_ID:role/jyogi-profile-sync-role \
  --handler bootstrap \
  --zip-file fileb://function.zip \
  --timeout 540 \
  --memory-size 256 \
  --environment "Variables={
    DISCORD_BOT_TOKEN=your_token,
    DISCORD_PROFILE_CHANNEL=your_channel_id,
    DISCORD_GUILD_ID=your_guild_id,
    JWT_SECRET=your_secret,
    DATABASE_PATH=/tmp/jyogi_auth.db
  }"
```

### 4. EventBridge (CloudWatch Events) でcron設定

```bash
# Eventルールを作成（毎時0分）
aws events put-rule \
  --name sync-profiles-hourly \
  --schedule-expression "cron(0 * * * ? *)"

# Lambda関数に権限を付与
aws lambda add-permission \
  --function-name sync-profiles \
  --statement-id sync-profiles-hourly \
  --action lambda:InvokeFunction \
  --principal events.amazonaws.com \
  --source-arn arn:aws:events:REGION:ACCOUNT_ID:rule/sync-profiles-hourly

# Lambdaターゲットを追加
aws events put-targets \
  --rule sync-profiles-hourly \
  --targets "Id"="1","Arn"="arn:aws:lambda:REGION:ACCOUNT_ID:function:sync-profiles"
```

## cronスケジュール例（EventBridge形式）

```
# 毎時0分
cron(0 * * * ? *)

# 毎日午前3時（UTC）
cron(0 3 * * ? *)

# 毎週月曜日午前9時（UTC）
cron(0 9 ? * MON *)

# 30分ごと
cron(0/30 * * * ? *)
```

⚠️ **注意**: EventBridgeのcronはUTC時間です

## 更新

```bash
# 再ビルド
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap cmd/sync-profiles-fn/main.go
zip function.zip bootstrap

# 関数を更新
aws lambda update-function-code \
  --function-name sync-profiles \
  --zip-file fileb://function.zip
```

## 手動実行

```bash
aws lambda invoke \
  --function-name sync-profiles \
  --payload '{}' \
  response.json

cat response.json
```

## ログの確認

```bash
aws logs tail /aws/lambda/sync-profiles --follow
```

## コスト見積もり

無料枠（月次）:
- リクエスト: 100万回
- コンピューティング時間: 40万GB秒

毎時1回実行（月720回）の場合、無料枠内で収まります。

## データベース永続化

Lambda関数は一時ストレージを使用するため、SQLiteは永続化されません。
本番環境では以下の選択肢があります:

1. **Amazon RDS (PostgreSQL/MySQL)**: 推奨
2. **Amazon DynamoDB**: NoSQLデータベース
3. **S3 + SQLite**: ステートレス関数
