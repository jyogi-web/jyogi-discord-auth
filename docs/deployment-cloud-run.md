# Google Cloud Run デプロイガイド

じょぎメンバー認証システムをGoogle Cloud Runにデプロイする手順を説明します。

## 概要

このガイドでは、以下の機能をCloud Runにデプロイします：

- ✅ Discord OAuth2認証
- ✅ JWT発行・検証
- ✅ セッション管理
- ✅ OAuth2 SSO（クライアントアプリ統合）

**注意**: プロフィール同期機能は含まれません（Phase 2で実装予定）

## 前提条件

### 1. Google Cloud Platform

- [ ] GCPアカウントを作成済み
- [ ] プロジェクトを作成済み（または既存のプロジェクトを使用）
- [ ] gcloud CLIをインストール済み

### 2. ローカル環境

- [ ] Docker Desktop インストール済み
- [ ] Go 1.23以上 インストール済み
- [ ] git インストール済み

### 3. Discord OAuth2アプリ

- [ ] Discord Developer Portalでアプリケーションを作成済み
- [ ] Client IDとClient Secretを取得済み
- [ ] Guild ID（じょぎサーバーID）を取得済み

## セットアップ手順

### Step 1: GCP CLIのセットアップ

#### 1.1 gcloud CLIのインストール

```bash
# macOS (Homebrew)
brew install --cask google-cloud-sdk

# または公式サイトからダウンロード
# https://cloud.google.com/sdk/docs/install
```

#### 1.2 gcloud CLIの初期化

```bash
# 認証
gcloud auth login

# プロジェクトを作成（または既存のプロジェクトを選択）
gcloud projects create YOUR_PROJECT_ID --name="じょぎ認証システム"

# プロジェクトを設定
gcloud config set project YOUR_PROJECT_ID

# 必要なAPIを有効化
gcloud services enable run.googleapis.com
gcloud services enable containerregistry.googleapis.com
```

### Step 2: 環境変数の設定

デプロイに必要な環境変数を設定します。

```bash
# GCPプロジェクトID
export GCP_PROJECT_ID="YOUR_PROJECT_ID"
export GCP_REGION="asia-northeast1"

# Discord OAuth2設定
export DISCORD_CLIENT_ID="your-discord-client-id"
export DISCORD_CLIENT_SECRET="your-discord-client-secret"
export DISCORD_REDIRECT_URI="https://YOUR_SERVICE_URL/auth/callback"  # 後で更新
export DISCORD_GUILD_ID="your-guild-id"

# JWT秘密鍵（強力なランダム文字列を生成）
export JWT_SECRET=$(openssl rand -base64 32)

# CORS設定（カンマ区切りで複数指定可能）
export CORS_ALLOWED_ORIGINS="https://your-frontend-app.com,https://another-app.com"
```

**重要**: これらの環境変数は、デプロイスクリプト実行時に必要です。永続化したい場合は `~/.bashrc` または `~/.zshrc` に追加してください。

#### JWT秘密鍵の永続化（推奨）

```bash
# JWT秘密鍵を生成して保存
echo "export JWT_SECRET=$(openssl rand -base64 32)" >> ~/.bashrc
source ~/.bashrc

# 確認
echo $JWT_SECRET
```

### Step 3: デプロイ実行

#### 3.1 リポジトリのクローン（未実施の場合）

```bash
git clone https://github.com/jyogi-web/jyogi-discord-auth.git
cd jyogi-discord-auth
```

#### 3.2 デプロイスクリプトの実行

```bash
# デプロイスクリプトを実行
./scripts/deploy-cloud-run.sh
```

スクリプトは以下の処理を自動実行します：

1. 環境変数チェック
2. GCPプロジェクト設定
3. Dockerイメージビルド
4. Container Registryにプッシュ
5. Cloud Runにデプロイ
6. マイグレーション実行（初回のみ）

#### 3.3 サービスURLの確認

デプロイ完了後、以下のようなURLが表示されます：

```
サービスURL: https://jyogi-auth-XXXXXXX-an.a.run.app
```

このURLをメモしてください。

### Step 4: Discord OAuth2アプリの設定更新

