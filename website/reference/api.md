# API リファレンス

## 認証 (Auth)

### Discordログイン

DiscordのOAuth2認証フローを開始します。

**Endpoint:** `GET /auth/login`

**Parameters:**

| Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `redirect_uri` | string | Optional | 認証完了後のリダイレクト先URI |

**Example:**

```bash
# ブラウザでアクセス
http://localhost:8080/auth/login?redirect_uri=http://localhost:3000/callback
```

### Discordコールバック

Discordからのリダイレクトを受け取り、セッションを作成します（内部使用）。

**Endpoint:** `GET /auth/callback`

### ログアウト

セッションを破棄してログアウトします。

**Endpoint:** `POST /auth/logout`

**Parameters:**

| Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `redirect_url` または `redirect_uri` | string | Optional | ログアウト後のリダイレクト先URL（許可リストに登録されたURLのみ） |

**Response:**

リダイレクトURLが指定されている場合：
- 検証が通った場合: 指定されたURLにリダイレクト（307 Temporary Redirect）
- 検証が通らない場合: JSON応答（不正なURLの警告をログに記録）

リダイレクトURLが指定されていない場合：
```json
{
  "success": true,
  "message": "Logout successful"
}
```

**Example:**

```bash
# JSON応答を返す
curl -X POST http://localhost:8080/auth/logout \
  -H "Cookie: session_token=..."

# 指定されたURLにリダイレクト
curl http://localhost:8080/auth/logout?redirect_url=http://localhost:3000 \
  -H "Cookie: session_token=..."

# ブラウザでアクセス（リダイレクト）
http://localhost:8080/auth/logout?redirect_uri=http://localhost:3000/logged-out
```

**セキュリティ:**
- リダイレクトURLは `CORS_ALLOWED_ORIGINS` 環境変数に登録されたURLのみ許可されます
- 内部パス（`/` で始まるパス）は常に許可されます
- Open Redirect攻撃を防止するため、不正なURLはリダイレクトされません

## OAuth2 (SSO)

クライアントアプリケーション向けのOAuth2エンドポイントです。

### 認可エンドポイント

**Endpoint:** `GET /oauth/authorize`

**Parameters:**

| Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `client_id` | string | Yes | クライアントID |
| `redirect_uri` | string | Yes | リダイレクトURI |
| `response_type` | string | Yes | `code` 固定 |
| `state` | string | Optional | CSRF対策用文字列 |

**Example:**

```bash
http://localhost:8080/oauth/authorize?client_id=your_id&redirect_uri=...&response_type=code&state=xyz
```

### トークンエンドポイント

認可コードをアクセストークンに交換します。

**Endpoint:** `POST /oauth/token`

**Parameters (Form Data):**

| Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `grant_type` | string | Yes | `authorization_code` または `refresh_token` |
| `code` | string | Yes | 認可コード (grant_type=authorization_code時) |
| `refresh_token` | string | Yes | リフレッシュトークン (grant_type=refresh_token時) |
| `client_id` | string | Yes | クライアントID |
| `client_secret` | string | Yes | クライアントシークレット |
| `redirect_uri` | string | Yes | リダイレクトURI |

**Response:**

```json
{
  "access_token": "eyJhbG...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "def...",
  "scope": "identify"
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=SplxlOBeZQQYbYS6WxSbIA" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "client_id=CLIENT_ID" \
  -d "client_secret=CLIENT_SECRET"
```

## 保護されたリソース (Protected)

以下のエンドポイントは有効なJWTが必要です。ヘッダーに `Authorization: Bearer <token>` を付与してください。

### ユーザー情報取得

ログイン中のユーザー情報を返します。プロフィール同期機能により、Discordの自己紹介チャンネルの内容も含まれます。

**Endpoint:** `GET /api/user`

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "discord_id": "123456789012345678",
  "username": "jyogi_taro",
  "display_name": "じょぎ太郎",
  "avatar_url": "https://cdn.discordapp.com/avatars/...",
  "last_login_at": "2024-01-01T12:00:00Z",
  "guild_nickname": "太郎 [B4]",
  "guild_roles": ["111111", "222222"],
  "joined_at": "2023-04-01T09:00:00Z",
  "profile": {
    "real_name": "定規 太郎",
    "student_id": "20X1234",
    "hobbies": "プログラミング, ゲーム",
    "what_to_do": "最強の認証システムを作る",
    "comment": "よろしくお願いします！"
  }
}
```

**Example:**

```bash
curl http://localhost:8080/api/user \
  -H "Authorization: Bearer eyJhbG..."
```

### JWT検証

トークンの有効性を検証します。

**Endpoint:** `GET /api/verify`

**Response:**

```json
{
  "valid": true,
  "user_id": "550e8400-...",
  "discord_id": "1234567890...",
  "username": "jyogi_taro"
}
```