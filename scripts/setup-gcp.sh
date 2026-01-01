#!/bin/bash
# Google Cloud Platform セットアップスクリプト
# 初回構築時や設定変更時に実行します

set -e

# 色付き出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== じょぎメンバー認証システム GCPセットアップ ===${NC}"
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
echo -e "${YELLOW}GCPプロジェクト設定${NC}"
gcloud config set project "$PROJECT_ID"
echo -e "${GREEN}✓ プロジェクト設定完了${NC}"
echo ""

# 必要なAPIを有効化（請求アカウント不要）
echo -e "${YELLOW}API有効化${NC}"
echo "必要なAPIを有効化中..."
gcloud services enable artifactregistry.googleapis.com run.googleapis.com secretmanager.googleapis.com --quiet 2>/dev/null || true
echo -e "${GREEN}✓ API有効化完了${NC}"
echo ""

# Artifact Registry設定
echo -e "${YELLOW}Artifact Registry設定${NC}"
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
echo ""

# Cloud RunサービスアカウントにSecret Managerのアクセス権限を付与
echo -e "${YELLOW}権限設定${NC}"
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')
SERVICE_ACCOUNT="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"
echo "サービスアカウント ($SERVICE_ACCOUNT) にシークレットアクセス権限を付与中..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/secretmanager.secretAccessor" \
    --quiet >/dev/null
echo -e "${GREEN}✓ 権限付与完了${NC}"
echo ""

# シークレットの作成・更新
echo -e "${YELLOW}シークレット設定${NC}"

create_secret() {
    local name="$1"
    local value="$2"
    
    if [ -z "$value" ]; then
        echo -e "${RED}警告: $name の値が空です。スキップします。${NC}"
        return
    fi
    
    # シークレットが存在しない場合は作成
    if ! gcloud secrets describe "$name" --project "$PROJECT_ID" &>/dev/null; then
        echo "シークレット $name を作成中..."
        gcloud secrets create "$name" --replication-policy="automatic" --project "$PROJECT_ID" --quiet
    fi
    
    # 新しいバージョンを追加
    echo -n "$value" | gcloud secrets versions add "$name" --data-file=- --project "$PROJECT_ID" --quiet >/dev/null
    echo "✓ $name を設定しました"
}

create_secret "jyogi-discord-client-id" "$DISCORD_CLIENT_ID"
create_secret "jyogi-discord-client-secret" "$DISCORD_CLIENT_SECRET"
create_secret "jyogi-discord-redirect-uri" "$DISCORD_REDIRECT_URI"
create_secret "jyogi-discord-guild-id" "$DISCORD_GUILD_ID"
create_secret "jyogi-jwt-secret" "$JWT_SECRET"

if [ -n "$DISCORD_BOT_TOKEN" ]; then
    create_secret "jyogi-discord-bot-token" "$DISCORD_BOT_TOKEN"
fi

echo -e "${GREEN}✓ シークレット設定完了${NC}"
echo ""

echo -e "${GREEN}=== セットアップ完了 ===${NC}"
echo "続いて ./scripts/deploy-cloud-run.sh を実行してデプロイしてください。"
echo ""
