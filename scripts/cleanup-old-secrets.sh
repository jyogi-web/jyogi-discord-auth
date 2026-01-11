#!/bin/bash
# 古いシークレットを削除するスクリプト

set -e

# .envファイルから環境変数を読み込む
if [ -f .env.deploy ]; then
    export $(cat .env.deploy | grep -v '^#' | grep -v '^$' | xargs)
elif [ -f .env ]; then
    export $(cat .env | grep -v '^#' | grep -v '^$' | xargs)
fi

PROJECT_ID="${GCP_PROJECT_ID:-your-gcp-project-id}"

if [ "$PROJECT_ID" = "your-gcp-project-id" ]; then
    echo "エラー: GCP_PROJECT_ID 環境変数を設定してください"
    exit 1
fi

# 削除対象の古いシークレットリスト
OLD_SECRETS=(
    "jyogi-discord-client-id"
    "jyogi-discord-client-secret"
    "jyogi-discord-redirect-uri"
    "jyogi-discord-guild-id"
    "jyogi-jwt-secret"
    "jyogi-tidb-host"
    "jyogi-tidb-port"
    "jyogi-tidb-username"
    "jyogi-tidb-password"
    "jyogi-tidb-database"
    "jyogi-discord-bot-token"
)

echo "プロジェクト: $PROJECT_ID"
echo "以下のシークレットを削除します（新しいJSON形式シークレットへの移行後のみ実行してください）:"
for SECRET in "${OLD_SECRETS[@]}"; do
    echo "- $SECRET"
done

echo ""
echo "注意: 一度削除すると復元できません。デプロイが成功し、動作確認が完了していることを確認してください。"
read -p "本当に削除しますか？ (y/N): " CONFIRM
if [ "$CONFIRM" != "y" ]; then
    echo "キャンセルしました"
    exit 0
fi

for SECRET in "${OLD_SECRETS[@]}"; do
    echo "削除中: $SECRET"
    gcloud secrets delete "$SECRET" --project "$PROJECT_ID" --quiet || echo "スキップ (存在しないか削除失敗): $SECRET"
done

echo "完了しました"
