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
if [ -f .env ]; then
    echo -e "${BLUE}📄 .envファイルから環境変数を読み込んでいます...${NC}"
    export $(cat .env | grep -v '^#' | grep -v '^$' | xargs)
    echo -e "${GREEN}✓ .envファイルを読み込みました${NC}"
    echo ""
fi

# プロジェクト設定
PROJECT_ID="${GCP_PROJECT_ID:-your-gcp-project-id}"
REGION="${GCP_REGION:-asia-northeast1}"
SERVICE_NAME="jyogi-auth"

# 環境変数チェック
echo -e "${YELLOW}[1/7] 環境変数チェック${NC}"
if [ "$PROJECT_ID" = "your-gcp-project-id" ]; then
    echo -e "${RED}エラー: GCP_PROJECT_ID 環境変数を設定してください${NC}"
    echo "例: export GCP_PROJECT_ID=your-project-id"
    exit 1
fi

required_vars=(
    "DISCORD_CLIENT_ID"
    "DISCORD_CLIENT_SECRET"
    "DISCORD_REDIRECT_URI"
    "DISCORD_GUILD_ID"
    "JWT_SECRET"
)

missing_vars=()
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        missing_vars+=("$var")
    fi
done

if [ ${#missing_vars[@]} -gt 0 ]; then
    echo -e "${RED}エラー: 以下の環境変数が設定されていません:${NC}"
    for var in "${missing_vars[@]}"; do
        echo "  - $var"
    done
    echo ""
    echo "環境変数を設定してから再実行してください。"
    echo "例: export DISCORD_CLIENT_ID=your-client-id"
    exit 1
fi

echo -e "${GREEN}✓ 全ての環境変数が設定されています${NC}"
echo ""

# GCPプロジェクト設定
echo -e "${YELLOW}[2/7] GCPプロジェクト設定${NC}"
gcloud config set project "$PROJECT_ID"
echo -e "${GREEN}✓ プロジェクト: $PROJECT_ID${NC}"
echo ""

# Artifact Registry設定
echo -e "${YELLOW}[3/7] Artifact Registry設定${NC}"
ARTIFACT_REGISTRY_REPO="jyogi-auth"
ARTIFACT_REGISTRY_LOCATION="$REGION"

# 必要なAPIを有効化（請求アカウント不要）
echo "必要なAPIを有効化中..."
gcloud services enable artifactregistry.googleapis.com run.googleapis.com --quiet 2>/dev/null || true

# Artifact Registryリポジトリが存在するか確認
if ! gcloud artifacts repositories describe "$ARTIFACT_REGISTRY_REPO" --location="$ARTIFACT_REGISTRY_LOCATION" &>/dev/null; then
    echo "Artifact Registryリポジトリを作成中..."
    gcloud artifacts repositories create "$ARTIFACT_REGISTRY_REPO" \
      --repository-format=docker \
      --location="$ARTIFACT_REGISTRY_LOCATION" \
      --description="じょぎ認証システム"
    echo -e "${GREEN}✓ リポジトリ作成完了${NC}"
else
    echo -e "${GREEN}✓ リポジトリ確認完了${NC}"
fi

# Docker認証
gcloud auth configure-docker "$ARTIFACT_REGISTRY_LOCATION-docker.pkg.dev"
echo ""

# Dockerイメージビルド
echo -e "${YELLOW}[4/7] Dockerイメージビルド${NC}"
IMAGE_NAME="$ARTIFACT_REGISTRY_LOCATION-docker.pkg.dev/$PROJECT_ID/$ARTIFACT_REGISTRY_REPO/$SERVICE_NAME:latest"
# Cloud Runはlinux/amd64が必要（MacなどのARM環境からのデプロイ用）
docker build --platform linux/amd64 -t "$IMAGE_NAME" .
echo -e "${GREEN}✓ イメージビルド完了: $IMAGE_NAME${NC}"
echo ""

# Artifact Registryにプッシュ
echo -e "${YELLOW}[5/7] Artifact Registryにプッシュ${NC}"
docker push "$IMAGE_NAME"
echo -e "${GREEN}✓ プッシュ完了${NC}"
echo ""

# Cloud Runにデプロイ
echo -e "${YELLOW}[6/7] Cloud Runにデプロイ${NC}"

# CORS設定（デフォルト値）
CORS_ALLOWED_ORIGINS="${CORS_ALLOWED_ORIGINS:-http://localhost:3000,http://localhost:8080}"
# gcloudの引数パースエラーを回避するため、カンマをセミコロンに変換して渡す
# バックエンド側でセミコロンも区切り文字として認識するように修正済み
SAFE_CORS_ORIGINS=$(echo "$CORS_ALLOWED_ORIGINS" | tr ',' ';')

gcloud run deploy "$SERVICE_NAME" \
  --image "$IMAGE_NAME" \
  --region "$REGION" \
  --platform managed \
  --allow-unauthenticated \
  --min-instances 0 \
  --max-instances 3 \
  --memory 256Mi \
  --cpu 1 \
  --port 8080 \
  --set-env-vars "DISCORD_CLIENT_ID=$DISCORD_CLIENT_ID" \
  --set-env-vars "DISCORD_CLIENT_SECRET=$DISCORD_CLIENT_SECRET" \
  --set-env-vars "DISCORD_REDIRECT_URI=$DISCORD_REDIRECT_URI" \
  --set-env-vars "DISCORD_GUILD_ID=$DISCORD_GUILD_ID" \
  --set-env-vars "JWT_SECRET=$JWT_SECRET" \
  --set-env-vars "DATABASE_PATH=/app/data/auth.db" \
  --set-env-vars "SERVER_PORT=8080" \
  --set-env-vars "HTTPS_ONLY=true" \
  --set-env-vars "CORS_ALLOWED_ORIGINS=$SAFE_CORS_ORIGINS" \
  --set-env-vars "ENV=production"

echo -e "${GREEN}✓ デプロイ完了${NC}"
echo ""

# サービスURLを取得
SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" --region "$REGION" --format 'value(status.url)')
echo -e "${GREEN}=== デプロイ成功 ===${NC}"
echo -e "サービスURL: ${GREEN}$SERVICE_URL${NC}"

# Redirect URIを本番URLに更新
PROD_REDIRECT_URI="${SERVICE_URL}/auth/callback"
echo -e "${YELLOW}DISCORD_REDIRECT_URI を更新中: $PROD_REDIRECT_URI${NC}"

gcloud run services update "$SERVICE_NAME" \
  --region "$REGION" \
  --update-env-vars "DISCORD_REDIRECT_URI=$PROD_REDIRECT_URI" \
  --quiet

echo ""

# マイグレーション実行（初回のみ）
echo -e "${YELLOW}[7/7] マイグレーション実行確認${NC}"
echo "初回デプロイの場合、データベースマイグレーションを実行する必要があります。"
echo ""
read -p "マイグレーションを実行しますか？ (y/N): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}マイグレーションJob作成中...${NC}"

    # 既存のJobがあれば削除
    if gcloud run jobs describe "$SERVICE_NAME-migrate" --region "$REGION" &>/dev/null; then
        echo "既存のマイグレーションJobを削除中..."
        gcloud run jobs delete "$SERVICE_NAME-migrate" --region "$REGION" --quiet
    fi

    # マイグレーションJob作成
    gcloud run jobs create "$SERVICE_NAME-migrate" \
      --image "$IMAGE_NAME" \
      --region "$REGION" \
      --set-env-vars "DATABASE_PATH=/app/data/auth.db" \
      --command /bin/sh \
      --args "-c,cd /app && ./scripts/migrate.sh up"

    echo -e "${YELLOW}マイグレーション実行中...${NC}"
    gcloud run jobs execute "$SERVICE_NAME-migrate" --region "$REGION" --wait

    echo -e "${GREEN}✓ マイグレーション完了${NC}"
else
    echo "マイグレーションをスキップしました。"
    echo "後で実行する場合は以下のコマンドを実行してください:"
    echo ""
    echo "  gcloud run jobs create $SERVICE_NAME-migrate \\"
    echo "    --image $IMAGE_NAME \\"
    echo "    --region $REGION \\"
    echo "    --set-env-vars \"DATABASE_PATH=/app/data/auth.db\" \\"
    echo "    --command /bin/sh \\"
    echo "    --args \"-c,cd /app && ./scripts/migrate.sh up\""
    echo ""
    echo "  gcloud run jobs execute $SERVICE_NAME-migrate --region $REGION --wait"
fi

echo ""
echo -e "${GREEN}=== デプロイ完了 ===${NC}"
echo ""
echo "次のステップ:"
echo "1. ヘルスチェック: curl $SERVICE_URL/health"
echo "2. Discord OAuth2アプリ設定で Redirect URI を更新:"
echo "   $SERVICE_URL/auth/callback"
echo "3. 動作確認: $SERVICE_URL/auth/login にアクセス"
echo ""
