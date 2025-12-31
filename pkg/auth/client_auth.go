package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost はbcryptのコスト（デフォルト: 10）
	BcryptCost = 10
)

// HashClientSecret はクライアントシークレットをbcryptでハッシュ化します
func HashClientSecret(secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(secret), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash secret: %w", err)
	}

	return string(hashedBytes), nil
}

// ValidateClientSecret はクライアントシークレットがハッシュと一致するか検証します
func ValidateClientSecret(secret, hashedSecret string) error {
	if secret == "" {
		return fmt.Errorf("secret cannot be empty")
	}
	if hashedSecret == "" {
		return fmt.Errorf("hashed secret cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(secret))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("invalid client secret")
		}
		return fmt.Errorf("failed to validate secret: %w", err)
	}

	return nil
}
