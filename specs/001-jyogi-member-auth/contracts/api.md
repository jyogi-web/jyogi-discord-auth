# API Contract: じょぎメンバー認証システム

**Date**: 2025-12-22
**Feature**: じょぎメンバー認証システム
**Branch**: `001-jyogi-member-auth`
**Base URL**: `https://auth.jyogi.example.com`

## Overview

このドキュメントでは、じょぎメンバー認証システムのREST API仕様を定義します。

---

## Authentication Endpoints

### 1. Discord Login (User Story 1: P1)

**Endpoint**: `GET /auth/login`
**Description**: Discord OAuth2ログインページにリダイレクト

**Request**:

```
GET /auth/login
```

**Response**:

```
HTTP/1.1 302 Found
Location: https://discord.com/api/oauth2/authorize?client_id=...&redirect_uri=...&response_type=code&scope=identify+guilds.members.read
```

---

### 2. Discord Callback (User Story 1 & 2: P1-P2)

**Endpoint**: `GET /auth/callback`
**Description**: Discord OAuth2コールバック。認可コードを受け取り、じょぎメンバーシップを確認後、セッションを作成

**Request**:

```
GET /auth/callback?code={authorization_code}&state={state}
```

**Response (Success - Member)**:

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "user": {
    "id": "uuid-1234",
    "discord_id": "123456789012345678",
    "username": "jyogi_member",
    "avatar_url": "https://cdn.discordapp.com/avatars/..."
  },
  "session_token": "session_token_here"
}
```

**Response (Error - Not a Member)**:

```json
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": "not_member",
  "message": "じょぎサーバーのメンバーではありません"
}
```

---

### 3. Logout (User Story 5: P5)

**Endpoint**: `POST /auth/logout`
**Description**: セッションを無効化してログアウト

**Request**:

```
POST /auth/logout
Authorization: Bearer {session_token}
```

**Response**:

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "message": "ログアウトしました"
}
```

---

## Token Endpoints (User Story 3: P3)

### 4. Issue JWT

**Endpoint**: `POST /token`
**Description**: セッショントークンを使ってJWT（アクセストークン）を発行

**Request**:

```json
POST /token
Content-Type: application/json

{
  "session_token": "session_token_here"
}
```

**Response**:

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_here"
}
```

---

### 5. Refresh Token

**Endpoint**: `POST /token/refresh`
**Description**: リフレッシュトークンを使ってアクセストークンを更新

**Request**:

```json
POST /token/refresh
Content-Type: application/json

{
  "refresh_token": "refresh_token_here"
}
```

**Response**:

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

---

## OAuth2 Endpoints (User Story 4: P4 - Client App Integration)

### 6. OAuth2 Authorize

**Endpoint**: `GET /oauth/authorize`
**Description**: クライアントアプリからの認証リクエストを受け付け、認可コードを発行

**Request**:

```
GET /oauth/authorize?client_id={client_id}&redirect_uri={redirect_uri}&response_type=code&state={state}
```

**Response (User not logged in)**:

```
HTTP/1.1 302 Found
Location: /auth/login
```

**Response (User logged in)**:

```
HTTP/1.1 302 Found
Location: {redirect_uri}?code={authorization_code}&state={state}
```

---

### 7. OAuth2 Token

**Endpoint**: `POST /oauth/token`
**Description**: 認可コードを使ってアクセストークン・リフレッシュトークンを取得

**Request**:

```
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code={authorization_code}&
redirect_uri={redirect_uri}&
client_id={client_id}&
client_secret={client_secret}
```

**Response**:

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_here"
}
```

---

## API Endpoints (Protected)

### 8. Verify JWT

**Endpoint**: `GET /api/verify`
**Description**: JWTの検証（クライアントアプリが認証状態を確認）

**Request**:

```
GET /api/verify
Authorization: Bearer {access_token}
```

**Response**:

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "valid": true,
  "user": {
    "id": "uuid-1234",
    "discord_id": "123456789012345678",
    "username": "jyogi_member"
  },
  "expires_at": "2025-12-22T12:00:00Z"
}
```

---

### 9. Get User Info

**Endpoint**: `GET /api/user`
**Description**: 認証済みユーザーの情報を取得

**Request**:

```
GET /api/user
Authorization: Bearer {access_token}
```

**Response**:

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": "uuid-1234",
  "discord_id": "123456789012345678",
  "username": "jyogi_member",
  "avatar_url": "https://cdn.discordapp.com/avatars/...",
  "created_at": "2025-01-01T00:00:00Z",
  "last_login_at": "2025-12-22T10:00:00Z"
}
```

---

## Error Responses

すべてのエンドポイントで共通のエラーレスポンス形式：

```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

**Error Codes**:

- `invalid_request`: リクエストパラメータが不正
- `unauthorized`: 認証が必要
- `forbidden`: アクセス権限なし
- `not_found`: リソースが見つからない
- `not_member`: じょぎサーバーメンバーではない
- `invalid_token`: トークンが無効または期限切れ
- `server_error`: サーバーエラー

---

## HTTP Status Codes

- `200 OK`: 成功
- `201 Created`: リソース作成成功
- `302 Found`: リダイレクト
- `400 Bad Request`: リクエストエラー
- `401 Unauthorized`: 認証エラー
- `403 Forbidden`: アクセス拒否
- `404 Not Found`: リソース未検出
- `500 Internal Server Error`: サーバーエラー

---

## Summary

合計9個のエンドポイントを定義：

**認証**:

1. `GET /auth/login` - Discord ログイン
2. `GET /auth/callback` - Discord コールバック
3. `POST /auth/logout` - ログアウト

**トークン**:
4. `POST /token` - JWT 発行
5. `POST /token/refresh` - トークン更新

**OAuth2 (SSO)**:
6. `GET /oauth/authorize` - 認可リクエスト
7. `POST /oauth/token` - トークン取得

**API (Protected)**:
8. `GET /api/verify` - JWT 検証
9. `GET /api/user` - ユーザー情報取得

次のステップ: quickstart.md作成
