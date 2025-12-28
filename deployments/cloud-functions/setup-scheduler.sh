#!/bin/bash

# Cloud Scheduler セットアップスクリプト

set -e

# 設定
PROJECT_ID=${GCLOUD_PROJECT:-$(gcloud config get-value project)}
REGION="asia-northeast1"
FUNCTION_NAME="sync-profiles"
JOB_NAME="sync-profiles-job"
SCHEDULE="0 * * * *"  # 毎時0分
SERVICE_ACCOUNT_NAME="cloud-scheduler-invoker"
SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

echo "⏰ Setting up Cloud Scheduler for profile sync..."
echo ""
echo "Project: $PROJECT_ID"
echo "Region: $REGION"
echo "Schedule: $SCHEDULE"
echo ""

# サービスアカウントを作成（存在しない場合）
echo "📋 Creating service account..."
if gcloud iam service-accounts describe $SERVICE_ACCOUNT_EMAIL --project $PROJECT_ID &>/dev/null; then
    echo "Service account already exists: $SERVICE_ACCOUNT_EMAIL"
else
    gcloud iam service-accounts create $SERVICE_ACCOUNT_NAME \
      --display-name "Cloud Scheduler Invoker" \
      --project $PROJECT_ID
    echo "✅ Service account created: $SERVICE_ACCOUNT_EMAIL"
fi

# 関数のURLを取得
echo ""
echo "🔍 Getting function URL..."
FUNCTION_URL=$(gcloud functions describe $FUNCTION_NAME \
  --gen2 \
  --region $REGION \
  --format='value(serviceConfig.uri)' \
  --project $PROJECT_ID)

if [ -z "$FUNCTION_URL" ]; then
    echo "❌ Error: Function $FUNCTION_NAME not found"
    echo "Please deploy the function first using ./deploy.sh"
    exit 1
fi

echo "Function URL: $FUNCTION_URL"

# サービスアカウントに関数の呼び出し権限を付与
echo ""
echo "🔐 Granting permissions..."
gcloud functions add-invoker-policy-binding $FUNCTION_NAME \
  --gen2 \
  --region $REGION \
  --member "serviceAccount:$SERVICE_ACCOUNT_EMAIL" \
  --project $PROJECT_ID

# Cloud Schedulerジョブを作成または更新
echo ""
echo "⏰ Creating scheduler job..."

if gcloud scheduler jobs describe $JOB_NAME --location $REGION --project $PROJECT_ID &>/dev/null; then
    echo "Job already exists. Updating..."
    gcloud scheduler jobs update http $JOB_NAME \
      --location $REGION \
      --schedule "$SCHEDULE" \
      --uri "${FUNCTION_URL}/sync" \
      --http-method POST \
      --oidc-service-account-email $SERVICE_ACCOUNT_EMAIL \
      --oidc-token-audience "$FUNCTION_URL" \
      --project $PROJECT_ID
else
    gcloud scheduler jobs create http $JOB_NAME \
      --location $REGION \
      --schedule "$SCHEDULE" \
      --uri "${FUNCTION_URL}/sync" \
      --http-method POST \
      --oidc-service-account-email $SERVICE_ACCOUNT_EMAIL \
      --oidc-token-audience "$FUNCTION_URL" \
      --project $PROJECT_ID
fi

echo ""
echo "✅ Cloud Scheduler setup complete!"
echo ""
echo "Job details:"
echo "  Name: $JOB_NAME"
echo "  Schedule: $SCHEDULE (cron format)"
echo "  URL: ${FUNCTION_URL}/sync"
echo ""
echo "Commands:"
echo "  Test job manually:"
echo "    gcloud scheduler jobs run $JOB_NAME --location $REGION"
echo ""
echo "  View job details:"
echo "    gcloud scheduler jobs describe $JOB_NAME --location $REGION"
echo ""
echo "  View execution logs:"
echo "    gcloud functions logs read $FUNCTION_NAME --gen2 --region $REGION --limit 50"
