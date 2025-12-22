# Implementation Plan: じょぎメンバー認証システム

**Branch**: `001-jyogi-member-auth` | **Date**: 2025-12-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-jyogi-member-auth/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Discord OAuth2を使用したじょぎメンバー専用の認証システム。Discordアカウントでログインし、じょぎサーバーのメンバーシップを確認後、JWTを発行する。他の内製ツールがSSOとして利用できる認証基盤を提供する。

**技術アプローチ**:

- Goで実装（net/http標準ライブラリベース + 最小限の外部依存）
- SQLiteをデータベースとして使用（抽象化層を設けて将来の移行を容易に）
- Discord OAuth2フローの実装
- JWTベースの認証トークン管理
- 無料運用を実現

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**:

- `github.com/golang-jwt/jwt` - JWT生成・検証
- `github.com/mattn/go-sqlite3` - SQLiteドライバ
- `golang.org/x/oauth2` - OAuth2クライアント
- `github.com/bwmarrin/discordgo` - Discord API クライアント（オプション）

**Storage**: SQLite（ファイルベース、抽象化層経由）
**Testing**: Go標準テスト（`testing`パッケージ）、テーブル駆動テスト
**Target Platform**: Linux/macOS/Windows サーバー（Docker対応）
**Project Type**: 単一Webアプリケーション（API + 最小限のHTML）
**Performance Goals**:

- 100同時ログインリクエスト処理可能
- JWT検証 < 10ms
- 認証フロー完了 < 30秒（Discord応答時間含む）

**Constraints**:

- 低コスト運用（月額$5-10程度、Fly.io/Railway/低価格VPS等のクラウドサービス利用）
- 200~500ユーザー規模
- 同時接続10~50人想定

**Scale/Scope**:

- ユーザー数: 200~500人
- エンドポイント数: ~10個（認証、トークン、検証、管理）
- データベーステーブル: 5個（users, sessions, clients, auth_codes, tokens）

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Status**: ✅ PASS

プロジェクト憲章（`.specify/memory/constitution.md`）に準拠：

- **シンプルさ優先**: Go標準ライブラリを優先、外部依存は最小限、YAGNI原則
- **テスト駆動開発（TDD）**: Red-Green-Refactorサイクル、テーブル駆動テスト、テストカバレッジ目標80%以上
- **レイヤードアーキテクチャ**: Handler → Service → Repository → DB/External API
- **リポジトリパターン**: データベース抽象化でSQLite→PostgreSQL移行を容易に
- **依存性注入**: コンストラクタで依存を注入、グローバル変数を避ける
- **セキュリティ基準**: JWT検証、入力検証、環境変数管理
- **ドキュメント**: GoDocコメント必須、コードレビュー基準準拠

