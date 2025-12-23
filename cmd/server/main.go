package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jyogi-web/jyogi-discord-auth/internal/config"
	"github.com/jyogi-web/jyogi-discord-auth/internal/handler"
	"github.com/jyogi-web/jyogi-discord-auth/internal/middleware"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository/sqlite"
	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
	"github.com/jyogi-web/jyogi-discord-auth/pkg/discord"
)

func main() {
	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting じょぎメンバー認証システム...")
	log.Printf("Environment: %s", cfg.Env)
	log.Printf("Server port: %s", cfg.ServerPort)
	log.Printf("HTTPS only: %v", cfg.HTTPSOnly)

	// データベースを初期化
	db, err := initDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// リポジトリを初期化
	userRepo := sqlite.NewUserRepository(db)
	sessionRepo := sqlite.NewSessionRepository(db)
	clientRepo := sqlite.NewClientRepository(db)
	authCodeRepo := sqlite.NewAuthCodeRepository(db)
	tokenRepo := sqlite.NewTokenRepository(db)

	// Discord OAuth2クライアントを初期化
	discordClient := discord.NewClient(
		cfg.DiscordClientID,
		cfg.DiscordClientSecret,
		cfg.DiscordRedirectURI,
	)

	// サービスを初期化
	authService := service.NewAuthService(
		discordClient,
		userRepo,
		sessionRepo,
		cfg.DiscordGuildID,
	)
	oauth2Service := service.NewOAuth2Service(
		clientRepo,
		authCodeRepo,
		tokenRepo,
		userRepo,
	)
	sessionCleanupService := service.NewSessionCleanupService(
		sessionRepo,
		1*time.Hour, // 1時間ごとにクリーンアップ
	)

	// ハンドラーを初期化
	authHandler := handler.NewAuthHandler(authService)
	tokenHandler := handler.NewTokenHandler(authService, cfg.JWTSecret)
	apiHandler := handler.NewAPIHandler()
	oauth2Handler := handler.NewOAuth2Handler(oauth2Service, authService)

	// HTTPルーターをセットアップ
	mux := http.NewServeMux()

	// ヘルスチェックエンドポイント
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 認証エンドポイント
	mux.HandleFunc("/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("/auth/callback", authHandler.HandleCallback)
	mux.HandleFunc("/auth/logout", authHandler.HandleLogout)
	mux.HandleFunc("/api/me", authHandler.HandleMe)

	// トークンエンドポイント
	mux.HandleFunc("/token", tokenHandler.HandleIssueToken)
	mux.HandleFunc("/token/refresh", tokenHandler.HandleRefreshToken)

	// OAuth2エンドポイント（クライアントアプリ統合用）
	mux.HandleFunc("/oauth/authorize", oauth2Handler.HandleAuthorize)
	mux.HandleFunc("/oauth/token", oauth2Handler.HandleToken)
	mux.HandleFunc("/oauth/verify", oauth2Handler.HandleVerifyToken)
	mux.HandleFunc("/oauth/userinfo", oauth2Handler.HandleUserInfo)

	// JWT認証が必要なAPIエンドポイント
	jwtAuthMiddleware := middleware.JWTAuth(cfg.JWTSecret)
	mux.Handle("/api/verify", jwtAuthMiddleware(http.HandlerFunc(apiHandler.HandleVerify)))
	mux.Handle("/api/user", jwtAuthMiddleware(http.HandlerFunc(apiHandler.HandleUser)))

	// ミドルウェアを適用
	handler := middleware.CORS(mux)
	handler = middleware.Logging(handler)
	handler = middleware.HTTPSOnly(cfg.HTTPSOnly)(handler)

	// HTTPサーバーを作成
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// セッションクリーンアップ用のコンテキスト
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()

	// バックグラウンドでセッションクリーンアップを開始
	go sessionCleanupService.Start(cleanupCtx)

	// ゴルーチンでサーバーを起動
	go func() {
		log.Printf("Server listening on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// グレースフルシャットダウン
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// initDatabase はSQLiteデータベース接続を初期化します
func initDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 接続をテスト
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// コネクションプール設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Printf("Database initialized: %s", dbPath)

	return db, nil
}
