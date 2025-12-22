# Research: じょぎメンバー認証システム

**Date**: 2025-12-22
**Feature**: じょぎメンバー認証システム
**Branch**: `001-jyogi-member-auth`

## Research Questions

このドキュメントでは、実装計画で必要となる技術的な調査結果をまとめます。

---

## R1: Discord OAuth2実装パターン（Go）

**Decision**: `golang.org/x/oauth2`パッケージを使用し、Discord用の設定を追加

**Rationale**:

- Go公式のOAuth2ライブラリで、メンテナンスが活発
- Discordエンドポイントの設定が簡単
- 標準的なOAuth2フローをサポート
- トークン更新、スコープ管理が組み込まれている

**Alternatives Considered**:

- **`github.com/bwmarrin/discordgo`**: Discord専用ライブラリだが、OAuth2クライアントとしてはオーバースペック。Bot機能が中心で、OAuth2は補助的。
- **自前実装**: OAuth2フローを完全に自作。セキュリティリスクが高く、メンテナンスコストも高い。

**Implementation Notes**:

```go
import "golang.org/x/oauth2"

// Discord OAuth2 Endpoints
var discordOAuthConfig = &oauth2.Config{
    ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
    ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
    RedirectURL:  os.Getenv("DISCORD_REDIRECT_URI"),
    Scopes:       []string{"identify", "guilds.members.read"},
    Endpoint: oauth2.Endpoint{
        AuthURL:  "https://discord.com/api/oauth2/authorize",
        TokenURL: "https://discord.com/api/oauth2/token",
    },
}
```

---

## R2: JWT生成・検証ライブラリ

**Decision**: `github.com/golang-jwt/jwt/v5`を使用

**Rationale**:

- Go標準的なJWTライブラリ（元`dgrijalva/jwt-go`のフォーク）
- セキュリティアップデートが活発
- シンプルなAPI
- RS256、HS256など複数のアルゴリズムをサポート

**Alternatives Considered**:

- **`github.com/lestrrat-go/jwx`**: より高機能だが、今回の用途にはオーバースペック
- **自前実装**: セキュリティリスクが極めて高い

**Implementation Notes**:

```go
import "github.com/golang-jwt/jwt/v5"

// JWT Claims
type Claims struct {
    UserID    string `json:"user_id"`
    DiscordID string `json:"discord_id"`
    Username  string `json:"username"`
    jwt.RegisteredClaims
}

// 署名にはHS256（共有鍵）を使用
// 将来的にRS256（公開鍵/秘密鍵）への移行も検討
```

---

## R3: SQLiteドライバーとデータベース抽象化

**Decision**:

- ドライバー: `github.com/mattn/go-sqlite3`
- 抽象化: リポジトリパターン + インターフェース定義

**Rationale**:

- `go-sqlite3`はGoで最も広く使われているSQLiteドライバー
- CGOが必要だが、クロスコンパイルも対応可能
- リポジトリパターンで、将来的にPostgreSQLへの移行が容易
- インターフェースでモック作成が簡単（テスト容易性）

**Alternatives Considered**:

- **`modernc.org/sqlite`**: Pure Go実装。CGO不要だが、パフォーマンスがやや劣る。将来の選択肢として保留。
- **GORM等のORM**: 今回の規模では不要。シンプルなSQL直接実行で十分。

**Implementation Notes**:

```go
// repository/interface.go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByDiscordID(ctx context.Context, discordID string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
}

// repository/sqlite/user.go
type sqliteUserRepo struct {
    db *sql.DB
}

func (r *sqliteUserRepo) Create(ctx context.Context, user *domain.User) error {
    query := `INSERT INTO users (discord_id, username, avatar_url) VALUES (?, ?, ?)`
    _, err := r.db.ExecContext(ctx, query, user.DiscordID, user.Username, user.AvatarURL)
    return err
}
```

---

## R4: Discord APIによるメンバーシップ確認

