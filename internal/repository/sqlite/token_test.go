package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// setupTokenTestDB はテスト用のインメモリデータベースをセットアップします
func setupTokenTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// テーブルを作成
	schema := `
	CREATE TABLE tokens (
		id TEXT PRIMARY KEY,
		token TEXT NOT NULL UNIQUE,
		token_type TEXT NOT NULL,
		user_id TEXT NOT NULL,
		client_id TEXT NOT NULL,
		expires_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		revoked INTEGER NOT NULL DEFAULT 0
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

func TestTokenRepository_Create(t *testing.T) {
	db := setupTokenTestDB(t)
	defer db.Close()

	repo := NewTokenRepository(db)
	ctx := context.Background()

	now := time.Now()
	token := &domain.Token{
		ID:        "token-1",
		Token:     "test_access_token_123",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-1",
		ClientID:  "client-1",
		ExpiresAt: now.Add(1 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
	}

	err := repo.Create(ctx, token)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// 作成されたトークンを取得して確認
	retrieved, err := repo.GetByToken(ctx, token.Token)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if retrieved.Token != token.Token {
		t.Errorf("Expected Token %s, got %s", token.Token, retrieved.Token)
	}
	if retrieved.TokenType != token.TokenType {
		t.Errorf("Expected TokenType %s, got %s", token.TokenType, retrieved.TokenType)
	}
	if retrieved.UserID != token.UserID {
		t.Errorf("Expected UserID %s, got %s", token.UserID, retrieved.UserID)
	}
	if retrieved.Revoked {
		t.Error("Expected Revoked to be false, got true")
	}
}

func TestTokenRepository_GetByToken(t *testing.T) {
	db := setupTokenTestDB(t)
	defer db.Close()

	repo := NewTokenRepository(db)
	ctx := context.Background()

	// テストデータを作成
	now := time.Now()
	token := &domain.Token{
		ID:        "token-2",
		Token:     "another_test_token_456",
		TokenType: domain.TokenTypeRefresh,
		UserID:    "user-2",
		ClientID:  "client-2",
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
	}

	if err := repo.Create(ctx, token); err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// トークンで取得
	retrieved, err := repo.GetByToken(ctx, token.Token)
	if err != nil {
		t.Fatalf("Failed to get token by token: %v", err)
	}

	if retrieved.Token != token.Token {
		t.Errorf("Expected Token %s, got %s", token.Token, retrieved.Token)
	}
}

func TestTokenRepository_GetByToken_NotFound(t *testing.T) {
	db := setupTokenTestDB(t)
	defer db.Close()

	repo := NewTokenRepository(db)
	ctx := context.Background()

	// 存在しないトークンで取得を試みる
	_, err := repo.GetByToken(ctx, "non_existent_token")
	if err == nil {
		t.Error("Expected error for non-existent token, got nil")
	}
}

func TestTokenRepository_GetByUserID(t *testing.T) {
	db := setupTokenTestDB(t)
	defer db.Close()

	repo := NewTokenRepository(db)
	ctx := context.Background()

	now := time.Now()
	userID := "user-3"

	// 同じユーザーに対して複数のトークンを作成
	tokens := []*domain.Token{
		{
			ID:        "token-3-1",
			Token:     "user3_access_token",
			TokenType: domain.TokenTypeAccess,
			UserID:    userID,
			ClientID:  "client-3",
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
			Revoked:   false,
		},
		{
			ID:        "token-3-2",
			Token:     "user3_refresh_token",
			TokenType: domain.TokenTypeRefresh,
			UserID:    userID,
			ClientID:  "client-3",
			ExpiresAt: now.Add(7 * 24 * time.Hour),
			CreatedAt: now,
			Revoked:   false,
		},
	}

	for _, token := range tokens {
		if err := repo.Create(ctx, token); err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}
	}

	// ユーザーIDでトークンを取得
	retrieved, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get tokens by user_id: %v", err)
	}

	if len(retrieved) != len(tokens) {
		t.Errorf("Expected %d tokens, got %d", len(tokens), len(retrieved))
	}
}

func TestTokenRepository_Revoke(t *testing.T) {
	db := setupTokenTestDB(t)
	defer db.Close()

	repo := NewTokenRepository(db)
	ctx := context.Background()

	// テストデータを作成
	now := time.Now()
	token := &domain.Token{
		ID:        "token-4",
		Token:     "revoke_test_token",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-4",
		ClientID:  "client-4",
		ExpiresAt: now.Add(1 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
	}

	if err := repo.Create(ctx, token); err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// トークンを取り消し
	err := repo.Revoke(ctx, token.Token)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// 取り消しフラグを確認
	retrieved, err := repo.GetByToken(ctx, token.Token)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if !retrieved.Revoked {
		t.Error("Expected Revoked to be true after Revoke, got false")
	}
}

func TestTokenRepository_DeleteExpired(t *testing.T) {
	db := setupTokenTestDB(t)
	defer db.Close()

	repo := NewTokenRepository(db)
	ctx := context.Background()

	now := time.Now()

	// 期限切れのトークンを作成
	expiredToken := &domain.Token{
		ID:        "token-expired",
		Token:     "expired_token",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-5",
		ClientID:  "client-5",
		ExpiresAt: now.Add(-1 * time.Hour), // 1時間前に期限切れ
		CreatedAt: now.Add(-2 * time.Hour),
		Revoked:   false,
	}

	// 有効なトークンを作成
	validToken := &domain.Token{
		ID:        "token-valid",
		Token:     "valid_token",
		TokenType: domain.TokenTypeAccess,
		UserID:    "user-6",
		ClientID:  "client-6",
		ExpiresAt: now.Add(1 * time.Hour), // まだ有効
		CreatedAt: now,
		Revoked:   false,
	}

	if err := repo.Create(ctx, expiredToken); err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}
	if err := repo.Create(ctx, validToken); err != nil {
		t.Fatalf("Failed to create valid token: %v", err)
	}

	// 期限切れのトークンを削除
	err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("Failed to delete expired tokens: %v", err)
	}

	// 期限切れのトークンが削除されたことを確認
	_, err = repo.GetByToken(ctx, expiredToken.Token)
	if err == nil {
		t.Error("Expected error when getting expired token, got nil")
	}

	// 有効なトークンは残っていることを確認
	retrieved, err := repo.GetByToken(ctx, validToken.Token)
	if err != nil {
		t.Fatalf("Expected valid token to still exist: %v", err)
	}
	if retrieved.Token != validToken.Token {
		t.Errorf("Expected Token %s, got %s", validToken.Token, retrieved.Token)
	}
}
