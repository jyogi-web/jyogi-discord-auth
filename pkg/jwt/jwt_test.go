package jwt

import (
	"testing"
	"time"
)

// TestGenerateToken はJWTトークン生成をテストします
func TestGenerateToken(t *testing.T) {
	secret := "test_secret_key_minimum_32_characters_long"
	userID := "test-user-id-123"
	discordID := "123456789"
	username := "testuser"

	// JWTトークンを生成
	token, err := GenerateToken(userID, discordID, username, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// トークンが空でないことを確認
	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// トークンが3つのパートに分かれていることを確認（header.payload.signature）
	parts := countTokenParts(token)
	if parts != 3 {
		t.Errorf("Expected token to have 3 parts, got %d", parts)
	}
}

// TestValidateToken はJWTトークン検証をテストします
func TestValidateToken(t *testing.T) {
	secret := "test_secret_key_minimum_32_characters_long"
	userID := "test-user-id-123"
	discordID := "123456789"
	username := "testuser"

	// 有効なトークンを生成
	token, err := GenerateToken(userID, discordID, username, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// トークンを検証
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// クレームの内容を確認
	if claims.UserID != userID {
		t.Errorf("Expected user_id %s, got %s", userID, claims.UserID)
	}

	if claims.DiscordID != discordID {
		t.Errorf("Expected discord_id %s, got %s", discordID, claims.DiscordID)
	}

	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}

	// 有効期限が未来であることを確認
	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not be expired")
	}
}

// TestValidateTokenWithInvalidSecret は無効なシークレットでの検証をテストします
func TestValidateTokenWithInvalidSecret(t *testing.T) {
	secret := "test_secret_key_minimum_32_characters_long"
	wrongSecret := "wrong_secret_key_minimum_32_characters"
	userID := "test-user-id-123"
	discordID := "123456789"
	username := "testuser"

	// トークンを生成
	token, err := GenerateToken(userID, discordID, username, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// 誤ったシークレットで検証
	_, err = ValidateToken(token, wrongSecret)
	if err == nil {
		t.Error("Expected validation to fail with wrong secret, but it succeeded")
	}
}

// TestValidateExpiredToken は期限切れトークンの検証をテストします
func TestValidateExpiredToken(t *testing.T) {
	secret := "test_secret_key_minimum_32_characters_long"
	userID := "test-user-id-123"
	discordID := "123456789"
	username := "testuser"

	// 既に期限切れのトークンを生成（-1時間）
	token, err := GenerateToken(userID, discordID, username, secret, -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// 期限切れトークンを検証
	_, err = ValidateToken(token, secret)
	if err == nil {
		t.Error("Expected validation to fail for expired token, but it succeeded")
	}
}

// TestValidateInvalidToken は無効なトークン形式の検証をテストします
func TestValidateInvalidToken(t *testing.T) {
	secret := "test_secret_key_minimum_32_characters_long"

	testCases := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"random string", "not.a.valid.token"},
		{"incomplete token", "header.payload"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateToken(tc.token, secret)
			if err == nil {
				t.Errorf("Expected validation to fail for %s, but it succeeded", tc.name)
			}
		})
	}
}

// countTokenParts はトークンのパート数を数えます（helper function）
func countTokenParts(token string) int {
	parts := 0
	for _, char := range token {
		if char == '.' {
			parts++
		}
	}
	return parts + 1
}
