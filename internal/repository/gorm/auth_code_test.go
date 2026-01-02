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

// setupAuthCodeTestDB は認可コードテスト用のインメモリGORMデータベースをセットアップします
func setupAuthCodeTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&AuthCode{}); err != nil {
		t.Fatalf("Failed to migrate test schema: %v", err)
	}

	return db
}

// TestAuthCodeRepository_Create は認可コード作成機能をテストします
func TestAuthCodeRepository_Create(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	authCode := &domain.AuthCode{
		Code:        "test-code-1",
		ClientID:    "client-1",
		UserID:      "user-1",
		RedirectURI: "https://example.com/callback",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}

	if err := repo.Create(ctx, authCode); err != nil {
		t.Fatalf("Failed to create auth code: %v", err)
	}

	// 作成された認可コードを取得
	retrieved, err := repo.GetByCode(ctx, authCode.Code)
	if err != nil {
		t.Fatalf("Failed to get created auth code: %v", err)
	}

	if retrieved.ClientID != authCode.ClientID {
		t.Errorf("Expected client_id %s, got %s", authCode.ClientID, retrieved.ClientID)
	}
}

// TestAuthCodeRepository_GetByCode は認可コード取得をテストします
func TestAuthCodeRepository_GetByCode(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	authCode := &domain.AuthCode{
		Code:        "test-code-2",
		ClientID:    "client-2",
		UserID:      "user-2",
		RedirectURI: "https://example.com/callback",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}

	if err := repo.Create(ctx, authCode); err != nil {
		t.Fatalf("Failed to create auth code: %v", err)
	}

	retrieved, err := repo.GetByCode(ctx, "test-code-2")
	if err != nil {
		t.Fatalf("Failed to get auth code: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected auth code, got nil")
	}

	if retrieved.Code != "test-code-2" {
		t.Errorf("Expected code %s, got %s", "test-code-2", retrieved.Code)
	}
}

// TestAuthCodeRepository_MarkAsUsed は認可コードの使用済みマークをテストします
func TestAuthCodeRepository_MarkAsUsed(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	authCode := &domain.AuthCode{
		Code:        "test-code-used",
		ClientID:    "client-3",
		UserID:      "user-3",
		RedirectURI: "https://example.com/callback",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
		Used:        false,
	}

	if err := repo.Create(ctx, authCode); err != nil {
		t.Fatalf("Failed to create auth code: %v", err)
	}

	// 使用済みとしてマーク
	if err := repo.MarkAsUsed(ctx, "test-code-used"); err != nil {
		t.Fatalf("Failed to mark auth code as used: %v", err)
	}

	// 使用済みになっていることを確認
	retrieved, err := repo.GetByCode(ctx, "test-code-used")
	if err != nil {
		t.Fatalf("Failed to get auth code: %v", err)
	}

	if !retrieved.Used {
		t.Error("Expected auth code to be marked as used")
	}
}

// TestAuthCodeRepository_DeleteExpired は期限切れ認可コードの削除をテストします
func TestAuthCodeRepository_DeleteExpired(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	// 有効な認可コード
	validCode := &domain.AuthCode{
		ID:          "valid-code-id",
		Code:        "valid-code",
		ClientID:    "client-4",
		UserID:      "user-4",
		RedirectURI: "https://example.com/callback",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}

	// 期限切れ認可コード
	expiredCode1 := &domain.AuthCode{
		ID:          "expired-code-1-id",
		Code:        "expired-code-1",
		ClientID:    "client-5",
		UserID:      "user-5",
		RedirectURI: "https://example.com/callback",
		ExpiresAt:   time.Now().Add(-10 * time.Minute),
		CreatedAt:   time.Now().Add(-20 * time.Minute),
	}

	expiredCode2 := &domain.AuthCode{
		ID:          "expired-code-2-id",
		Code:        "expired-code-2",
		ClientID:    "client-6",
		UserID:      "user-6",
		RedirectURI: "https://example.com/callback",
		ExpiresAt:   time.Now().Add(-5 * time.Minute),
		CreatedAt:   time.Now().Add(-15 * time.Minute),
	}

	// すべての認可コードを作成
	for _, code := range []*domain.AuthCode{validCode, expiredCode1, expiredCode2} {
		if err := repo.Create(ctx, code); err != nil {
			t.Fatalf("Failed to create auth code: %v", err)
		}
	}

	// 期限切れ認可コードを削除
	if err := repo.DeleteExpired(ctx); err != nil {
		t.Fatalf("Failed to delete expired auth codes: %v", err)
	}

	// 有効な認可コードは残っているはず
	valid, err := repo.GetByCode(ctx, "valid-code")
	if err != nil {
		t.Errorf("Expected no error for valid code, got: %v", err)
	}
	if valid == nil {
		t.Error("Expected valid code to remain after DeleteExpired")
	}

	// 期限切れ認可コードは削除されているはず
	expired1, err := repo.GetByCode(ctx, "expired-code-1")
	if err == nil && expired1 != nil {
		t.Error("Expected expired code 1 to be deleted")
	}

	expired2, err := repo.GetByCode(ctx, "expired-code-2")
	if err == nil && expired2 != nil {
		t.Error("Expected expired code 2 to be deleted")
	}
}

// TestAuthCodeRepository_UniqueConstraint は認可コードの一意性制約をテストします
func TestAuthCodeRepository_UniqueConstraint(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	code1 := &domain.AuthCode{
		Code:        "unique-code",
		ClientID:    "client-8",
		UserID:      "user-8",
		RedirectURI: "https://example.com/callback",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}

	if err := repo.Create(ctx, code1); err != nil {
		t.Fatalf("Failed to create first auth code: %v", err)
	}

	// 同じコードで別の認可コードを作成しようとする
	code2 := &domain.AuthCode{
		Code:        "unique-code", // 重複
		ClientID:    "client-9",
		UserID:      "user-9",
		RedirectURI: "https://example.com/callback",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, code2)
	if err == nil {
		t.Error("Expected error when creating auth code with duplicate code, got nil")
	}
}
