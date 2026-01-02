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

// setupTokenTestDB はトークンテスト用のインメモリGORMデータベースをセットアップします
func setupTokenTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&Token{}); err != nil {
		t.Fatalf("Failed to migrate test schema: %v", err)
	}

	return db
}

// TestTokenRepository_Create はトークン作成機能をテストします
func TestTokenRepository_Create(t *testing.T) {
	db := setupTokenTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	token := &domain.Token{
		ID:        "test-token-1",
		Token:     "access-token-1",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-1",
		ClientID:  "client-1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := repo.Create(ctx, token); err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// 作成されたトークンを取得
	retrieved, err := repo.GetByToken(ctx, token.Token)
	if err != nil {
		t.Fatalf("Failed to get created token: %v", err)
	}

	if retrieved.TokenType != token.TokenType {
		t.Errorf("Expected token_type %v, got %v", token.TokenType, retrieved.TokenType)
	}
}

// TestTokenRepository_GetByToken はトークン取得をテストします
func TestTokenRepository_GetByToken(t *testing.T) {
	db := setupTokenTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	token := &domain.Token{
		ID:        "test-token-2",
		Token:     "access-token-2",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-2",
		ClientID:  "client-2",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := repo.Create(ctx, token); err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	retrieved, err := repo.GetByToken(ctx, "access-token-2")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected token, got nil")
	}

	if retrieved.Token != "access-token-2" {
		t.Errorf("Expected token %s, got %s", "access-token-2", retrieved.Token)
	}
}

// TestTokenRepository_Revoke はトークン取り消し機能をテストします
func TestTokenRepository_Revoke(t *testing.T) {
	db := setupTokenTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	token := &domain.Token{
		ID:        "test-token-revoke",
		Token:     "revoke-token",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-3",
		ClientID:  "client-3",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := repo.Create(ctx, token); err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// トークンを取り消し
	if err := repo.Revoke(ctx, "revoke-token"); err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// 取り消されたことを確認
	retrieved, err := repo.GetByToken(ctx, "revoke-token")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if !retrieved.Revoked {
		t.Error("Expected token to be revoked")
	}
}

// TestTokenRepository_DeleteExpired は期限切れトークンの削除をテストします
func TestTokenRepository_DeleteExpired(t *testing.T) {
	db := setupTokenTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	// 有効なトークン
	validToken := &domain.Token{
		ID:        "valid-token",
		Token:     "valid-access-token",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-4",
		ClientID:  "client-4",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	// 期限切れトークン1
	expiredToken1 := &domain.Token{
		ID:        "expired-token-1",
		Token:     "expired-access-token-1",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-5",
		ClientID:  "client-5",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
		Revoked:   false,
	}

	// 期限切れトークン2
	expiredToken2 := &domain.Token{
		ID:        "expired-token-2",
		Token:     "expired-refresh-token",
		TokenType: domain.TokenTypeRefresh,
		UserID:    "user-6",
		ClientID:  "client-6",
		ExpiresAt: time.Now().Add(-30 * time.Minute),
		CreatedAt: time.Now().Add(-1 * time.Hour),
		Revoked:   false,
	}

	// すべてのトークンを作成
	for _, token := range []*domain.Token{validToken, expiredToken1, expiredToken2} {
		if err := repo.Create(ctx, token); err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}
	}

	// 期限切れトークンを削除
	if err := repo.DeleteExpired(ctx); err != nil {
		t.Fatalf("Failed to delete expired tokens: %v", err)
	}

	// 有効なトークンは残っているはず
	valid, err := repo.GetByToken(ctx, "valid-access-token")
	if err != nil {
		t.Errorf("Expected no error for valid token, got: %v", err)
	}
	if valid == nil {
		t.Error("Expected valid token to remain after DeleteExpired")
	}

	// 期限切れトークンは削除されているはず
	expired1, err := repo.GetByToken(ctx, "expired-access-token-1")
	if err == nil && expired1 != nil {
		t.Error("Expected expired token 1 to be deleted")
	}

	expired2, err := repo.GetByToken(ctx, "expired-refresh-token")
	if err == nil && expired2 != nil {
		t.Error("Expected expired token 2 to be deleted")
	}
}

// TestTokenRepository_TokenTypes はアクセストークンとリフレッシュトークンをテストします
func TestTokenRepository_TokenTypes(t *testing.T) {
	db := setupTokenTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	accessToken := &domain.Token{
		ID:        "access-token-id",
		Token:     "access-token-value",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-7",
		ClientID:  "client-7",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	refreshToken := &domain.Token{
		ID:        "refresh-token-id",
		Token:     "refresh-token-value",
		TokenType: domain.TokenTypeRefresh,
		UserID:    "user-7",
		ClientID:  "client-7",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	// 両方のトークンを作成
	if err := repo.Create(ctx, accessToken); err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}
	if err := repo.Create(ctx, refreshToken); err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// アクセストークンを取得
	retrievedAccess, err := repo.GetByToken(ctx, "access-token-value")
	if err != nil {
		t.Fatalf("Failed to get access token: %v", err)
	}
	if retrievedAccess.TokenType != domain.TokenTypeAccess {
		t.Errorf("Expected token_type %v, got %v", domain.TokenTypeAccess, retrievedAccess.TokenType)
	}

	// リフレッシュトークンを取得
	retrievedRefresh, err := repo.GetByToken(ctx, "refresh-token-value")
	if err != nil {
		t.Fatalf("Failed to get refresh token: %v", err)
	}
	if retrievedRefresh.TokenType != domain.TokenTypeRefresh {
		t.Errorf("Expected token_type %v, got %v", domain.TokenTypeRefresh, retrievedRefresh.TokenType)
	}
}

// TestTokenRepository_UniqueConstraint はトークンの一意性制約をテストします
func TestTokenRepository_UniqueConstraint(t *testing.T) {
	db := setupTokenTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	token1 := &domain.Token{
		ID:        "token-1",
		Token:     "unique-token",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-8",
		ClientID:  "client-8",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := repo.Create(ctx, token1); err != nil {
		t.Fatalf("Failed to create first token: %v", err)
	}

	// 同じトークン値で別のトークンを作成しようとする
	token2 := &domain.Token{
		ID:        "token-2",
		Token:     "unique-token", // 重複
		TokenType: domain.TokenTypeRefresh,
		UserID:    "user-9",
		ClientID:  "client-9",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, token2)
	if err == nil {
		t.Error("Expected error when creating token with duplicate token value, got nil")
	}
}
