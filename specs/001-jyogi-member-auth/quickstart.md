# Quickstart & Test Scenarios: じょぎメンバー認証システム

**Date**: 2025-12-22
**Feature**: じょぎメンバー認証システム
**Branch**: `001-jyogi-member-auth`

## Overview

このドキュメントでは、じょぎメンバー認証システムのクイックスタートガイドと、各ユーザーストーリーのテストシナリオを定義します。

---

## Prerequisites

### 1. Discord Developer Portal設定

1. Discord Developer Portalでアプリケーションを作成
2. OAuth2設定:
   - Redirect URIs: `http://localhost:8080/auth/callback`（開発環境）
   - Scopes: `identify`, `guilds.members.read`
3. Client ID と Client Secret を取得

### 2. 環境変数設定

`.env`ファイルを作成：

```env
# Discord OAuth2
DISCORD_CLIENT_ID=your_client_id_here
DISCORD_CLIENT_SECRET=your_client_secret_here
DISCORD_REDIRECT_URI=http://localhost:8080/auth/callback
DISCORD_GUILD_ID=your_jyogi_server_id_here

# JWT
JWT_SECRET=your_jwt_secret_here_min_32_chars

# Database
DATABASE_PATH=./jyogi_auth.db

# Server
SERVER_PORT=8080
HTTPS_ONLY=false

# Environment
ENV=development
```

### 3. データベース初期化

```bash
# マイグレーション実行
./scripts/migrate.sh
```

---

## Quick Start

### 開発環境で起動

```bash
# 依存関係インストール
go mod download

# サーバー起動
go run cmd/server/main.go
```

サーバーが起動したら、`http://localhost:8080`にアクセス。

---

## Test Scenarios

### User Story 1: Discord OAuth2ログイン (P1)

**Test Case 1.1: 正常ログインフロー**

**Steps**:

1. ブラウザで `http://localhost:8080/auth/login` にアクセス
2. Discordログインページにリダイレクトされる
3. Discordアカウントでログイン
4. 認証を許可
5. `/auth/callback` にリダイレクトされる
6. ユーザー情報（Discord ID、ユーザー名、アバター）が表示される

**Expected Result**:

- ステータスコード: 200
- レスポンス: ユーザー情報とセッショントークン

**Test Data**:

- Discordアカウント: じょぎサーバーメンバー

---

**Test Case 1.2: 認証拒否**

**Steps**:

1. `http://localhost:8080/auth/login` にアクセス
2. Discordログインページで「キャンセル」をクリック

**Expected Result**:

- ログインページに戻る
- エラーメッセージ表示: 「認証がキャンセルされました」

---

### User Story 2: じょぎサーバーメンバーシップ確認 (P2)

**Test Case 2.1: メンバーシップ確認成功**

**Steps**:

1. じょぎサーバーメンバーのDiscordアカウントでログイン
2. `/auth/callback` でメンバーシップ確認が実行される

**Expected Result**:

- ステータスコード: 200
- 認証成功、ダッシュボードにアクセス可能

**Test Data**:

- Discordアカウント: じょぎサーバーメンバー（Guild ID一致）

---

**Test Case 2.2: 非メンバーによるアクセス拒否**

**Steps**:

1. じょぎサーバー非メンバーのDiscordアカウントでログイン
2. `/auth/callback` でメンバーシップ確認が実行される

**Expected Result**:

- ステータスコード: 403
- エラーメッセージ: 「じょぎメンバーではありません」

**Test Data**:

- Discordアカウント: じょぎサーバー非メンバー

---

### User Story 3: JWT発行と検証 (P3)

**Test Case 3.1: JWT発行**

**Steps**:

1. ログイン成功後、セッショントークンを取得
2. `POST /token` でJWT発行リクエスト

**Request**:

```bash
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/json" \
  -d '{"session_token":"session_token_here"}'
```

**Expected Result**:

- ステータスコード: 200
- レスポンス: アクセストークン（JWT）、リフレッシュトークン

---

**Test Case 3.2: JWT検証**

**Steps**:

1. 発行されたJWTを使って `/api/verify` にアクセス

**Request**:

