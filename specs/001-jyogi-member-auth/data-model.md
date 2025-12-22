# Data Model: じょぎメンバー認証システム

**Date**: 2025-12-22
**Feature**: じょぎメンバー認証システム
**Branch**: `001-jyogi-member-auth`

## Overview

このドキュメントでは、じょぎメンバー認証システムのデータモデルを定義します。SQLiteをデータベースとして使用し、将来的にPostgreSQLへの移行を容易にするため、標準的なSQL構文を使用します。

---

## Entity Relationship Diagram

```
┌─────────────┐
│    User     │
└─────────────┘
      │ 1
      │
      │ *
┌─────────────┐      ┌──────────────┐
│   Session   │      │ ClientApp    │
└─────────────┘      └──────────────┘
                            │ 1
                            │
                            │ *
                     ┌──────────────┐
                     │  AuthCode    │
                     └──────────────┘
                            │ 1
                            │
                            │ 1
                     ┌──────────────┐
                     │    Token     │
                     └──────────────┘
                            │ *
                            │
                            │ 1
                     ┌──────────────┐
                     │     User     │
                     └──────────────┘
```

---

## Entities

### 1. User（ユーザー）

じょぎメンバーを表すエンティティ。

**Fields**:

- `id` (UUID, PRIMARY KEY): ユーザーID
- `discord_id` (VARCHAR(255), UNIQUE, NOT NULL): Discord ユーザーID
- `username` (VARCHAR(255), NOT NULL): Discordユーザー名
- `avatar_url` (TEXT): アバターURL
- `created_at` (TIMESTAMP, NOT NULL): 作成日時
- `updated_at` (TIMESTAMP, NOT NULL): 更新日時
- `last_login_at` (TIMESTAMP): 最終ログイン日時

**Validation Rules**:

- `discord_id` はユニーク
- `username` は必須
- `created_at`, `updated_at` は自動設定

**SQL**:

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    discord_id TEXT UNIQUE NOT NULL,
    username TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

