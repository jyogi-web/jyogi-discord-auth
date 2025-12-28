# プロフィール同期機能

## 概要

Discord自己紹介チャンネルからメンバーのプロフィール情報を自動的に取得してデータベースに保存する機能です。

Botをサーバーに導入して下さい

<https://discord.com/oauth2/authorize?client_id=1452527263500337353&permissions=66560&scope=bot>

## アーキテクチャ

```
Discord API → ProfileService → ProfileRepository → SQLite
     ↓
  Parser
```

### コンポーネント

1. **Discord API Client** (`pkg/discord/client.go`)
   - チャンネルメッセージを取得

2. **Profile Parser** (`pkg/discord/parser.go`)
   - メッセージから構造化データを抽出

3. **Profile Service** (`internal/service/profile.go`)
   - 同期ロジックを管理

4. **Scheduler** (`internal/service/scheduler.go`)
   - 定期実行を制御

5. **Profile Repository** (`internal/repository/sqlite/profile.go`)
   - データベース操作

## データモデル

### Profile

```go
type Profile struct {
    ID               string    // UUID
    UserID           string    // ユーザーID（外部キー）
    DiscordMessageID string    // DiscordメッセージID（一意）
    RealName         string    // 本名
    StudentID        string    // 学籍番号
    Hobbies          string    // 趣味
    WhatToDo         string    // じょぎでやりたいこと
    Comment          string    // ひとこと
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

## 対応フォーマット

パーサーは以下のようなバリエーションに対応しています:

### 基本フォーマット

```
⭕本名: じょぎ太郎
⭕学籍番号: 20X1234
⭕趣味: カラオケ、ゲーム、アニメ鑑賞
⭕じょぎでやりたいこと: ゲーム作成
⭕ひとこと: よろしくお願いします！
```

### サポートする記号

- `⭕` (Check mark button)
- `○` (White circle)
- `◯` (Large circle)
- 記号なし

### サポートする区切り文字

- `:` (半角コロン)
- `：` (全角コロン)

### その他の柔軟性

- スペースの有無を許容
- 一部フィールドのみの投稿も許容
- 複数行対応

## 使用方法

### 環境変数設定

```bash
# .env ファイルに追加
DISCORD_BOT_TOKEN=your_bot_token_here
DISCORD_PROFILE_CHANNEL=channel_id_here
```

### 1回のみ実行

```bash
# Makefileを使用
make sync-profiles

# または直接実行
go run ./cmd/sync-profiles -once
```

### 定期実行（デーモンモード）

```bash
# デフォルト（60分間隔）
make sync-profiles-daemon

# カスタム間隔（例: 30分間隔）
go run ./cmd/sync-profiles -interval 30
```

### ビルドして実行

```bash
# ビルド
make build-sync-profiles

# 実行
./bin/sync-profiles -once
```

## 同期処理の流れ

1. Discord APIからチャンネルメッセージを取得（最大100件）
2. 各メッセージをパースしてプロフィールデータを抽出
3. 有効なプロフィールのみを処理
4. Discord IDでユーザーを検索
5. ユーザーが存在しない場合は新規作成
6. プロフィールをUpsert（存在すれば更新、なければ作成）
7. 結果をログに出力

## ログ出力例

```
Starting profile synchronization...
Retrieved 28 messages from channel
Created new user: koba (discord_id: 123456789)
Synced profile for user koba (message: 987654321)
Skipping message 111222333: not a valid profile
Profile synchronization completed: 25 success, 3 skipped, 0 errors
```

## エラーハンドリング

- Discord API呼び出しエラー: ログに記録して次の同期を待つ
- パースエラー: メッセージをスキップして続行
- データベースエラー: ログに記録して次のメッセージを処理

## パフォーマンス考慮事項

- Discord APIレート制限: 1チャンネルあたり最大100メッセージ/リクエスト
- 同期間隔: デフォルト60分（調整可能）
- データベース操作: Upsertでトランザクション数を最小化

## セキュリティ

- Bot Token は環境変数で管理
- データベースアクセスは内部リポジトリ層でカプセル化
- SQLインジェクション対策: パラメータ化クエリを使用

## 今後の拡張案

1. **Webhook対応**
   - Discord Webhookを使用してリアルタイム同期

2. **差分検出**
   - 変更があったメッセージのみを更新

3. **履歴管理**
   - プロフィール変更履歴を保存

4. **通知機能**
   - 同期エラー時の通知

5. **管理画面**
   - プロフィール一覧・編集機能

## トラブルシューティング

### Bot Tokenエラー

```
Error: DISCORD_BOT_TOKEN is required
```

→ `.env`ファイルに`DISCORD_BOT_TOKEN`を設定してください

### チャンネルアクセスエラー

```
discord API returned status 403: Missing Access
```

→ Botに以下の権限が付与されているか確認:

- `View Channels`
- `Read Message History`

### パースエラーが多い

```
Skipping message XXX: not a valid profile
```

→ チャンネルに無関係なメッセージが含まれている可能性があります。これは正常な動作です。

## テスト

パーサーのテストを実行:

```bash
go test -v ./pkg/discord/...
```

カバレッジ付き:

```bash
go test -v -cover ./pkg/discord/...
```

## 参考資料

- [Discord API Documentation](https://discord.com/developers/docs/intro)
- [Discord Bot Permissions Calculator](https://discordapi.com/permissions.html)
