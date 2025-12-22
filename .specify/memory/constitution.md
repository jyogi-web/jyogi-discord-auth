# Project Constitution: じょぎメンバー認証システム

**Last Updated**: 2025-12-22
**Project**: Discord OAuth2 Authentication System for Jyogi Members

## Purpose

このドキュメントは、じょぎメンバー認証システムの開発における基本原則、開発手法、コード品質基準を定義します。すべての開発者はこの憲章に従い、一貫性のある高品質なコードを維持します。

---

## Core Principles

### 1. シンプルさ優先（Simplicity First）

- **標準ライブラリを優先**: 外部依存は最小限に抑え、Go標準ライブラリを最大限活用する
- **YAGNI (You Aren't Gonna Need It)**: 現在必要な機能のみを実装し、将来の仮想的な要件のための過剰設計を避ける
- **明確な責務分離**: 各パッケージ、関数、型は単一の責務を持つ
- **抽象化は必要最小限**: 3回以上繰り返す場合のみ抽象化を検討する

### 2. テスト駆動開発（Test-Driven Development）

すべての新機能・バグ修正はTDDサイクルに従う：

1. **Red**: 失敗するテストを書く
2. **Green**: テストを通す最小限のコードを書く
3. **Refactor**: コードをリファクタリングして品質を向上させる

### 3. 保守性と可読性（Maintainability & Readability）

- **コードは書くより読まれる**: 他の開発者（未来の自分含む）が理解しやすいコードを書く
- **自己説明的な命名**: 変数、関数、型の名前で意図を明確に伝える
- **コメントは「なぜ」を説明**: 「何を」はコードで表現し、コメントは日本語で理由や背景を説明する

### 4. 段階的な拡張性（Incremental Scalability）

- **現在のスケールに最適化**: 200~500ユーザー規模に適した設計
- **将来の移行を容易に**: 抽象化層（リポジトリパターン等）で、SQLite→PostgreSQL等の移行を可能に
- **過剰なスケール対策は避ける**: 1万ユーザー想定の設計は不要

---

## Development Methodology

### Test-Driven Development (TDD)

#### テスト戦略

**テストピラミッド**:

```
        /\
       /  \  E2E Tests (統合テスト)
      /    \
     /------\ Integration Tests (サービス層テスト)
    /        \
   /----------\ Unit Tests (関数・メソッドテスト)
  /____________\
```

- **Unit Tests（最多）**: 個別の関数・メソッドをテスト
- **Integration Tests（中程度）**: サービス層、リポジトリ層の統合をテスト
- **E2E Tests（最少）**: HTTPエンドポイントから完全なフローをテスト

#### テストカバレッジ目標

- **全体**: 80%以上
- **ビジネスロジック（service層）**: 90%以上
- **HTTPハンドラー**: 70%以上
- **リポジトリ**: 85%以上

#### テストの書き方

**テーブル駆動テスト（Table-Driven Tests）**を標準とする：

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   User
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid user",
            input:   User{DiscordID: "123", Username: "test"},
            wantErr: false,
        },
        {
            name:    "missing discord_id",
            input:   User{Username: "test"},
            wantErr: true,
            errMsg:  "discord_id is required",
        },
        // more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.input.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.wantErr && err.Error() != tt.errMsg {
                t.Errorf("error message = %v, want %v", err.Error(), tt.errMsg)
            }
        })
    }
}
```

#### モック戦略

- **インターフェースベース**: リポジトリ、外部APIクライアントはインターフェースで定義
- **テスト用インメモリ実装**: `internal/repository/memory/` にテスト用の実装を配置
- **外部依存のモック**: Discord API等の外部依存は`httptest`でモック

---

## Go Best Practices

### コーディング規約

#### 命名規則

- **パッケージ名**: 小文字、単数形、短く明確（例: `user`, `auth`, `token`）
- **関数・メソッド**: キャメルケース、動詞で始める（例: `CreateUser`, `ValidateToken`）
- **変数**: キャメルケース、名詞（例: `userID`, `sessionToken`）
- **定数**: キャメルケース、プレフィックスで分類（例: `TokenTypeAccess`, `ErrorCodeInvalidRequest`）
- **インターフェース**: 末尾に`-er`（例: `UserRepository`, `TokenValidator`）

#### エラーハンドリング

```go
// ✅ Good: エラーをラップして文脈を追加
if err := repo.Create(ctx, user); err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ✅ Good: カスタムエラー型で詳細情報を提供
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ❌ Bad: エラーを無視
_ = repo.Create(ctx, user)

