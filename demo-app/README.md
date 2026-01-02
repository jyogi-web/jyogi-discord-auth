# じょぎ認証デモアプリ

じょぎDiscord認証サーバーのデモンストレーション用Next.jsアプリケーションです。

## 機能

- Discord OAuth2ログイン
- プロフィール表示
- じょぎメンバー一覧

## 技術スタック

- Next.js 16 LTS (App Router)
- TypeScript
- Tailwind CSS

## セットアップ

### 1. 依存関係のインストール

```bash
npm install
```

### 2. 環境変数の設定

`.env.local.example`をコピーして`.env.local`を作成:

```bash
cp .env.local.example .env.local
```

`.env.local`を編集して、認証サーバーのURLを設定:

```bash
# 認証サーバーURL（ローカル開発時）
NEXT_PUBLIC_AUTH_SERVER_URL=http://localhost:8080

# アプリURL
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

### 3. 認証サーバーの設定

#### 3.1 認証サーバーを起動

別のターミナルで認証サーバーを起動:

```bash
cd /path/to/jyogi-discord-auth
make run
```

#### 3.2 CORS設定を確認

認証サーバーの`.env`ファイルで、以下の設定が含まれていることを確認:

```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

#### 3.3 Discord Developer Portalの設定

1. [Discord Developer Portal](https://discord.com/developers/applications)にアクセス
2. アプリケーションを選択
3. OAuth2 > Redirectsで以下のURLを追加:
   ```
   http://localhost:8080/auth/callback
   ```
4. OAuth2 > Scopesで以下を選択:
   - `identify` - ユーザー情報取得
   - `guilds.members.read` - サーバーメンバーシップ確認

### 4. 開発サーバーの起動

```bash
npm run dev
```

ブラウザで http://localhost:3000 にアクセスしてください。

## 認証フロー

```
1. ユーザーが「Discordでログイン」をクリック
   ↓
2. 認証サーバー (localhost:8080/auth/login) にリダイレクト
   ↓
3. Discord OAuth2認証画面が表示される
   ↓
4. 認証成功後、認証サーバーがセッションCookieを発行
   ↓
5. デモアプリの /auth/callback にリダイレクト
   ↓
6. ダッシュボード (/dashboard) に遷移
   ↓
7. /api/me と /api/members を呼び出してデータを表示
```

## トラブルシューティング

### ログインできない

1. **認証サーバーが起動しているか確認**
   ```bash
   curl http://localhost:8080/health
   ```

2. **CORS設定を確認**
   - 認証サーバーの`.env`に`CORS_ALLOWED_ORIGINS=http://localhost:3000`が含まれているか
   - 変更後は認証サーバーを再起動

3. **Discord Developer Portalの設定を確認**
   - Redirect URIに`http://localhost:8080/auth/callback`が登録されているか
   - Client IDとClient Secretが正しく設定されているか

### メンバー一覧が表示されない

1. **じょぎサーバーメンバーかどうか確認**
   - じょぎDiscordサーバーに参加していないとログインできません

2. **セッションCookieが設定されているか確認**
   - ブラウザの開発者ツールでCookieを確認
   - `session_token`が存在し、有効期限内か確認

3. **APIエラーを確認**
   - ブラウザの開発者ツールのNetworkタブで、`/api/members`のレスポンスを確認

## 本番デプロイ

### Vercelへのデプロイ

1. **Vercel CLIをインストール**
   ```bash
   npm i -g vercel
   ```

2. **デプロイ**
   ```bash
   vercel
   ```

3. **環境変数を設定**
   ```bash
   vercel env add NEXT_PUBLIC_AUTH_SERVER_URL
   # 値: https://jyogi-auth-XXXXXXX-an.a.run.app

   vercel env add NEXT_PUBLIC_APP_URL
   # 値: https://jyogi-demo.vercel.app
   ```

4. **認証サーバー側のCORS設定を更新**

   認証サーバーの環境変数に本番URLを追加:
   ```bash
   CORS_ALLOWED_ORIGINS=https://jyogi-demo.vercel.app
   ```

5. **Discord Developer PortalのRedirect URIを更新**

   本番の認証サーバーURLを追加:
   ```
   https://jyogi-auth-XXXXXXX-an.a.run.app/auth/callback
   ```

## ライセンス

MIT
