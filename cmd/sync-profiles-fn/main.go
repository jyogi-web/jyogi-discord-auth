package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jyogi-web/jyogi-discord-auth/internal/config"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository/sqlite"
	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
)

// Response はHTTPレスポンスの構造体
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Stats   *Stats `json:"stats,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Stats は同期結果の統計情報
type Stats struct {
	Success int `json:"success"`
	Skipped int `json:"skipped"`
	Errors  int `json:"errors"`
}

// グローバル変数として設定とサービスを保持（コールド起動を最小化）
var (
	profileService *service.ProfileService
	initError      error
)

// init は関数の初期化時に一度だけ実行される
func init() {
	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		initError = err
		log.Printf("Failed to load config: %v", err)
		return
	}

	// Bot TokenとChannel IDが設定されているか確認
	if cfg.DiscordBotToken == "" {
		initError = err
		log.Fatal("DISCORD_BOT_TOKEN is required")
		return
	}
	if cfg.DiscordProfileChannel == "" {
		initError = err
		log.Fatal("DISCORD_PROFILE_CHANNEL is required")
		return
	}

	// データベースに接続
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		initError = err
		log.Printf("Failed to open database: %v", err)
		return
	}

	// リポジトリを作成
	profileRepo := sqlite.NewProfileRepository(db)
	userRepo := sqlite.NewUserRepository(db)

	// プロフィールサービスを作成
	profileService = service.NewProfileService(
		profileRepo,
		userRepo,
		cfg.DiscordBotToken,
		cfg.DiscordProfileChannel,
	)

	log.Println("Profile sync function initialized successfully")
}

// SyncProfilesHandler はHTTPリクエストを処理するハンドラー
func SyncProfilesHandler(w http.ResponseWriter, r *http.Request) {
	// 初期化エラーがあればエラーレスポンスを返す
	if initError != nil {
		respondError(w, http.StatusInternalServerError, "Initialization failed", initError)
		return
	}

	// POSTまたはGETメソッドのみ許可
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	log.Println("Starting profile synchronization via HTTP...")

	// プロフィールを同期
	ctx := context.Background()
	if err := profileService.SyncProfiles(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Profile sync failed", err)
		return
	}

	// 成功レスポンスを返す
	response := Response{
		Success: true,
		Message: "Profile synchronization completed successfully",
	}

	respondJSON(w, http.StatusOK, response)
}

// respondJSON はJSON形式でレスポンスを返す
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// respondError はエラーレスポンスを返す
func respondError(w http.ResponseWriter, statusCode int, message string, err error) {
	response := Response{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
		log.Printf("Error: %s - %v", message, err)
	}

	respondJSON(w, statusCode, response)
}

// main はローカルでのテスト用
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/", SyncProfilesHandler)
	http.HandleFunc("/sync", SyncProfilesHandler)

	log.Printf("Starting profile sync function on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