// ❌ Bad: パニックを使う（回復不可能なエラー以外）
if err != nil {
    panic(err)
}
```

#### コンテキスト（context.Context）の使用

- すべてのI/O操作、外部API呼び出しに`context.Context`を渡す
- タイムアウト、キャンセルを適切に処理する

```go
func (s *AuthService) Login(ctx context.Context, code string) (*User, error) {
    // タイムアウト設定
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Discord APIコール
    token, err := s.oauth.Exchange(ctx, code)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code: %w", err)
    }

    // ...
}
```

#### Goルーチンとチャネル

- Goルーチンのリークを防ぐため、必ず終了条件を設ける
- `sync.WaitGroup`や`context.Context`で適切に管理

```go
// ✅ Good: コンテキストでGoルーチンをキャンセル
func (s *SessionService) StartCleanup(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return // キャンセルされたら終了
        case <-ticker.C:
            if err := s.cleanupExpiredSessions(ctx); err != nil {
                log.Printf("cleanup error: %v", err)
            }
        }
    }
}
```

### 依存性管理

- **最小限の外部依存**: 外部ライブラリは必要最小限に
- **バージョン固定**: `go.mod`で依存バージョンを明示的に管理
- **定期的な更新**: セキュリティパッチは迅速に適用

**承認済み依存ライブラリ**:

- `golang.org/x/oauth2` - OAuth2クライアント
- `github.com/golang-jwt/jwt/v5` - JWT生成・検証
- `github.com/mattn/go-sqlite3` - SQLiteドライバ
- `github.com/joho/godotenv` - 環境変数管理（開発環境のみ）

新規依存を追加する場合は、必要性を検討し、代替案を評価すること。

---

## Architecture Standards

### レイヤードアーキテクチャ

```
┌─────────────────────────────────────────┐
│  HTTP Handler Layer                     │  ← HTTPリクエスト処理
│  (internal/handler/)                    │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Service Layer                          │  ← ビジネスロジック
│  (internal/service/)                    │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Repository Layer                       │  ← データアクセス抽象化
│  (internal/repository/)                 │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Database / External API                │  ← SQLite, Discord API
└─────────────────────────────────────────┘
```

**責務**:

- **Handler**: HTTPリクエストのパース、バリデーション、レスポンス生成
- **Service**: ビジネスロジック、複数リポジトリの調整
- **Repository**: データアクセス、DB操作の抽象化

### リポジトリパターン

すべてのデータアクセスはインターフェースで定義：

```go
// internal/repository/interface.go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
    GetByDiscordID(ctx context.Context, discordID string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
}

// internal/repository/sqlite/user.go
type sqliteUserRepo struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
    return &sqliteUserRepo{db: db}
}
```

**利点**:

- テスト容易性（モック実装が簡単）
- データベース移行の容易性（SQLite → PostgreSQL）
- ビジネスロジックとデータアクセスの分離

### 依存性注入（Dependency Injection）

- コンストラクタ関数で依存を注入
- グローバル変数を避ける

```go
// ✅ Good: 依存性注入
type AuthService struct {
    userRepo       repository.UserRepository
    sessionRepo    repository.SessionRepository
    oauth          *oauth2.Config
    discordClient  *discord.Client
}

