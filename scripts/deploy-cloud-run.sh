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

# Bot Tokenシークレットフラグの構築
if gcloud secrets describe "jyogi-discord-bot-token" --project "$PROJECT_ID" &>/dev/null; then
    BOT_TOKEN_SECRET_FLAG="--set-secrets=DISCORD_BOT_TOKEN=jyogi-discord-bot-token:latest"
else
    BOT_TOKEN_SECRET_FLAG=""
fi

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
  --set-secrets "DISCORD_CLIENT_ID=jyogi-discord-client-id:latest" \
  --set-secrets "DISCORD_CLIENT_SECRET=jyogi-discord-client-secret:latest" \
  --set-secrets "DISCORD_REDIRECT_URI=jyogi-discord-redirect-uri:latest" \
  --set-secrets "DISCORD_GUILD_ID=jyogi-discord-guild-id:latest" \
  --set-secrets "JWT_SECRET=jyogi-jwt-secret:latest" \
  --set-secrets "TIDB_DB_HOST=jyogi-tidb-host:latest" \
  --set-secrets "TIDB_DB_PORT=jyogi-tidb-port:latest" \
  --set-secrets "TIDB_DB_USERNAME=jyogi-tidb-username:latest" \
  --set-secrets "TIDB_DB_PASSWORD=jyogi-tidb-password:latest" \
  --set-secrets "TIDB_DB_DATABASE=jyogi-tidb-database:latest" \
  $BOT_TOKEN_SECRET_FLAG \
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
echo -e "${YELLOW}DISCORD_REDIRECT_URI をチェック中: $PROD_REDIRECT_URI${NC}"

# 現在のシークレット値を取得
CURRENT_REDIRECT_URI=$(gcloud secrets versions access latest --secret="jyogi-discord-redirect-uri" --project "$PROJECT_ID" --quiet 2>/dev/null || echo "")

if [ "$PROD_REDIRECT_URI" != "$CURRENT_REDIRECT_URI" ]; then
    echo -e "${YELLOW}Redirect URIが変更されました。更新を実行します。${NC}"
    echo "Current: $CURRENT_REDIRECT_URI"
    echo "New:     $PROD_REDIRECT_URI"

    # Secret ManagerのRedirect URIを更新
    echo -n "$PROD_REDIRECT_URI" | gcloud secrets versions add "jyogi-discord-redirect-uri" --data-file=- --project "$PROJECT_ID" --quiet >/dev/null

    # 新しいシークレット値を反映させるために新しいリビジョンを作成
    echo "新しい設定を反映させるためにサービスを更新中..."
    gcloud run services update "$SERVICE_NAME" \
      --region "$REGION" \
      --force-new-revision \
      --quiet
    
    echo -e "${GREEN}✓ Redirect URIを更新しました${NC}"
else
    echo -e "${GREEN}✓ Redirect URIは最新です。更新をスキップします。${NC}"
fi

echo ""

# マイグレーションはAutoMigrateにより起動時に実行されるため、別途Jobは不要
echo -e "${YELLOW}マイグレーションはアプリケーション起動時に自動実行されます${NC}"

echo ""
echo -e "${GREEN}=== デプロイ完了 ===${NC}"
echo ""
echo "次のステップ:"
echo "1. ヘルスチェック: curl $SERVICE_URL/health"
echo "2. Discord OAuth2アプリ設定で Redirect URI を更新（必要があれば）:"
echo "   $SERVICE_URL/auth/callback"
echo "3. 動作確認: $SERVICE_URL/auth/login にアクセス"
echo ""