この計画は憲章の全原則に準拠しています。

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
jyogi-discord-auth/
├── cmd/
│   └── server/
│       └── main.go              # エントリーポイント（HTTPサーバー起動）
│
├── internal/
│   ├── domain/                  # ドメインモデル（エンティティ）
│   │   ├── user.go
│   │   ├── session.go
│   │   ├── client.go
│   │   ├── auth_code.go
│   │   └── token.go
│   │
│   ├── repository/              # データアクセス層（DB抽象化）
│   │   ├── interface.go         # リポジトリインターフェース
│   │   ├── sqlite/              # SQLite実装
│   │   │   ├── user.go
│   │   │   ├── session.go
│   │   │   ├── client.go
│   │   │   ├── auth_code.go
│   │   │   └── token.go
│   │   └── memory/              # インメモリ実装（テスト用）
│   │       └── ...
│   │
│   ├── service/                 # ビジネスロジック層
│   │   ├── auth.go              # 認証サービス（Discord OAuth2）
│   │   ├── token.go             # トークン管理サービス（JWT）
│   │   ├── membership.go        # メンバーシップ確認サービス
│   │   └── session.go           # セッション管理サービス
│   │
│   ├── handler/                 # HTTPハンドラー（API エンドポイント）
│   │   ├── auth.go              # /auth/login, /auth/callback
│   │   ├── token.go             # /token, /token/refresh
│   │   ├── oauth.go             # /oauth/authorize, /oauth/token
│   │   └── api.go               # /api/verify, /api/user
│   │
│   ├── middleware/              # HTTPミドルウェア
│   │   ├── auth.go              # JWT検証ミドルウェア
│   │   ├── cors.go              # CORS設定
│   │   └── logging.go           # ログ記録
│   │
│   └── config/                  # 設定管理
│       └── config.go            # 環境変数読み込み
│
├── pkg/                         # 公開パッケージ（将来的にライブラリ化可能）
│   ├── discord/                 # Discord APIクライアント
│   │   └── client.go
│   └── jwt/                     # JWTユーティリティ
│       └── jwt.go
│
├── web/                         # 静的ファイル（HTML/CSS/JS）
│   ├── templates/
│   │   ├── login.html
│   │   └── dashboard.html
│   └── static/
│       ├── css/
│       └── js/
│
├── migrations/                  # データベースマイグレーション
│   ├── 001_init.sql
│   ├── 002_add_clients.sql
│   └── ...
│
├── scripts/                     # 開発・運用スクリプト
│   ├── setup.sh
│   └── migrate.sh
│
├── tests/                       # テスト
│   ├── integration/             # 統合テスト
│   │   ├── auth_flow_test.go
│   │   └── oauth_flow_test.go
│   └── testdata/                # テストデータ
│
├── .env.example                 # 環境変数サンプル
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
└── README.md
```

**Structure Decision**: 標準的なGoプロジェクト構成を採用。

- **`cmd/`**: アプリケーションエントリーポイント
- **`internal/`**: 内部パッケージ（外部からインポート不可）
  - `domain`: ドメインモデル（エンティティ定義）
  - `repository`: データアクセス層（DB抽象化で将来の移行を容易に）
  - `service`: ビジネスロジック
  - `handler`: HTTPハンドラー
  - `middleware`: HTTPミドルウェア
  - `config`: 設定管理
- **`pkg/`**: 公開パッケージ（将来的にライブラリとして再利用可能）
- **`web/`**: 静的ファイル（HTML/CSS/JS）
- **`migrations/`**: SQLマイグレーション
- **`tests/`**: 統合テスト

この構成により、明確な責務分離と、将来的なPostgreSQLへの移行が容易になります。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

**Status**: ✅ PASS - No constitution violations

この計画はプロジェクト憲章（`.specify/memory/constitution.md`）の全原則に準拠しており、違反はありません。複雑性の追加は必要な技術的要件（リポジトリパターン、レイヤードアーキテクチャ）によるもので、憲章で承認されています。

---

## Phase 0: Research - ✅ COMPLETE

**Output**: `research.md`

すべての技術的な調査が完了しました：

- Discord OAuth2実装パターン（golang.org/x/oauth2）
- JWT生成・検証ライブラリ（golang-jwt/jwt）
- SQLiteドライバーとDB抽象化
- Discord APIによるメンバーシップ確認
- セッション管理戦略
- 環境変数管理
- HTTPS/HTTP制御

---

## Phase 1: Design & Contracts - ✅ COMPLETE

**Output**: `data-model.md`, `contracts/api.md`, `quickstart.md`

### Data Model

5つのエンティティを定義：

1. User - じょぎメンバー
2. Session - ログインセッション
3. ClientApp - 内製ツール（クライアントアプリ）
4. AuthCode - OAuth2認可コード
5. Token - アクセス/リフレッシュトークン

すべてのテーブルにインデックス、外部キー制約を設定。リポジトリパターンで将来のPostgreSQL移行を容易に。

### API Contracts

9個のRESTful APIエンドポイントを定義：

- 認証: `/auth/login`, `/auth/callback`, `/auth/logout`
- トークン: `/token`, `/token/refresh`
- OAuth2: `/oauth/authorize`, `/oauth/token`
- API: `/api/verify`, `/api/user`

### Test Scenarios

5つのユーザーストーリーをカバーする13個のテストケースを定義。統合テスト、パフォーマンステストのシナリオも含む。

---

## Next Steps

計画フェーズが完了しました。次は実装タスクの分解です：

```bash
/speckit.tasks
```

これで、実装可能なタスクリストが生成されます。