CREATE INDEX idx_users_discord_id ON users(discord_id);
```

---

### 2. Session（セッション）

ユーザーのログイン状態を表すエンティティ。

**Fields**:

- `id` (UUID, PRIMARY KEY): セッションID
- `user_id` (UUID, FOREIGN KEY, NOT NULL): ユーザーID (users.id)
- `token` (TEXT, UNIQUE, NOT NULL): セッショントークン（ランダム生成）
- `expires_at` (TIMESTAMP, NOT NULL): 有効期限
- `created_at` (TIMESTAMP, NOT NULL): 作成日時

**Validation Rules**:

- `user_id` は users テーブルの id を参照
- `token` はユニーク
- `expires_at` は作成時刻 + 設定した有効期間（例: 7日）

**SQL**:

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

---

### 3. ClientApp（クライアントアプリ）

認証サーバーを使用する内製ツールを表すエンティティ。

**Fields**:

- `id` (UUID, PRIMARY KEY): クライアントID
- `client_id` (VARCHAR(255), UNIQUE, NOT NULL): OAuth2クライアントID
- `client_secret` (TEXT, NOT NULL): OAuth2クライアントシークレット（ハッシュ化）
- `name` (VARCHAR(255), NOT NULL): アプリケーション名
- `redirect_uris` (TEXT, NOT NULL): リダイレクトURI（JSON配列形式）
- `created_at` (TIMESTAMP, NOT NULL): 作成日時
- `updated_at` (TIMESTAMP, NOT NULL): 更新日時

**Validation Rules**:

- `client_id` はユニーク
- `client_secret` はbcryptでハッシュ化して保存
- `redirect_uris` は JSON 配列形式（例: `["https://app1.example.com/callback"]`）

**SQL**:

```sql
CREATE TABLE client_apps (
    id TEXT PRIMARY KEY,
    client_id TEXT UNIQUE NOT NULL,
    client_secret TEXT NOT NULL,
    name TEXT NOT NULL,
    redirect_uris TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_client_apps_client_id ON client_apps(client_id);
```

---

### 4. AuthCode（認可コード）

OAuth2フローで使用される一時的な認可コード。

**Fields**:

- `id` (UUID, PRIMARY KEY): 認可コードID
- `code` (VARCHAR(255), UNIQUE, NOT NULL): 認可コード（ランダム生成）
- `client_id` (TEXT, FOREIGN KEY, NOT NULL): クライアントID (client_apps.client_id)
- `user_id` (UUID, FOREIGN KEY, NOT NULL): ユーザーID (users.id)
- `redirect_uri` (TEXT, NOT NULL): リダイレクトURI
- `expires_at` (TIMESTAMP, NOT NULL): 有効期限
- `created_at` (TIMESTAMP, NOT NULL): 作成日時
- `used` (BOOLEAN, NOT NULL, DEFAULT FALSE): 使用済みフラグ

**Validation Rules**:

- `code` はユニーク、10分間有効
- 使用後は `used` フラグを TRUE に設定
- 期限切れまたは使用済みのコードは無効

**SQL**:

```sql
CREATE TABLE auth_codes (
    id TEXT PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    client_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (client_id) REFERENCES client_apps(client_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_codes_code ON auth_codes(code);
CREATE INDEX idx_auth_codes_user_id ON auth_codes(user_id);
CREATE INDEX idx_auth_codes_expires_at ON auth_codes(expires_at);
```

---

### 5. Token（トークン）

アクセストークンとリフレッシュトークンを表すエンティティ。

**Fields**:

- `id` (UUID, PRIMARY KEY): トークンID
- `token` (TEXT, UNIQUE, NOT NULL): トークン値（JWT または ランダム文字列）
- `token_type` (VARCHAR(50), NOT NULL): トークンタイプ (`access` または `refresh`)
- `user_id` (UUID, FOREIGN KEY, NOT NULL): ユーザーID (users.id)
- `client_id` (TEXT, FOREIGN KEY, NOT NULL): クライアントID (client_apps.client_id)
- `expires_at` (TIMESTAMP, NOT NULL): 有効期限
- `created_at` (TIMESTAMP, NOT NULL): 作成日時
- `revoked` (BOOLEAN, NOT NULL, DEFAULT FALSE): 取り消しフラグ

**Validation Rules**:

- `token` はユニーク
- `token_type` は `access` または `refresh`
- アクセストークン: 1時間有効
- リフレッシュトークン: 30日有効
- 取り消されたトークンは無効

**SQL**:

```sql
CREATE TABLE tokens (
    id TEXT PRIMARY KEY,
    token TEXT UNIQUE NOT NULL,
    token_type TEXT NOT NULL CHECK(token_type IN ('access', 'refresh')),
    user_id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client_apps(client_id) ON DELETE CASCADE
);

CREATE INDEX idx_tokens_token ON tokens(token);
CREATE INDEX idx_tokens_user_id ON tokens(user_id);
CREATE INDEX idx_tokens_client_id ON tokens(client_id);
CREATE INDEX idx_tokens_expires_at ON tokens(expires_at);
```

---

## Relationships

### User ↔ Session (1:N)

- 1人のユーザーは複数のセッションを持つことができる（複数デバイスからのログイン）
- セッションは1人のユーザーに紐づく
- ユーザーが削除されると、関連するセッションも削除される（CASCADE）

### User ↔ AuthCode (1:N)

- 1人のユーザーは複数の認可コードを持つことができる（複数のクライアントアプリからの認証）
- 認可コードは1人のユーザーに紐づく
- ユーザーが削除されると、関連する認可コードも削除される（CASCADE）

### User ↔ Token (1:N)

- 1人のユーザーは複数のトークンを持つことができる（複数のクライアントアプリからのアクセス）
- トークンは1人のユーザーに紐づく
- ユーザーが削除されると、関連するトークンも削除される（CASCADE）

### ClientApp ↔ AuthCode (1:N)

- 1つのクライアントアプリは複数の認可コードを発行できる
- 認可コードは1つのクライアントアプリに紐づく
- クライアントアプリが削除されると、関連する認可コードも削除される（CASCADE）

### ClientApp ↔ Token (1:N)

- 1つのクライアントアプリは複数のトークンを発行できる
- トークンは1つのクライアントアプリに紐づく
- クライアントアプリが削除されると、関連するトークンも削除される（CASCADE）

---

## Migration Strategy

### Initial Migration (001_init.sql)

すべてのテーブルを作成します。

```sql
-- See SQL definitions above
```

### Future Migrations

- **002_add_clients.sql**: 初期クライアントアプリを登録
- **003_add_indexes.sql**: 追加のインデックス（パフォーマンス最適化）
- **004_add_user_roles.sql**: ユーザーロール機能追加（Phase 2）

---

## Data Retention Policy

### セッション

- 期限切れセッションは定期的に削除（cron or バックグラウンドタスク）
- 削除間隔: 1日1回

### 認可コード

- 期限切れまたは使用済みコードは定期的に削除
- 削除間隔: 1時間1回

### トークン

- 期限切れトークンは定期的に削除
- 削除間隔: 1日1回
- 取り消されたトークンも削除対象

---

## Security Considerations

1. **パスワード/シークレットのハッシュ化**:
   - `client_secret` は bcrypt でハッシュ化
   - コスト: 12（推奨）

2. **トークンのランダム性**:
   - セッショントークン、認可コード: `crypto/rand` で生成（32バイト以上）
   - JWT: 署名鍵は環境変数で管理、ローテーション可能に

3. **インデックス**:
   - 頻繁に検索されるフィールド（`discord_id`, `token`, `code`）にインデックスを設定
   - `expires_at` にもインデックス（期限切れデータの削除を高速化）

4. **外部キー制約**:
   - CASCADE削除で、ユーザー削除時に関連データも自動削除
   - データ整合性を保つ

---

## Database Abstraction Layer

将来的なPostgreSQLへの移行を容易にするため、リポジトリパターンを使用します。

**Interface Example**:

```go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
    GetByDiscordID(ctx context.Context, discordID string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
}
```

**Implementation**:

- `internal/repository/sqlite/user.go`: SQLite実装
- `internal/repository/postgres/user.go`: PostgreSQL実装（将来）
- `internal/repository/memory/user.go`: インメモリ実装（テスト用）

---

## Summary

データモデルは5つのエンティティで構成されます：

1. **User**: じょぎメンバー
2. **Session**: ログインセッション
3. **ClientApp**: 内製ツール（クライアントアプリ）
4. **AuthCode**: OAuth2認可コード
5. **Token**: アクセス/リフレッシュトークン

すべてのテーブルにインデックスを設定し、外部キー制約でデータ整合性を保ちます。リポジトリパターンで、将来的なPostgreSQLへの移行を容易にします。

次のステップ: API契約定義（contracts/）
