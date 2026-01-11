#!/bin/bash
# じょぎメンバー認証システム 全デプロイ実行スクリプト
# インフラ設定（シークレット含む）の更新とアプリケーションのデプロイを連続して行います。

set -e

# 色付き出力
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== じょぎメンバー認証システム 全デプロイ開始 ===${NC}"
echo ""

# 1. GCPセットアップ（API有効化、リポジトリ作成、シークレット更新）
./scripts/setup-gcp.sh

echo ""
echo -e "${GREEN}=== GCPセットアップ完了 ===${NC}"
echo ""

# 2. アプリケーションデプロイ（ビルド、プッシュ、Cloud Runデプロイ）
./scripts/deploy-cloud-run.sh

echo ""
echo -e "${GREEN}=== じょぎメンバー認証システム 全デプロイ完了 ===${NC}"
