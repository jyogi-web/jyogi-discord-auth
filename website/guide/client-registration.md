# クライアントアプリケーション登録

認証システムを利用する新しいアプリケーション（クライアント）を登録する手順です。
現在、管理画面は提供されていないため、データベースへの直接操作が必要です。

## 利用者の方へ

新しいアプリを作成する場合は、システム管理者に以下の情報を伝えて `CLIENT_ID` と `CLIENT_SECRET` の発行を依頼してください。

1. **アプリ名**: 例「部費管理システム」
2. **リダイレクトURI**: 認証後に戻ってくるURL (例: `http://localhost:3000/api/auth/callback`)
   - 開発用と本番用で分ける場合はそれぞれ申請してください。

---

## 管理者向け手順

`client_apps` テーブルにレコードを追加することで、クライアントを登録します。

### 1. IDとシークレットの生成

ランダムな文字列を生成します。

```bash
# CLIENT_ID生成
openssl rand -hex 16
# 例: 4a8b...

# CLIENT_SECRET生成
openssl rand -base64 32
# 例: xYz123...
```

### 2. データベースへの登録

**注意**: `client_secret` はハッシュ化せずに保存されています（OAuth2の仕様上、バックエンドアプリからの送信と照合するため）。
※ 将来的にはハッシュ化して保存する機能が追加される可能性があります。

#### 開発環境 (SQLite)

```bash
sqlite3 jyogi_auth.db
```

```sql
INSERT INTO client_apps (id, client_id, client_secret, name, redirect_uris, created_at, updated_at)
VALUES (
  hex(randomblob(16)), -- UUID (簡易的)
  'YOUR_GENERATED_CLIENT_ID',
  'YOUR_GENERATED_CLIENT_SECRET',
  'Demo App',
  '["http://localhost:3000/api/auth/callback"]',
  datetime('now'),
  datetime('now')
);
```

#### 本番環境 (TiDB / MySQL)

```sql
INSERT INTO client_apps (id, client_id, client_secret, name, redirect_uris, created_at, updated_at)
VALUES (
  UUID(),
  'YOUR_GENERATED_CLIENT_ID',
  'YOUR_GENERATED_CLIENT_SECRET',
  'Production App',
  '["https://myapp.com/api/auth/callback"]',
  NOW(),
  NOW()
);
```

### 3. 設定値の共有

生成した `CLIENT_ID` と `CLIENT_SECRET` を安全な方法で開発者に共有してください。
これらの値は `.env` ファイル等で管理されます。
