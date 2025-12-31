#!/bin/bash

# Google Cloud Functions デプロイスクリプト

set -e

# 設定
PROJECT_ID=${GCLOUD_PROJECT:-$(gcloud config get-value project)}
REGION="asia-northeast1"
FUNCTION_NAME="sync-profiles"
RUNTIME="go123"
MEMORY="256MB"
TIMEOUT="540s"

echo "🚀 Deploying profile sync function to Google Cloud Functions..."
echo ""
echo "Project: $PROJECT_ID"
echo "Region: $REGION"
echo "Function: $FUNCTION_NAME"
echo ""

# .env.yamlファイルの存在確認
if [ ! -f .env.yaml ]; then
    echo "❌ Error: .env.yaml file not found"
    echo "Please create .env.yaml file with your environment variables"
    echo "See README.md for details"
    exit 1
fi

# デプロイ
echo "📦 Deploying function..."
gcloud functions deploy $FUNCTION_NAME \
  --gen2 \
  --runtime $RUNTIME \
  --region $REGION \
  --source ../../ \
  --entry-point SyncProfilesHandler \
  --trigger-http \
  --no-allow-unauthenticated \
  --env-vars-file .env.yaml \
  --memory $MEMORY \
  --timeout $TIMEOUT \
  --project $PROJECT_ID

echo ""
echo "✅ Function deployed successfully!"
echo ""

# 関数のURLを取得
FUNCTION_URL=$(gcloud functions describe $FUNCTION_NAME \
  --gen2 \
  --region $REGION \
  --format='value(serviceConfig.uri)')

echo "Function URL: $FUNCTION_URL"
echo ""
echo "Next steps:"
echo "  1. Set up Cloud Scheduler (see README.md)"
echo "  2. Test the function:"
echo "     curl -X POST \"$FUNCTION_URL/sync\" \\"
echo "       -H \"Authorization: Bearer \$(gcloud auth print-identity-token)\""
