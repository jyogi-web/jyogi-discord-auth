package service

import (
	"context"
	"log"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

// SessionCleanupService はセッションクリーンアップサービスを表します
type SessionCleanupService struct {
	sessionRepo repository.SessionRepository
	interval    time.Duration
}

// NewSessionCleanupService は新しいセッションクリーンアップサービスを作成します
// interval: クリーンアップ実行間隔（推奨: 1時間）
func NewSessionCleanupService(sessionRepo repository.SessionRepository, interval time.Duration) *SessionCleanupService {
	return &SessionCleanupService{
		sessionRepo: sessionRepo,
		interval:    interval,
	}
}

// Start はバックグラウンドでセッションクリーンアップを開始します
// ctxがキャンセルされるとゴルーチンは終了します（グレースフルシャットダウン対応）
func (s *SessionCleanupService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Printf("Session cleanup service started (interval: %v)", s.interval)

	// 起動時に一度実行
	if err := s.cleanup(ctx); err != nil {
		log.Printf("Initial session cleanup failed: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.cleanup(ctx); err != nil {
				log.Printf("Session cleanup failed: %v", err)
			}
		case <-ctx.Done():
			log.Println("Session cleanup service stopped")
			return
		}
	}
}

// cleanup は期限切れセッションを削除します
func (s *SessionCleanupService) cleanup(ctx context.Context) error {
	log.Println("Running session cleanup...")

	if err := s.sessionRepo.DeleteExpired(ctx); err != nil {
		return err
	}

	return nil
}
