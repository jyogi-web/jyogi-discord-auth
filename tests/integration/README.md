# 統合テスト

じょぎメンバー認証システムの統合テストスイートです。

## 必要な環境

- Python 3.7+
- bcrypt モジュール（JWT生成テスト用）
- サーバーが起動していること（http://localhost:8080）

## セットアップ

bcryptモジュールをインストール：

```bash
pip3 install bcrypt
```

## テスト実行

### 1. サーバーを起動

```bash
# プロジェクトルートから
./jyogi_auth
# または
go run cmd/server/main.go
```

### 2. 統合テストを実行

```bash
# tests/integration/ ディレクトリから
python3 test_all_flows.py

# またはプロジェクトルートから
python3 tests/integration/test_all_flows.py
```

## テスト内容

以下の全フローを自動テストします：

### 1. ヘルスチェック
- `GET /health` エンドポイントの動作確認

### 2. JWT発行フロー
- セッショントークンからJWT発行
- `POST /token` エンドポイントのテスト

### 3. JWT検証フロー
- 発行されたJWTの検証
- `GET /api/verify` エンドポイントのテスト

### 4. JWT認証によるユーザー情報取得
- JWTを使用したユーザー情報取得
- `GET /api/user` エンドポイントのテスト

### 5. JWT更新フロー
- アクセストークンの更新
- `POST /token/refresh` エンドポイントのテスト

### 6. OAuth2/SSOフロー
- 認可コード発行（`GET /oauth/authorize`）
- トークン交換（`POST /oauth/token`）
- 完全なOAuth2 authorization code flowのテスト

### 7. ログアウトフロー
- セッション無効化
- ログアウト後のアクセス拒否確認
- `POST /auth/logout` エンドポイントのテスト

## 期待される出力

全テストが成功した場合：

```
============================================================
じょぎメンバー認証システム 統合テスト
============================================================

ℹ️  Setting up test data...
✅ Test data created:
  User ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  Session Token: test_session_1234567890
  Client ID: test_client_1234567890

============================================================
Test: Health Check
============================================================
✅ Health check passed

============================================================
Test: JWT Issuance
============================================================
✅ JWT issued successfully
  Token Type: Bearer
  Expires In: 604800 seconds
  Access Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

... (他のテスト結果)

============================================================
Test Summary
============================================================

  Health Check: PASS
  JWT Issuance: PASS
  JWT Verification: PASS
  JWT User Info: PASS
  JWT Refresh: PASS
  OAuth2/SSO Flow: PASS
  Logout Flow: PASS

Total: 7/7 tests passed

✅ All tests passed!
```

## トラブルシューティング

### サーバーが起動していない

```
urllib.error.URLError: <urlopen error [Errno 61] Connection refused>
```

**解決方法**: サーバーを起動してから再実行してください。

### bcryptモジュールがない

```
ModuleNotFoundError: No module named 'bcrypt'
```

**解決方法**: `pip3 install bcrypt` を実行してください。

### データベースが見つからない

```
sqlite3.OperationalError: unable to open database file
```

**解決方法**: プロジェクトルートディレクトリから実行していることを確認してください。

## CI/CD統合

GitHub Actionsでの実行例：

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install Python dependencies
        run: pip install bcrypt

      - name: Build server
        run: go build -o jyogi_auth cmd/server/main.go

      - name: Run migrations
        run: ./scripts/migrate.sh

      - name: Start server
        run: ./jyogi_auth &
        env:
          DISCORD_CLIENT_ID: ${{ secrets.DISCORD_CLIENT_ID }}
          DISCORD_CLIENT_SECRET: ${{ secrets.DISCORD_CLIENT_SECRET }}
          JWT_SECRET: test_secret_key

      - name: Wait for server
        run: sleep 3

      - name: Run integration tests
        run: python3 tests/integration/test_all_flows.py
```

## 注意事項

- テストは実際のデータベースを使用します
- テスト実行後、テストデータが残ります
- 本番環境では実行しないでください
