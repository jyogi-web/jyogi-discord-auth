package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// ClientApp はこの認証サーバーを使用するアプリケーション（SSO用）を表します
type ClientApp struct {
	ID           string
	OwnerID      string // クライアントアプリの作成者ID
	ClientID     string
	ClientSecret string // bcryptでハッシュ化
	Name         string
	RedirectURIs []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate はクライアントアプリのデータが有効かどうかを確認します
func (c *ClientApp) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.RedirectURIs) == 0 {
		return fmt.Errorf("at least one redirect_uri is required")
	}
	return nil
}

// RedirectURIsToJSON はリダイレクトURIのスライスを保存用のJSON文字列に変換します
func (c *ClientApp) RedirectURIsToJSON() (string, error) {
	data, err := json.Marshal(c.RedirectURIs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal redirect_uris: %w", err)
	}
	return string(data), nil
}

// RedirectURIsFromJSON はJSON文字列をリダイレクトURIのスライスにパースします
func (c *ClientApp) RedirectURIsFromJSON(jsonStr string) error {
	if err := json.Unmarshal([]byte(jsonStr), &c.RedirectURIs); err != nil {
		return fmt.Errorf("failed to unmarshal redirect_uris: %w", err)
	}
	return nil
}

// IsRedirectURIValid は指定されたURIが許可リストに含まれているかを確認します
func (c *ClientApp) IsRedirectURIValid(uri string) bool {
	for _, allowed := range c.RedirectURIs {
		if allowed == uri {
			return true
		}
	}
	return false
}
