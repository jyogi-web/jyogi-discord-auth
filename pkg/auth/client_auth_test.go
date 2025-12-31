package auth

import (
	"strings"
	"testing"
)

func TestHashClientSecret(t *testing.T) {
	secret := "my_super_secret_key_123"

	hashed, err := HashClientSecret(secret)
	if err != nil {
		t.Fatalf("Failed to hash secret: %v", err)
	}

	if hashed == "" {
		t.Error("Expected non-empty hashed secret")
	}

	// ハッシュは元のシークレットと異なるべき
	if hashed == secret {
		t.Error("Hashed secret should not equal original secret")
	}

	// ハッシュはbcryptフォーマットで始まるべき
	if !strings.HasPrefix(hashed, "$2a$") && !strings.HasPrefix(hashed, "$2b$") {
		t.Errorf("Expected bcrypt hash format, got: %s", hashed)
	}
}

func TestHashClientSecret_Empty(t *testing.T) {
	_, err := HashClientSecret("")
	if err == nil {
		t.Error("Expected error for empty secret, got nil")
	}
}

func TestValidateClientSecret(t *testing.T) {
	secret := "test_secret_password"

	// シークレットをハッシュ化
	hashed, err := HashClientSecret(secret)
	if err != nil {
		t.Fatalf("Failed to hash secret: %v", err)
	}

	// 正しいシークレットで検証
	err = ValidateClientSecret(secret, hashed)
	if err != nil {
		t.Errorf("Expected validation to succeed, got error: %v", err)
	}
}

func TestValidateClientSecret_WrongSecret(t *testing.T) {
	secret := "correct_secret"
	wrongSecret := "wrong_secret"

	// シークレットをハッシュ化
	hashed, err := HashClientSecret(secret)
	if err != nil {
		t.Fatalf("Failed to hash secret: %v", err)
	}

	// 間違ったシークレットで検証
	err = ValidateClientSecret(wrongSecret, hashed)
	if err == nil {
		t.Error("Expected validation to fail for wrong secret, got nil")
	}

	// エラーメッセージを確認
	if !strings.Contains(err.Error(), "invalid client secret") {
		t.Errorf("Expected 'invalid client secret' error, got: %v", err)
	}
}

func TestValidateClientSecret_EmptySecret(t *testing.T) {
	err := ValidateClientSecret("", "some_hash")
	if err == nil {
		t.Error("Expected error for empty secret, got nil")
	}
}

func TestValidateClientSecret_EmptyHash(t *testing.T) {
	err := ValidateClientSecret("some_secret", "")
	if err == nil {
		t.Error("Expected error for empty hash, got nil")
	}
}

func TestHashClientSecret_Deterministic(t *testing.T) {
	secret := "test_deterministic_hashing"

	// 同じシークレットを2回ハッシュ化
	hash1, err := HashClientSecret(secret)
	if err != nil {
		t.Fatalf("Failed to hash secret (1): %v", err)
	}

	hash2, err := HashClientSecret(secret)
	if err != nil {
		t.Fatalf("Failed to hash secret (2): %v", err)
	}

	// bcryptはソルト付きなので、ハッシュは毎回異なるべき
	if hash1 == hash2 {
		t.Error("Expected different hashes for same secret (bcrypt uses random salt)")
	}

	// しかし、どちらのハッシュも元のシークレットで検証できるべき
	if err := ValidateClientSecret(secret, hash1); err != nil {
		t.Errorf("Failed to validate against hash1: %v", err)
	}
	if err := ValidateClientSecret(secret, hash2); err != nil {
		t.Errorf("Failed to validate against hash2: %v", err)
	}
}
