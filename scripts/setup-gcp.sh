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

create_or_update_secret() {
    local SECRET_NAME=$1
    local SECRET_VALUE=$2

    if [ -z "$SECRET_VALUE" ]; then
        echo -e "${RED}警告: $SECRET_NAME の値が空です。スキップします。${NC}"
        return
    fi

    if gcloud secrets describe "$SECRET_NAME" --project "$PROJECT_ID" &>/dev/null; then
        echo -e "${YELLOW}既存のシークレット $SECRET_NAME を更新中...${NC}"
        echo -n "$SECRET_VALUE" | gcloud secrets versions add "$SECRET_NAME" --data-file=- --project "$PROJECT_ID" --quiet
        
        # 古いバージョンを無効化（最新版のみアクティブに）
    # 最新バージョンを明示的に取得（降順でソート）
   LATEST_VERSION=$(gcloud secrets versions list "$SECRET_NAME" --project "$PROJECT_ID" --format="value(name)" --filter="state=ENABLED" --limit=1 --sort-by="~name")
    # 最新版以外の有効なバージョンを取得
    VERSIONS=$(gcloud secrets versions list "$SECRET_NAME" --project "$PROJECT_ID" --format="value(name)" --filter="state=ENABLED AND name!=$LATEST_VERSION")
        for VERSION in $VERSIONS; do
            gcloud secrets versions disable "$VERSION" --secret="$SECRET_NAME" --project "$PROJECT_ID" --quiet
        done
    else
        echo -e "${YELLOW}新規シークレット $SECRET_NAME を作成中...${NC}"
        echo -n "$SECRET_VALUE" | gcloud secrets create "$SECRET_NAME" --data-file=- --project "$PROJECT_ID" --replication-policy="automatic" --quiet
    fi
}

# Discord設定をJSON化
DISCORD_CONFIG_JSON=$(jq -n \
  --arg client_id "$DISCORD_CLIENT_ID" \
  --arg client_secret "$DISCORD_CLIENT_SECRET" \
  --arg redirect_uri "$DISCORD_REDIRECT_URI" \
  --arg guild_id "$DISCORD_GUILD_ID" \
  --arg jwt_secret "$JWT_SECRET" \
  --arg bot_token "${DISCORD_BOT_TOKEN:-}" \
  '{client_id: $client_id, client_secret: $client_secret, redirect_uri: $redirect_uri, guild_id: $guild_id, jwt_secret: $jwt_secret, bot_token: $bot_token}')

# TiDB設定をJSON化
TIDB_CONFIG_JSON=$(jq -n \
  --arg host "$TIDB_DB_HOST" \
  --arg port "${TIDB_DB_PORT:-4000}" \
  --arg username "$TIDB_DB_USERNAME" \
  --arg password "$TIDB_DB_PASSWORD" \
  --arg database "$TIDB_DB_DATABASE" \
  '{host: $host, port: ($port|tonumber), username: $username, password: $password, database: $database}')

# 統合シークレットの作成
# NOTE
#  Secret Managerの管理数を抑えるため、関連設定を1つのシークレットにまとめています
#  アクティブバージョンは月6つ無料枠の範囲内に収まるように注意してください
create_or_update_secret "jyogi-discord-config" "$DISCORD_CONFIG_JSON"
create_or_update_secret "jyogi-tidb-config" "$TIDB_CONFIG_JSON"

echo -e "${GREEN}✓ シークレット設定完了${NC}"
echo ""

echo -e "${GREEN}=== セットアップ完了 ===${NC}"
echo "続いて ./scripts/deploy-cloud-run.sh を実行してデプロイしてください。"
echo ""