func NewAuthService(
    userRepo repository.UserRepository,
    sessionRepo repository.SessionRepository,
    oauth *oauth2.Config,
    discordClient *discord.Client,
) *AuthService {
    return &AuthService{
        userRepo:      userRepo,
        sessionRepo:   sessionRepo,
        oauth:         oauth,
        discordClient: discordClient,
    }
}
```

---

## Code Quality Standards

### コードレビュー基準

すべてのPRは以下の基準を満たす必要がある：

- [ ] テストが追加されている（新機能・バグ修正）
- [ ] すべてのテストがパスする（`go test ./...`）
- [ ] `go vet`でエラーなし
- [ ] `gofmt`でフォーマット済み
- [ ] GoDocコメントが追加されている（公開関数・型）
- [ ] エラーハンドリングが適切
- [ ] リソースリークなし（ファイル、DB接続、Goルーチン）

### ドキュメント

#### GoDoc

公開関数・型には必ずGoDocコメントを追加：

```go
// CreateUser creates a new user in the database.
// It returns an error if the user already exists or if the database operation fails.
func (r *sqliteUserRepo) CreateUser(ctx context.Context, user *domain.User) error {
    // ...
}
```

#### README

各パッケージに`README.md`を追加（必要に応じて）：

- パッケージの目的
- 使用例
- 重要な注意事項

---

## Security Standards

### 認証・認可

- すべての保護エンドポイントにJWT検証ミドルウェアを適用
- トークンの有効期限を適切に設定（アクセストークン: 1時間、リフレッシュトークン: 30日）
- クライアントシークレットはbcryptでハッシュ化（コスト: 12）

### 入力検証

- すべてのユーザー入力を検証
- SQLインジェクション対策（プリペアドステートメント使用）
- XSS対策（HTMLエスケープ）

### 環境変数管理

- **開発環境**: `.env`ファイル（`.gitignore`に追加）
- **本番環境**: 環境変数を直接設定（Docker Secrets、クラウドプロバイダーの環境変数機能）
- シークレット（`JWT_SECRET`, `DISCORD_CLIENT_SECRET`）は絶対にコミットしない

---

## Performance Standards

### 目標

- **JWT検証**: < 10ms
- **認証フロー完了**: < 30秒（Discord応答時間含む）
- **同時リクエスト処理**: 100リクエスト/秒

### 最適化戦略

- **データベースインデックス**: 頻繁に検索されるフィールド（`discord_id`, `token`, `code`, `expires_at`）
- **コネクションプーリング**: `sql.DB`で適切なプール設定
- **不要なデータ削除**: 期限切れセッション・トークンの定期削除

---

## Deployment Standards

### 対象環境

- **開発**: ローカルマシン（macOS/Linux/Windows）
- **本番**: 低コストクラウドサービス（Fly.io, Railway, 低価格VPS）
  - 予算: 月額$5-10程度
  - HTTPS必須
  - 環境変数で設定管理

### Docker

- Dockerfile提供（マルチステージビルドで最小イメージサイズ）
- docker-compose.yml提供（開発環境用）

### CI/CD

- GitHub Actionsでテスト自動実行
- mainブランチへのマージ前にすべてのテストがパス必須

---

## Maintenance

### データベースマイグレーション

- すべてのスキーマ変更は`migrations/`にSQLファイルで管理
- マイグレーションはバージョン管理（`001_init.sql`, `002_add_clients.sql`, ...）
- ロールバック用のダウンマイグレーションも提供

### ログ

- 構造化ログ（JSON形式推奨）
- ログレベル: DEBUG, INFO, WARN, ERROR
- 本番環境ではINFO以上のみ出力

### モニタリング

- ヘルスチェックエンドポイント（`/health`）
- メトリクス収集（将来的にPrometheus等）

---

## Violation Policy

### この憲章に違反する場合

1. **正当な理由がある場合**: PRコメントで理由を明記し、レビュアーの承認を得る
2. **一時的な技術的負債**: TODOコメントで追跡し、Issueを作成
3. **憲章の更新が必要**: このドキュメントを更新し、チームで合意

### 技術的負債の管理

- すべての`TODO`, `FIXME`, `HACK`コメントはGitHub Issueとして追跡
- 定期的に技術的負債を見直し、優先順位付け

---

## Version History

| Version | Date       | Changes                                    |
|---------|------------|--------------------------------------------|
| 1.0     | 2025-12-22 | 初版作成（TDD、Go規約、アーキテクチャ基準） |

---

**この憲章は生きたドキュメントです。プロジェクトの成長に合わせて更新していきます。**
