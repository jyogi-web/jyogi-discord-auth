package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/config"
	"github.com/jyogi-web/jyogi-discord-auth/internal/handler"
	"github.com/jyogi-web/jyogi-discord-auth/internal/middleware"
	gormRepo "github.com/jyogi-web/jyogi-discord-auth/internal/repository/gorm"
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

	// データベースを初期化 (TiDB)
	db, err := gormRepo.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	// GORMのDB接続はCloseする必要がない（コネクションプールで管理される）
	// sql.DBを取得してCloseすることは可能だが、main関数の最後で強制終了されるので必須ではない

	// リポジトリを初期化 (GORM implementation)
	userRepo := gormRepo.NewUserRepository(db)
	sessionRepo := gormRepo.NewSessionRepository(db)
	clientRepo := gormRepo.NewClientRepository(db)
	authCodeRepo := gormRepo.NewAuthCodeRepository(db)
	tokenRepo := gormRepo.NewTokenRepository(db)
	profileRepo := gormRepo.NewProfileRepository(db)

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
		profileRepo,
		cfg.DiscordGuildID,
	)
	oauth2Service := service.NewOAuth2Service(
		clientRepo,
		authCodeRepo,
		tokenRepo,
		userRepo,
	)
	clientService := service.NewClientService(clientRepo)
	sessionCleanupService := service.NewSessionCleanupService(
		sessionRepo,
		1*time.Hour, // 1時間ごとにクリーンアップ
	)
	// プロフィールサービス（もしあれば）

	// ハンドラーを初期化
	authHandler := handler.NewAuthHandler(authService, cfg.CORSAllowedOrigins)
	tokenHandler := handler.NewTokenHandler(authService, cfg.JWTSecret)
	apiHandler := handler.NewAPIHandler(authService)
	oauth2Handler := handler.NewOAuth2Handler(oauth2Service, authService)
	clientHandler := handler.NewClientHandler(clientService, authService)

	// セッション認証ミドルウェア
	sessionAuthMiddleware := middleware.SessionAuth(authService)

	// HTTPルーターをセットアップ
	mux := http.NewServeMux()

	// ヘルスチェックエンドポイント
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// ホーム画面
	mux.HandleFunc("/", clientHandler.HandleIndex)

	// 認証エンドポイント
	mux.HandleFunc("/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("/auth/callback", authHandler.HandleCallback)
	mux.HandleFunc("/auth/logout", authHandler.HandleLogout)
	mux.HandleFunc("/api/me", authHandler.HandleMe)
	mux.HandleFunc("/api/members", authHandler.HandleMembers)

	// クライアント管理エンドポイント
	mux.Handle("/clients", sessionAuthMiddleware(http.HandlerFunc(clientHandler.HandleListClients))) // クライアント一覧
	mux.Handle("/clients/register", sessionAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientHandler.HandleRegisterForm(w, r)
		} else if r.Method == http.MethodPost {
			clientHandler.HandleRegisterSubmit(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	// クライアント編集・削除 (動的ルート)
	mux.Handle("/clients/", sessionAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /clients/:id/edit または /clients/:id (DELETE)
		if strings.HasSuffix(r.URL.Path, "/edit") {
			clientHandler.HandleEditClientForm(w, r)
		} else if r.Method == http.MethodPost {
			clientHandler.HandleUpdateClient(w, r)
		} else if r.Method == http.MethodDelete {
			clientHandler.HandleDeleteClient(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})))

	// トークンエンドポイント
	mux.HandleFunc("/token", tokenHandler.HandleIssueToken)
	mux.HandleFunc("/token/refresh", tokenHandler.HandleRefreshToken)

	// OAuth2エンドポイント（クライアントアプリ統合用）
	mux.HandleFunc("/oauth/authorize", oauth2Handler.HandleAuthorize)
	mux.HandleFunc("/oauth/token", oauth2Handler.HandleToken)
	mux.HandleFunc("/oauth/verify", oauth2Handler.HandleVerifyToken)
	mux.HandleFunc("/oauth/userinfo", oauth2Handler.HandleUserInfo)
	mux.HandleFunc("/oauth/user/{id}", oauth2Handler.HandleUserByID)

	// JWT認証が必要なAPIエンドポイント
	jwtAuthMiddleware := middleware.JWTAuth(cfg.JWTSecret)
	mux.Handle("/api/verify", jwtAuthMiddleware(http.HandlerFunc(apiHandler.HandleVerify)))
	mux.Handle("/api/user", jwtAuthMiddleware(http.HandlerFunc(apiHandler.HandleUser)))
	mux.Handle("/api/user/{id}", jwtAuthMiddleware(http.HandlerFunc(apiHandler.HandleUserByID)))

	// ミドルウェアを適用
	handler := middleware.CORS(cfg.CORSAllowedOrigins)(mux)
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