```bash
curl -X GET http://localhost:8080/api/verify \
  -H "Authorization: Bearer {access_token}"
```

**Expected Result**:

- ステータスコード: 200
- レスポンス: `{"valid": true, "user": {...}}`

---

**Test Case 3.3: 無効なJWT**

**Steps**:

1. 無効なJWTで `/api/verify` にアクセス

**Request**:

```bash
curl -X GET http://localhost:8080/api/verify \
  -H "Authorization: Bearer invalid_token"
```

**Expected Result**:

- ステータスコード: 401
- エラーメッセージ: 「トークンが無効です」

---

### User Story 4: クライアントアプリ統合（SSO） (P4)

**Test Case 4.1: OAuth2認可フロー**

**Precondition**: クライアントアプリが登録されている

**Steps**:

1. クライアントアプリから `/oauth/authorize` にリクエスト

**Request**:

```
GET /oauth/authorize?client_id={client_id}&redirect_uri=http://client.example.com/callback&response_type=code&state=random_state
```

2. ユーザーがログイン済みなら認可コードを発行
3. クライアントアプリにリダイレクト

**Expected Result**:

- ステータスコード: 302
- リダイレクト先: `http://client.example.com/callback?code={auth_code}&state=random_state`

---

**Test Case 4.2: トークン取得**

**Steps**:

1. クライアントアプリが認可コードを使って `/oauth/token` にリクエスト

**Request**:

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code={auth_code}&redirect_uri=http://client.example.com/callback&client_id={client_id}&client_secret={client_secret}"
```

**Expected Result**:

- ステータスコード: 200
- レスポンス: アクセストークン、リフレッシュトークン

---

### User Story 5: セッション管理とログアウト (P5)

**Test Case 5.1: ログアウト**

**Steps**:

1. ログイン済みユーザーが `/auth/logout` にアクセス

**Request**:

```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer {session_token}"
```

**Expected Result**:

- ステータスコード: 200
- セッションが無効化される
- 以降、保護されたエンドポイントにアクセスできない

---

**Test Case 5.2: ログアウト後のアクセス拒否**

**Steps**:

1. ログアウト後、`/api/user` にアクセス

**Request**:

```bash
curl -X GET http://localhost:8080/api/user \
  -H "Authorization: Bearer {session_token}"
```

**Expected Result**:

- ステータスコード: 401
- エラーメッセージ: 「認証が必要です」

---

## Integration Test Example

### Complete Authentication Flow

```go
// tests/integration/auth_flow_test.go
func TestCompleteAuthFlow(t *testing.T) {
    // 1. Discord OAuth2 Login
    resp := httptest.NewRequest("GET", "/auth/login", nil)
    // Assert redirect to Discord

    // 2. Discord Callback (Mock)
    resp = httptest.NewRequest("GET", "/auth/callback?code=mock_code", nil)
    // Assert user created and session token returned

    // 3. Issue JWT
    token := requestJWT(sessionToken)
    // Assert JWT issued

    // 4. Verify JWT
    resp = httptest.NewRequest("GET", "/api/verify", nil)
    resp.Header.Set("Authorization", "Bearer "+token)
    // Assert valid

    // 5. Logout
    resp = httptest.NewRequest("POST", "/auth/logout", nil)
    resp.Header.Set("Authorization", "Bearer "+sessionToken)
    // Assert session invalidated
}
```

---

## Performance Testing

### Load Test Scenario

```bash
# 100同時ログインリクエスト
ab -n 100 -c 100 http://localhost:8080/api/verify \
  -H "Authorization: Bearer {valid_token}"
```

**Expected Result**:

- すべてのリクエストが成功
- JWT検証 < 10ms
- エラー率 < 1%

---

## Summary

テストシナリオは5つのユーザーストーリーをカバー：

1. Discord OAuth2ログイン (4 test cases)
2. メンバーシップ確認 (2 test cases)
3. JWT発行と検証 (3 test cases)
4. クライアントアプリ統合 (2 test cases)
5. セッション管理とログアウト (2 test cases)

合計: 13 test cases

すべてのテストケースは、統合テストとして `tests/integration/` に実装されます。

次のステップ: `/speckit.tasks` でタスク分解