1. [Discord Developer Portal](https://discord.com/developers/applications) にアクセス
2. アプリケーションを選択
3. "OAuth2" → "General" に移動
4. "Redirects" セクションで以下のURLを追加：

```
https://jyogi-auth-XXXXXXX-an.a.run.app/auth/callback
```

5. "Save Changes" をクリック

### Step 5: 動作確認

#### 5.1 ヘルスチェック

```bash
curl https://YOUR_SERVICE_URL/health
# 期待結果: OK
```

#### 5.2 認証フロー確認

1. ブラウザで以下のURLにアクセス：

```
https://YOUR_SERVICE_URL/auth/login
```

2. Discordログイン画面が表示されることを確認
3. ログインしてコールバックが成功することを確認
4. セッションCookieが設定されることを確認

#### 5.3 トークン発行確認

```bash
# セッションCookieを使ってトークン発行
curl -X POST https://YOUR_SERVICE_URL/token \
  -H "Cookie: session=YOUR_SESSION_COOKIE"

# 期待結果: {"token": "eyJhbGci..."}
```

## トラブルシューティング

### デプロイエラー

#### エラー: "permission denied"

```bash
# Docker Desktopが起動していることを確認
open -a Docker

# gcloud認証を再実行
gcloud auth login
gcloud auth configure-docker
```

#### エラー: "quota exceeded"

無料枠を超過している可能性があります。GCP Consoleで使用量を確認してください。

### 認証エラー

#### エラー: "invalid_client"

Discord OAuth2設定を確認してください：

- Client IDとClient Secretが正しいか
- Redirect URIが正しく設定されているか

#### エラー: "guild_not_found"

DISCORD_GUILD_IDが正しいか確認してください：

```bash
# Discord Developer Modeを有効化し、サーバーIDをコピー
```

### データベースエラー

#### エラー: "database is locked"

Cloud Runの同時実行数を1に制限してください：

```bash
gcloud run services update jyogi-auth \
  --region asia-northeast1 \
  --max-instances 1
```

## マイグレーション

### 初回マイグレーション

デプロイスクリプトで自動実行されますが、手動で実行する場合：

```bash
# マイグレーションJob作成
gcloud run jobs create jyogi-auth-migrate \
  --image gcr.io/$GCP_PROJECT_ID/jyogi-auth:latest \
  --region asia-northeast1 \
  --set-env-vars "DATABASE_PATH=/app/data/auth.db" \
  --command /bin/sh \
  --args "-c,cd /app && ./scripts/migrate.sh up"

# 実行
gcloud run jobs execute jyogi-auth-migrate --region asia-northeast1 --wait
```

### マイグレーションのロールバック

```bash
# ロールバックJob作成
gcloud run jobs create jyogi-auth-rollback \
  --image gcr.io/$GCP_PROJECT_ID/jyogi-auth:latest \
  --region asia-northeast1 \
  --set-env-vars "DATABASE_PATH=/app/data/auth.db" \
  --command /bin/sh \
  --args "-c,cd /app && ./scripts/migrate.sh down"

# 実行
gcloud run jobs execute jyogi-auth-rollback --region asia-northeast1 --wait
```

## 環境変数の更新

デプロイ後に環境変数を更新する場合：

```bash
gcloud run services update jyogi-auth \
  --region asia-northeast1 \
  --set-env-vars "DISCORD_CLIENT_ID=new-value"
```

複数の環境変数を一度に更新：

```bash
gcloud run services update jyogi-auth \
  --region asia-northeast1 \
  --set-env-vars "DISCORD_CLIENT_ID=new-value,JWT_SECRET=new-secret"
```

## ログの確認

### リアルタイムログ

```bash
gcloud run services logs tail jyogi-auth --region asia-northeast1
```

### 過去のログを検索

```bash
# 最新100件
gcloud run services logs read jyogi-auth --region asia-northeast1 --limit 100

# エラーログのみ
gcloud run services logs read jyogi-auth --region asia-northeast1 --filter "severity=ERROR"
```

### GCP Consoleでログ確認

1. [GCP Console](https://console.cloud.google.com/) にアクセス
2. "Cloud Run" → サービス "jyogi-auth" を選択
3. "ログ" タブをクリック

## パフォーマンス最適化

### コールドスタート対策

最小インスタンス数を1に設定（有料）：

```bash
gcloud run services update jyogi-auth \
  --region asia-northeast1 \
  --min-instances 1
```

**注意**: 最小インスタンス数を1にすると、常時1インスタンスが起動するため課金が発生します（月額約$8）。

### メモリ増量

レスポンス速度を改善したい場合：

```bash
gcloud run services update jyogi-auth \
  --region asia-northeast1 \
  --memory 512Mi
```

## セキュリティ

### Secret Managerの使用（推奨）

機密情報をSecret Managerで管理する場合：

```bash
# シークレット作成
echo -n "your-discord-client-secret" | gcloud secrets create discord-client-secret --data-file=-

# Cloud Runからシークレットを参照
gcloud run services update jyogi-auth \
  --region asia-northeast1 \
  --set-secrets "DISCORD_CLIENT_SECRET=discord-client-secret:latest"
```

### HTTPS強制

Cloud RunはデフォルトでHTTPSを強制します。HTTPアクセスは自動的にHTTPSにリダイレクトされます。

## コスト管理

### 無料枠

- リクエスト: 月200万回まで無料
- vCPU: 月36万秒まで無料
- メモリ: 月18万GiB秒まで無料

### コスト確認

```bash
# 現在のコストを確認
gcloud billing accounts list
gcloud billing projects describe $GCP_PROJECT_ID
```

GCP Consoleの「請求」セクションで詳細を確認できます。

## データバックアップ（Phase 2）

現在のPhase 1では、SQLiteデータはCloud Runのローカルストレージに保存されるため、インスタンスが破棄されるとデータが消失します。

**対策**:

1. 定期的にアクセスしてインスタンスを維持
2. Phase 2でCloud Storageバックアップを実装
3. Cloud SQL（PostgreSQL）に移行

## まとめ

これでじょぎメンバー認証システムがCloud Runにデプロイされました！

次のステップ：

1. クライアントアプリから認証APIを利用
2. OAuth2 SSOでSSO統合
3. Phase 2でプロフィール同期機能を追加

ご不明な点があれば、[GitHub Issues](https://github.com/jyogi-web/jyogi-discord-auth/issues)でお問い合わせください。