**Decision**: Discord REST API（`/users/@me/guilds/{guild.id}/member`）を直接使用

**Rationale**:

- OAuth2トークンを使って、ユーザーの所属サーバーを確認できる
- スコープ `guilds.members.read` が必要
- シンプルなHTTPリクエストで実装可能

**Alternatives Considered**:

- **`discordgo`ライブラリ**: Bot用のライブラリで、今回の用途には不要な機能が多い
- **`guilds`スコープのみ**: サーバー一覧は取得できるが、メンバーシップ確認には不十分

**Implementation Notes**:

```go
// Discord API: メンバーシップ確認
// GET /users/@me/guilds/{guild_id}/member
// Header: Authorization: Bearer {oauth2_token}

func (s *MembershipService) CheckMembership(ctx context.Context, token, guildID string) (bool, error) {
    url := fmt.Sprintf("https://discord.com/api/users/@me/guilds/%s/member", guildID)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    // 200 OK = メンバー, 404 = 非メンバー
    return resp.StatusCode == 200, nil
}
```

---

## R5: セッション管理戦略

**Decision**:

- 初期実装: SQLiteでセッション管理
- 将来: Redisへの移行を検討

**Rationale**:

- 200~500ユーザー、同時接続10~50人ならSQLiteで十分
- セッションテーブル: `sessions(id, user_id, token, expires_at)`
- 定期的な期限切れセッション削除（cron or バックグラウンドゴルーチン）

**Alternatives Considered**:

- **Redis**: より高速だが、追加の依存関係と運用コスト。将来の拡張時に検討。
- **JWT のみ（ステートレス）**: セッション無効化（ログアウト）が困難。

**Implementation Notes**:

```go
// セッション期限切れの定期削除
func (s *SessionService) StartCleanup(interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            s.repo.DeleteExpired(context.Background())
        }
    }()
}
```

---

## R6: 環境変数管理

**Decision**: `.env`ファイル + `godotenv`パッケージ（開発環境）、環境変数直接設定（本番環境）

**Rationale**:

- 開発環境では`.env`ファイルで設定を簡単に管理
- 本番環境では、環境変数を直接設定（Docker、systemdなど）
- セキュリティ: `.env`はgitignore、`.env.example`でサンプルを提供

**Alternatives Considered**:

- **設定ファイル（YAML/TOML）**: 環境変数の方がデプロイ時の柔軟性が高い

**Implementation Notes**:

```go
import "github.com/joho/godotenv"

func init() {
    // 開発環境でのみ.envを読み込み
    if os.Getenv("ENV") != "production" {
        godotenv.Load()
    }
}

type Config struct {
    DiscordClientID     string
    DiscordClientSecret string
    DiscordRedirectURI  string
    DiscordGuildID      string
    JWTSecret           string
    DatabasePath        string
    ServerPort          string
    HTTPSOnly           bool
}
```

---

## R7: HTTPS/HTTP制御

**Decision**: 環境変数`HTTPS_ONLY`で制御、開発環境ではHTTP許可、本番ではHTTPS必須

**Rationale**:

- 開発環境（localhost）では証明書設定なしでHTTP使用可能
- 本番環境ではHTTPSを強制（セキュリティ）
- ミドルウェアでHTTPS強制を実装

**Implementation Notes**:

```go
func HTTPSOnlyMiddleware(httpsOnly bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if httpsOnly && r.Header.Get("X-Forwarded-Proto") != "https" {
                http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## Summary

すべての技術的な調査が完了しました。主な決定事項：

1. **Discord OAuth2**: `golang.org/x/oauth2`
2. **JWT**: `github.com/golang-jwt/jwt/v5`
3. **データベース**: SQLite + リポジトリパターン
4. **メンバーシップ確認**: Discord REST API
5. **セッション管理**: SQLite（将来Redis検討）
6. **環境変数**: godotenv + 環境変数
7. **HTTPS制御**: 環境変数で切り替え可能

次のステップ: データモデル設計（data-model.md）
