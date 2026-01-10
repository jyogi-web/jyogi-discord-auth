# データベース設計

じょぎメンバー認証システムのデータベーススキーマ定義（TiDB）。

## ER図

```
┌─────────────┐
│    User     │
└─────────────┘
      │ 1
      │
      │ *
┌─────────────┐      ┌──────────────┐      ┌──────────────┐
│   Session   │      │ ClientApp    │      │   Profile    │
└─────────────┘      └──────────────┘      └──────────────┘
      |                     │ 1                   |
      |                     │                     |
      |                     │ *                   |
      |              ┌──────────────┐             |
      |              │  AuthCode    │             |
      |              └──────────────┘             |
      |                     │ 1                   |
      |                     │                     |
      |                     │ 1                   |
      |              ┌──────────────┐             |
      |              │    Token     │             |
      |              └──────────────┘             |
      |                     │ *                   |
      |                     │                     |
      |                     │ 1                   |
      |              ┌──────────────┐             |
      └──────────────┤     User     ├─────────────┘
                     └──────────────┘
```

## エンティティ

### 1. User（ユーザー）

じょぎメンバーを表すエンティティ。

**Fields**:

- `id` (VARCHAR(36), PRIMARY KEY): ユーザーID (UUID)
- `discord_id` (VARCHAR(255), UNIQUE, NOT NULL): Discord ユーザーID
- `username` (VARCHAR(255), NOT NULL): Discordユーザー名
- `display_name` (VARCHAR(255)): 表示名
- `avatar_url` (VARCHAR(512)): アバターURL
- `guild_roles` (TEXT): ギルド内ロール（JSON文字列配列）
- `guild_nickname` (VARCHAR(255)): ギルド内ニックネーム
- `joined_at` (DATETIME): ギルド参加日時
- `created_at` (DATETIME, NOT NULL): 作成日時
- `updated_at` (DATETIME, NOT NULL): 更新日時
- `last_login_at` (DATETIME): 最終ログイン日時

**SQL**:

```sql
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    discord_id VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    avatar_url VARCHAR(512),
    guild_roles TEXT,
    guild_nickname VARCHAR(255),
    joined_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_users_joined_at ON users(joined_at);
```

### 2. Session（セッション）

ユーザーのログイン状態を表すエンティティ。

**Fields**:

- `id` (TEXT, PRIMARY KEY): セッションID (UUID)
- `user_id` (TEXT, FOREIGN KEY, NOT NULL): ユーザーID (users.id)
- `token` (TEXT, UNIQUE, NOT NULL): セッショントークン
- `expires_at` (TIMESTAMP, NOT NULL): 有効期限
- `created_at` (TIMESTAMP, NOT NULL): 作成日時

**SQL**:

```sql
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
```

### 3. ClientApp（クライアントアプリ）

認証サーバーを使用する内製ツールを表すエンティティ。

**Fields**:

- `id` (TEXT, PRIMARY KEY): クライアントID (UUID)
- `client_id` (TEXT, UNIQUE, NOT NULL): OAuth2クライアントID
- `client_secret` (TEXT, NOT NULL): OAuth2クライアントシークレット（ハッシュ化）
- `name` (TEXT, NOT NULL): アプリケーション名
- `redirect_uris` (TEXT, NOT NULL): リダイレクトURI（JSON配列形式）
- `created_at` (TIMESTAMP, NOT NULL): 作成日時
- `updated_at` (TIMESTAMP, NOT NULL): 更新日時

**SQL**:

```sql
CREATE TABLE IF NOT EXISTS client_apps (
    id TEXT PRIMARY KEY,
    client_id TEXT UNIQUE NOT NULL,
    client_secret TEXT NOT NULL,
    name TEXT NOT NULL,
    redirect_uris TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_client_apps_client_id ON client_apps(client_id);
```

### 4. AuthCode（認可コード）

OAuth2フローで使用される一時的な認可コード。

**Fields**:

- `id` (TEXT, PRIMARY KEY): 認可コードID (UUID)
- `code` (TEXT, UNIQUE, NOT NULL): 認可コード
- `client_id` (TEXT, FOREIGN KEY, NOT NULL): クライアントID (client_apps.client_id)
- `user_id` (TEXT, FOREIGN KEY, NOT NULL): ユーザーID (users.id)
- `redirect_uri` (TEXT, NOT NULL): リダイレクトURI
- `expires_at` (TIMESTAMP, NOT NULL): 有効期限
- `created_at` (TIMESTAMP, NOT NULL): 作成日時
- `used` (BOOLEAN, NOT NULL, DEFAULT 0): 使用済みフラグ

**SQL**:

```sql
CREATE TABLE IF NOT EXISTS auth_codes (
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

CREATE INDEX IF NOT EXISTS idx_auth_codes_code ON auth_codes(code);
CREATE INDEX IF NOT EXISTS idx_auth_codes_user_id ON auth_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_codes_expires_at ON auth_codes(expires_at);
```

### 5. Token（トークン）

アクセストークンとリフレッシュトークンを表すエンティティ。

**Fields**:

- `id` (TEXT, PRIMARY KEY): トークンID (UUID)
- `token` (TEXT, UNIQUE, NOT NULL): トークン値
- `token_type` (TEXT, NOT NULL): トークンタイプ (`access` または `refresh`)
- `user_id` (TEXT, FOREIGN KEY, NOT NULL): ユーザーID (users.id)
- `client_id` (TEXT, FOREIGN KEY, NOT NULL): クライアントID (client_apps.client_id)
- `expires_at` (TIMESTAMP, NOT NULL): 有効期限
- `created_at` (TIMESTAMP, NOT NULL): 作成日時
- `revoked` (BOOLEAN, NOT NULL, DEFAULT 0): 取り消しフラグ

**SQL**:

```sql
CREATE TABLE IF NOT EXISTS tokens (
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

CREATE INDEX IF NOT EXISTS idx_tokens_token ON tokens(token);
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_client_id ON tokens(client_id);
CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON tokens(expires_at);
```

### 6. Profile（プロフィール）

Discord自己紹介チャンネルから取得したユーザー情報を保存するエンティティ。

**Fields**:

- `id` (TEXT, PRIMARY KEY): プロフィールID (UUID)
- `user_id` (TEXT, FOREIGN KEY, NOT NULL): ユーザーID (users.id)
- `discord_message_id` (TEXT, UNIQUE, NOT NULL): DiscordメッセージID
- `real_name` (TEXT): 名前
- `student_id` (TEXT): 学籍番号
- `hobbies` (TEXT): 趣味
- `what_to_do` (TEXT): やりたいこと
- `comment` (TEXT): ひとこと
- `created_at` (TIMESTAMP, NOT NULL): 作成日時
- `updated_at` (TIMESTAMP, NOT NULL): 更新日時

**SQL**:

```sql
CREATE TABLE IF NOT EXISTS profiles (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    discord_message_id TEXT UNIQUE NOT NULL,
    real_name TEXT,
    student_id TEXT,
    hobbies TEXT,
    what_to_do TEXT,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_profiles_discord_message_id ON profiles(discord_message_id);
```
