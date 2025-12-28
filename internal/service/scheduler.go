package service

import (
	"context"
	"log"
	"time"
)

// Scheduler は定期実行タスクを管理します
type Scheduler struct {
	profileService *ProfileService
	interval       time.Duration
	stopChan       chan struct{}
}

// NewScheduler は新しいSchedulerを作成します
func NewScheduler(profileService *ProfileService, interval time.Duration) *Scheduler {
	return &Scheduler{
		profileService: profileService,
		interval:       interval,
		stopChan:       make(chan struct{}),
	}
}

// Start はスケジューラーを起動します
func (s *Scheduler) Start(ctx context.Context) {
	log.Printf("Starting profile sync scheduler (interval: %v)", s.interval)

	// 起動時に即座に1回実行
	if err := s.profileService.SyncProfiles(ctx); err != nil {
		log.Printf("Error in initial profile sync: %v", err)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Running scheduled profile sync...")
			if err := s.profileService.SyncProfiles(ctx); err != nil {
				log.Printf("Error in scheduled profile sync: %v", err)
			}
		case <-s.stopChan:
			log.Println("Stopping profile sync scheduler")
			return
		case <-ctx.Done():
			log.Println("Context cancelled, stopping profile sync scheduler")
			return
		}
	}
}

// Stop はスケジューラーを停止します
func (s *Scheduler) Stop() {
	close(s.stopChan)
}
