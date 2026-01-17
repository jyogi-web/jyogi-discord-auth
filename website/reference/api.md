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

### 現在のユーザー情報取得

セッション認証を使用して、現在ログインしているユーザーの情報を取得します。

**Endpoint:** `GET /api/me`

**Authentication:** セッションCookie (`session_token`)

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "discord_id": "123456789012345678",
  "username": "jyogi_taro",
  "avatar_url": "https://cdn.discordapp.com/avatars/...",
  "last_login_at": "2024-01-01T12:00:00Z"
}
```

**Error Response:**

```json
{
  "error": "unauthorized",
  "message": "No active session"
}
```

**Example:**

```bash
curl http://localhost:8080/api/me \
  -H "Cookie: session_token=..."
```

### メンバー一覧取得

じょぎサーバーのメンバー一覧をプロフィール情報付きで取得します。ページネーションに対応しています。

**Endpoint:** `GET /api/members`

**Authentication:** セッションCookie (`session_token`)

**Parameters:**

| Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `limit` | integer | Optional | 取得件数（デフォルト: 50、最大: 100） |
| `offset` | integer | Optional | オフセット（デフォルト: 0） |

**Response:**

```json
{
  "members": [
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
        "comment": "よろしくお願いします!"
      }
    }
  ],
  "limit": 50,
  "offset": 0,
  "count": 1
}
```

**Example:**

```bash
curl http://localhost:8080/api/members?limit=10&offset=0 \
  -H "Cookie: session_token=..."
```

## OAuth2 (SSO)

クライアントアプリケーション向けのOAuth2エンドポイントです。

### 認可エンドポイント

OAuth2認可フローを開始します。ユーザーがログインしていない場合は、自動的に `/auth/login` にリダイレクトされます。

**Endpoint:** `GET /oauth/authorize`

**Parameters:**

| Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `client_id` | string | Yes | クライアントID |
| `redirect_uri` | string | Yes | リダイレクトURI（事前に登録されたURIのみ許可） |
| `response_type` | string | Yes | `code` 固定 |
| `state` | string | Optional | CSRF対策用文字列（推奨） |

**Response (ユーザー未ログイン):**

```
HTTP/1.1 302 Found
Location: /auth/login
```

**Response (ユーザーログイン済み):**

```
HTTP/1.1 302 Found
Location: {redirect_uri}?code={authorization_code}&state={state}
```

**Example:**

```bash
# ブラウザでアクセス
http://localhost:8080/oauth/authorize?client_id=your_client_id&redirect_uri=http://localhost:3000/callback&response_type=code&state=random_state_string
```

### トークンエンドポイント

認可コードをアクセストークンとリフレッシュトークンに交換します。

**Endpoint:** `POST /oauth/token`

**Content-Type:** `application/x-www-form-urlencoded`

**Parameters (Form Data):**

| Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `grant_type` | string | Yes | `authorization_code` 固定 |
| `code` | string | Yes | 認可コード |
| `client_id` | string | Yes | クライアントID |
| `client_secret` | string | Yes | クライアントシークレット |
| `redirect_uri` | string | Yes | 認可時に使用したリダイレクトURI |

**Response:**

```json
{
  "access_token": "eyJhbG...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "def..."
}
```

**Error Response:**

```json
{
  "error": "invalid_grant",
  "error_description": "authorization code expired"
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=SplxlOBeZQQYbYS6WxSbIA" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "client_id=CLIENT_ID" \
  -d "client_secret=CLIENT_SECRET"
```

**注意:**
- 認可コードは10分間有効で、一度のみ使用可能です
- アクセストークンは1時間有効です
- リフレッシュトークンは7日間有効です
- `grant_type=refresh_token` によるトークン更新は現在未実装です

### ユーザー情報エンドポイント

アクセストークンに紐づくユーザー情報を取得します。プロフィール同期機能により、Discordの自己紹介チャンネルの内容も含まれます。

**Endpoint:** `GET /oauth/userinfo`

**Headers:**

```
Authorization: Bearer {access_token}
```

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
    "comment": "よろしくお願いします!"
  }
}
```

**Error Response:**

```json
{
  "error": "invalid_token",
  "message": "Token is invalid or expired"
}
```

**Example:**

```bash
curl http://localhost:8080/oauth/userinfo \
  -H "Authorization: Bearer eyJhbG..."
```

### トークン検証（未実装）

OAuth2アクセストークンの検証エンドポイント。

**Endpoint:** `GET /oauth/verify`

**Status:** 未実装

**Response:**

```json
{
  "error": "not_implemented",
  "error_description": "token verification not yet implemented"
}
```

**注意:**
- このエンドポイントは現在未実装です
- JWT検証には `/api/verify` を使用してください

## トークン (Token)

セッション認証を使用してJWTトークンを発行・更新するエンドポイントです。

### JWT発行

セッショントークン（Cookie）を使用してJWTアクセストークンを発行します。

**Endpoint:** `POST /token`

**Authentication:** セッションCookie (`session_token`)

**Response:**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 604800
}
```

**Error Response:**

```json
{
  "error": "unauthorized",
  "message": "No active session"
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/token \
  -H "Cookie: session_token=..."
```

**注意:**
- セッション認証が必要です（Cookieに`session_token`が必要）
- 発行されたJWTは7日間有効です

### JWT更新

既存のJWTアクセストークンを使用して新しいJWTを発行します。

**Endpoint:** `POST /token/refresh`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 604800
}
```

**Error Response:**

```json
{
  "error": "invalid_token",
  "message": "Token is invalid or expired"
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/token/refresh \
  -H "Authorization: Bearer eyJhbG..."
```

**注意:**
- 既存のアクセストークンが必要です
- トークンを検証後、新しいJWTを発行します（7日間有効）
- OAuth2の`refresh_token`とは異なり、既存のアクセストークンを使用します

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

### 特定ユーザーの情報取得

指定されたIDのユーザー情報を取得します。プロフィール同期機能により、Discordの自己紹介チャンネルの内容も含まれます。

**Endpoint:** `GET /api/user/{id}`

**Parameters:**

| Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | string | Yes | ユーザーID（URLパス） |

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

**Error Response:**

```json
{
  "error": "user_not_found",
  "message": "User not found"
}
```

**Example:**

```bash
curl http://localhost:8080/api/user/550e8400-e29b-41d4-a716-446655440000 \
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