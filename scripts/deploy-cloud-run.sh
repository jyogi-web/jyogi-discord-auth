#!/bin/bash
# Google Cloud Run デプロイスクリプト
# じょぎメンバー認証システムをCloud Runにデプロイします

set -e

# 色付き出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== じょぎメンバー認証システム Cloud Run デプロイ ===${NC}"
echo ""

# .envファイルから環境変数を読み込む
# 環境変数の読み込み (.env.deploy を優先)
if [ -f .env.deploy ]; then
    echo -e "${BLUE}📄 .env.deploy ファイルから環境変数を読み込んでいます...${NC}"
    export $(cat .env.deploy | grep -v '^#' | grep -v '^$' | xargs)
    echo -e "${GREEN}✓ .env.deploy ファイルを読み込みました${NC}"
elif [ -f .env ]; then
    echo -e "${BLUE}📄 .env ファイルから環境変数を読み込んでいます...${NC}"
    export $(cat .env | grep -v '^#' | grep -v '^$' | xargs)
    echo -e "${GREEN}✓ .env ファイルを読み込みました${NC}"
fi
echo ""

# プロジェクト設定
PROJECT_ID="${GCP_PROJECT_ID:-your-gcp-project-id}"
REGION="${GCP_REGION:-asia-northeast1}"
SERVICE_NAME="jyogi-auth"
ARTIFACT_REGISTRY_REPO="jyogi-auth"
ARTIFACT_REGISTRY_LOCATION="$REGION"

# 環境変数チェック
echo -e "${YELLOW}環境変数チェック${NC}"
if [ "$PROJECT_ID" = "your-gcp-project-id" ]; then
    echo -e "${RED}エラー: GCP_PROJECT_ID 環境変数を設定してください${NC}"
    echo "例: export GCP_PROJECT_ID=your-project-id"
    exit 1
fi

echo -e "${GREEN}✓ プロジェクトID: $PROJECT_ID${NC}"
echo ""

# GCPプロジェクト設定
gcloud config set project "$PROJECT_ID" --quiet

# Docker認証
echo -e "${YELLOW}Docker認証${NC}"
gcloud auth configure-docker "$ARTIFACT_REGISTRY_LOCATION-docker.pkg.dev" --quiet
echo ""

# Dockerイメージビルド
echo -e "${YELLOW}Dockerイメージビルド${NC}"
IMAGE_NAME="$ARTIFACT_REGISTRY_LOCATION-docker.pkg.dev/$PROJECT_ID/$ARTIFACT_REGISTRY_REPO/$SERVICE_NAME:latest"
# Cloud Runはlinux/amd64が必要（MacなどのARM環境からのデプロイ用）
docker build --platform linux/amd64 -t "$IMAGE_NAME" .
echo -e "${GREEN}✓ イメージビルド完了: $IMAGE_NAME${NC}"
echo ""

# Artifact Registryにプッシュ
echo -e "${YELLOW}Artifact Registryにプッシュ${NC}"
docker push "$IMAGE_NAME"
echo -e "${GREEN}✓ プッシュ完了${NC}"
echo ""

# Cloud Runにデプロイ
echo -e "${YELLOW}Cloud Runにデプロイ${NC}"

# CORS設定（デフォルト値）
CORS_ALLOWED_ORIGINS="${CORS_ALLOWED_ORIGINS:-http://localhost:3000,http://localhost:8080}"
# gcloudの引数パースエラーを回避するため、カンマをセミコロンに変換して渡す
SAFE_CORS_ORIGINS=$(echo "$CORS_ALLOWED_ORIGINS" | tr ',' ';')

gcloud run deploy "$SERVICE_NAME" \
  --image "$IMAGE_NAME" \
  --region "$REGION" \
  --platform managed \
  --allow-unauthenticated \
  --min-instances 0 \
  --max-instances 1 \
  --memory 256Mi \
  --cpu 1 \
  --port 8080 \
  --set-secrets "DISCORD_CONFIG=jyogi-discord-config:latest" \
  --set-secrets "TIDB_CONFIG=jyogi-tidb-config:latest" \
  --set-env-vars "SERVER_PORT=8080" \
  --set-env-vars "HTTPS_ONLY=true" \
  --set-env-vars "CORS_ALLOWED_ORIGINS=$SAFE_CORS_ORIGINS" \
  --set-env-vars "ENV=production"

echo -e "${GREEN}✓ デプロイ完了${NC}"
echo ""

# サービスURLを取得
SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" --region "$REGION" --format 'value(status.url)')

echo -e "サービスURL: ${GREEN}$SERVICE_URL${NC}"

# Redirect URIの案内
PROD_REDIRECT_URI="${SERVICE_URL}/auth/callback"
echo -e "Redirect URI: ${BLUE}$PROD_REDIRECT_URI${NC}"
echo ""
echo -e "${YELLOW}注意: Redirect URIに変更がある場合は、.envファイルの DISCORD_REDIRECT_URI を更新し、${NC}"
echo -e "${YELLOW}      ./scripts/setup-gcp.sh を実行してシークレットを更新して再デプロイしてください。${NC}"

echo ""
