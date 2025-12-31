# Google Cloud Functions デプロイ

プロフィール同期機能をGoogle Cloud Functionsとしてデプロイし、Cloud Schedulerでcron実行します。

## 前提条件

1. Google Cloud Projectが作成済み
2. `gcloud` CLIがインストール済み
3. 必要なAPIが有効化されている:
   - Cloud Functions API
   - Cloud Scheduler API
   - Cloud Build API

## セットアップ

### 1. Google Cloud CLIの認証

```bash
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
```

### 2. 環境変数の設定

`.env.yaml`ファイルを作成:

```yaml
DISCORD_BOT_TOKEN: "your_bot_token_here"
DISCORD_PROFILE_CHANNEL: "your_channel_id_here"
DISCORD_GUILD_ID: "your_guild_id_here"
DATABASE_PATH: "/tmp/jyogi_auth.db"
JWT_SECRET: "your_jwt_secret_minimum_32_characters"
```

⚠️ **注意**: `.env.yaml`は`.gitignore`に追加してください

### 3. デプロイ

```bash
cd deployments/cloud-functions
./deploy.sh
```

または手動で:

```bash
gcloud functions deploy sync-profiles \
  --gen2 \
  --runtime go123 \
  --region asia-northeast1 \
  --source ../../ \
  --entry-point SyncProfilesHandler \
  --trigger-http \
  --no-allow-unauthenticated \
  --env-vars-file .env.yaml \
  --memory 256MB \
  --timeout 540s
```

### 4. Cloud Schedulerでcron設定

```bash
# サービスアカウントを作成（初回のみ）
gcloud iam service-accounts create cloud-scheduler-invoker \
  --display-name "Cloud Scheduler Invoker"

# 関数のURLを取得
FUNCTION_URL=$(gcloud functions describe sync-profiles \
  --gen2 \
  --region asia-northeast1 \
  --format='value(serviceConfig.uri)')

# Cloud Schedulerジョブを作成（毎時0分に実行）
gcloud scheduler jobs create http sync-profiles-job \
  --location asia-northeast1 \
  --schedule "0 * * * *" \
  --uri "$FUNCTION_URL/sync" \
  --http-method POST \
  --oidc-service-account-email cloud-scheduler-invoker@YOUR_PROJECT_ID.iam.gserviceaccount.com \
  --oidc-token-audience "$FUNCTION_URL"

# サービスアカウントに権限を付与
gcloud functions add-invoker-policy-binding sync-profiles \
  --gen2 \
  --region asia-northeast1 \
  --member "serviceAccount:cloud-scheduler-invoker@YOUR_PROJECT_ID.iam.gserviceaccount.com"
```

## cronスケジュール例

```bash
# 毎時0分
"0 * * * *"

# 毎日午前3時
"0 3 * * *"

# 毎週月曜日午前9時
"0 9 * * 1"

# 30分ごと
"*/30 * * * *"
```

## 手動実行

```bash
# 認証付きでリクエスト
FUNCTION_URL=$(gcloud functions describe sync-profiles \
  --gen2 \
  --region asia-northeast1 \
  --format='value(serviceConfig.uri)')

curl -X POST "$FUNCTION_URL/sync" \
  -H "Authorization: Bearer $(gcloud auth print-identity-token)"
```

## ログの確認

```bash
gcloud functions logs read sync-profiles \
  --gen2 \
  --region asia-northeast1 \
  --limit 50
```

## トラブルシューティング

### データベースエラー

Cloud Functionsは一時ストレージを使用するため、SQLiteは永続化されません。
本番環境では以下の選択肢があります:

1. **Cloud SQL (PostgreSQL/MySQL)**: 推奨
2. **Firestore**: NoSQLデータベース
3. **Cloud Storage + SQLite**: ステートレス関数

### タイムアウト

デフォルトは60秒です。処理時間が長い場合は`--timeout`で調整:

```bash
--timeout 540s  # 最大9分
```

### メモリ不足

デフォルトは256MBです。必要に応じて増やす:

```bash
--memory 512MB
```

## コスト見積もり

無料枠（月次）:
- 呼び出し: 200万回
- コンピューティング時間: 40万GB秒
- ネットワーク: 5GB

毎時1回実行（月720回）の場合、無料枠内で収まります。
