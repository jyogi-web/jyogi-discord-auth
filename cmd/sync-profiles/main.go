package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/config"
	gormRepo "github.com/jyogi-web/jyogi-discord-auth/internal/repository/gorm"
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

	// データベースに接続 (TiDB)
	db, err := gormRepo.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// リポジトリを作成 (GORM)
	profileRepo := gormRepo.NewProfileRepository(db)
	userRepo := gormRepo.NewUserRepository(db)

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
