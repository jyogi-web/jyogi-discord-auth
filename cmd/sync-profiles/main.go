package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jyogi-web/jyogi-discord-auth/internal/config"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository/sqlite"
	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
)

func main() {
	// コマンドラインフラグを定義
	once := flag.Bool("once", false, "Run profile sync once and exit")
	intervalMinutes := flag.Int("interval", 60, "Sync interval in minutes (default: 60)")
	flag.Parse()

	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Bot TokenとChannel IDが設定されているか確認
	if cfg.DiscordBotToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN is required")
	}
	if cfg.DiscordProfileChannel == "" {
		log.Fatal("DISCORD_PROFILE_CHANNEL is required")
	}

	// データベースに接続
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// リポジトリを作成
	profileRepo := sqlite.NewProfileRepository(db)
	userRepo := sqlite.NewUserRepository(db)

	// プロフィールサービスを作成
	profileService := service.NewProfileService(
		profileRepo,
		userRepo,
		cfg.DiscordBotToken,
		cfg.DiscordProfileChannel,
	)

	ctx := context.Background()

	if *once {
		// 1回だけ実行
		log.Println("Running profile sync once...")
		if err := profileService.SyncProfiles(ctx); err != nil {
			log.Fatalf("Profile sync failed: %v", err)
		}
		log.Println("Profile sync completed successfully")
		return
	}

	// 定期実行モード
	interval := time.Duration(*intervalMinutes) * time.Minute
	scheduler := service.NewScheduler(profileService, interval)

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// スケジューラーをゴルーチンで起動
	go scheduler.Start(ctx)

	// シグナルを待つ
	sig := <-sigChan
	fmt.Printf("\nReceived signal: %v\n", sig)

	// グレースフルシャットダウン
	log.Println("Shutting down...")
	scheduler.Stop()
	time.Sleep(1 * time.Second)
	log.Println("Shutdown complete")
}
