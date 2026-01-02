package gorm

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// setupTestDB はテスト用のインメモリGORMデータベースをセットアップします
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// AutoMigrateでテーブルを作成
	if err := db.AutoMigrate(&Session{}); err != nil {
		t.Fatalf("Failed to migrate test schema: %v", err)
	}

	return db
}

// TestSessionRepository_Create はセッション作成機能をテストします
func TestSessionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	session := &domain.Session{
		ID:        "test-session-1",
		UserID:    "test-user-1",
		Token:     "test-token-1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 作成されたセッションを取得
	retrieved, err := repo.GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get created session: %v", err)
	}

	if retrieved.Token != session.Token {
		t.Errorf("Expected token %s, got %s", session.Token, retrieved.Token)
	}
}

// TestSessionRepository_GetByToken_Valid は有効なトークンでセッション取得をテストします
func TestSessionRepository_GetByToken_Valid(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	session := &domain.Session{
		ID:        "test-session-2",
		UserID:    "test-user-2",
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1時間後に期限切れ
		CreatedAt: time.Now(),
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 有効なトークンで取得
	retrieved, err := repo.GetByToken(ctx, "valid-token")
	if err != nil {
		t.Fatalf("Failed to get session by token: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected session, got nil")
	}

	if retrieved.Token != "valid-token" {
		t.Errorf("Expected token %s, got %s", "valid-token", retrieved.Token)
	}
}

// TestSessionRepository_GetByToken_Expired は期限切れトークンでセッション取得をテストします
// Critical Issue #1: 期限切れセッションのフィルタリングを確認
func TestSessionRepository_GetByToken_Expired(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	// 期限切れセッションを作成
	expiredSession := &domain.Session{
		ID:        "test-session-expired",
		UserID:    "test-user-3",
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1時間前に期限切れ
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	if err := repo.Create(ctx, expiredSession); err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// 期限切れトークンで取得を試みる
	retrieved, err := repo.GetByToken(ctx, "expired-token")

	// Critical Issue #2: SQLite実装との互換性（nil, nilを返すべき）
	if err != nil {
		t.Errorf("Expected nil error for expired session, got: %v", err)
	}

	if retrieved != nil {
		t.Errorf("Expected nil session for expired token, got: %+v", retrieved)
	}
}

// TestSessionRepository_GetByToken_NotFound は存在しないトークンでセッション取得をテストします
func TestSessionRepository_GetByToken_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	// 存在しないトークンで取得
	retrieved, err := repo.GetByToken(ctx, "non-existent-token")

	// SQLite実装との互換性（nil, nilを返すべき）
	if err != nil {
		t.Errorf("Expected nil error for non-existent token, got: %v", err)
	}

	if retrieved != nil {
		t.Errorf("Expected nil session for non-existent token, got: %+v", retrieved)
	}
}

// TestSessionRepository_DeleteByToken はトークンによるセッション削除をテストします
func TestSessionRepository_DeleteByToken(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	session := &domain.Session{
		ID:        "test-session-delete",
		UserID:    "test-user-4",
		Token:     "delete-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// トークンで削除
	if err := repo.DeleteByToken(ctx, "delete-token"); err != nil {
		t.Fatalf("Failed to delete session by token: %v", err)
	}

	// 削除されたことを確認
	retrieved, err := repo.GetByToken(ctx, "delete-token")
	if err != nil {
		t.Errorf("Expected nil error after deletion, got: %v", err)
	}

	if retrieved != nil {
		t.Errorf("Expected nil session after deletion, got: %+v", retrieved)
	}
}

// TestSessionRepository_DeleteByToken_NotFound は存在しないトークンの削除をテストします
// Critical Issue #3: SQLite実装との互換性（エラーを返さないべき）
func TestSessionRepository_DeleteByToken_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	// 存在しないトークンで削除
	err := repo.DeleteByToken(ctx, "non-existent-token")

	// SQLite実装との互換性（エラーを返さない）
	if err != nil {
		t.Errorf("Expected no error for deleting non-existent token, got: %v", err)
	}
}

// TestSessionRepository_DeleteExpired は期限切れセッションの削除をテストします
func TestSessionRepository_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	// 有効なセッションを作成
	validSession := &domain.Session{
		ID:        "valid-session",
		UserID:    "user-1",
		Token:     "valid-token-1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	// 期限切れセッションを作成
	expiredSession1 := &domain.Session{
		ID:        "expired-session-1",
		UserID:    "user-2",
		Token:     "expired-token-1",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	expiredSession2 := &domain.Session{
		ID:        "expired-session-2",
		UserID:    "user-3",
		Token:     "expired-token-2",
		ExpiresAt: time.Now().Add(-2 * time.Hour),
		CreatedAt: time.Now().Add(-3 * time.Hour),
	}

	// すべてのセッションを作成
	for _, s := range []*domain.Session{validSession, expiredSession1, expiredSession2} {
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	// 期限切れセッションを削除
	if err := repo.DeleteExpired(ctx); err != nil {
		t.Fatalf("Failed to delete expired sessions: %v", err)
	}

	// 有効なセッションは残っているはず
	valid, err := repo.GetByToken(ctx, "valid-token-1")
	if err != nil {
		t.Errorf("Expected no error for valid session, got: %v", err)
	}
	if valid == nil {
		t.Error("Expected valid session to remain after DeleteExpired")
	}

	// 期限切れセッションは削除されているはず
	expired1, err := repo.GetByToken(ctx, "expired-token-1")
	if err != nil {
		t.Errorf("Expected no error for expired session lookup, got: %v", err)
	}
	if expired1 != nil {
		t.Error("Expected expired session 1 to be deleted")
	}

	expired2, err := repo.GetByToken(ctx, "expired-token-2")
	if err != nil {
		t.Errorf("Expected no error for expired session lookup, got: %v", err)
	}
	if expired2 != nil {
		t.Error("Expected expired session 2 to be deleted")
	}
}
