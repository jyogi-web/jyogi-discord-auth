# トラブルシューティング

よくあるエラーとその対処法について説明します。

## 認証関連

### "access_denied" エラー

Discordの認証画面で「キャンセル」をクリックした場合や、Discord側で認可を拒否された場合に発生します。
ユーザーが意図的にキャンセルした場合は問題ありませんが、予期せず発生する場合は `DISCORD_CLIENT_ID` が正しいか確認してください。

### "redirect_uri_mismatch" エラー

Discord Developer Portalに登録されているリダイレクトURIと、アプリケーションが送信している `redirect_uri` が一致していません。
- `.env` の `DISCORD_REDIRECT_URI` を確認してください。
- Developer Portalの「OAuth2」設定で、リダイレクトURIが完全一致（末尾のスラッシュ有無も含む）しているか確認してください。

### JWT検証エラー (401 Unauthorized)

APIリクエスト時に `Authorization: Bearer <token>` ヘッダーが正しく設定されているか確認してください。
また、トークンの有効期限切れ（デフォルト24時間）の可能性もあります。 `/token/refresh` エンドポイントで更新を試みてください。

## データベース・起動関連

### データベース接続エラー

```
Failed to connect to TiDB ...
```

- **開発環境 (SQLite)**: フォルダの書き込み権限があるか確認してください。
- **本番環境 (TiDB)**: `TIDB_DB_HOST`, `TIDB_DB_USERNAME`, `TIDB_DB_PASSWORD` が正しいか確認してください。Cloud Runから接続する場合、VPCコネクタの設定やIP制限も確認が必要です。

### ポート競合

```
bind: address already in use
```

`SERVER_PORT`（デフォルト8080）で指定したポートが既に使用されています。
別のプロセスを停止するか、`.env` でポート番号を変更してください。

## CORSエラー

ブラウザのコンソールにCORS関連のエラーが表示される場合：

- `.env` の `CORS_ALLOWED_ORIGINS` に、リクエスト元のオリジン（例: `http://localhost:3000`）が含まれているか確認してください。
- プロトコル（http/https）やポート番号も含めて完全一致する必要があります。
