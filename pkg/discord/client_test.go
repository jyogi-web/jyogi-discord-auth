package discord

import (
	"testing"
)

// TestNewClient はDiscord OAuth2クライアントが正しく初期化されることを確認します
func TestNewClient(t *testing.T) {
	clientID := "test_client_id"
	clientSecret := "test_client_secret"
	redirectURI := "http://localhost:8080/auth/callback"

	client := NewClient(clientID, clientSecret, redirectURI)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	// OAuth2設定が正しく設定されているか確認
	if client.config.ClientID != clientID {
		t.Errorf("ClientID = %v, want %v", client.config.ClientID, clientID)
	}

	if client.config.ClientSecret != clientSecret {
		t.Errorf("ClientSecret = %v, want %v", client.config.ClientSecret, clientSecret)
	}

	if client.config.RedirectURL != redirectURI {
		t.Errorf("RedirectURL = %v, want %v", client.config.RedirectURL, redirectURI)
	}

	// Scopesが正しく設定されているか確認
	expectedScopes := []string{"identify", "guilds.members.read"}
	if len(client.config.Scopes) != len(expectedScopes) {
		t.Fatalf("Scopes length = %v, want %v", len(client.config.Scopes), len(expectedScopes))
	}

	for i, scope := range expectedScopes {
		if client.config.Scopes[i] != scope {
			t.Errorf("Scope[%d] = %v, want %v", i, client.config.Scopes[i], scope)
		}
	}

	// Discord APIエンドポイントが正しく設定されているか確認
	expectedAuthURL := "https://discord.com/api/oauth2/authorize"
	expectedTokenURL := "https://discord.com/api/oauth2/token"

	if client.config.Endpoint.AuthURL != expectedAuthURL {
		t.Errorf("AuthURL = %v, want %v", client.config.Endpoint.AuthURL, expectedAuthURL)
	}

	if client.config.Endpoint.TokenURL != expectedTokenURL {
		t.Errorf("TokenURL = %v, want %v", client.config.Endpoint.TokenURL, expectedTokenURL)
	}
}

// TestGetAuthURL は認証URLが正しく生成されることを確認します
func TestGetAuthURL(t *testing.T) {
	client := NewClient("test_client_id", "test_client_secret", "http://localhost:8080/auth/callback")

	state := "test_state_123"
	authURL := client.GetAuthURL(state)

	if authURL == "" {
		t.Fatal("GetAuthURL returned empty string")
	}

	// URLに必要なパラメータが含まれているか確認
	expectedParams := []string{
		"client_id=test_client_id",
		"redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fauth%2Fcallback",
		"response_type=code",
		"scope=identify+guilds.members.read",
		"state=test_state_123",
	}

	for _, param := range expectedParams {
		if !contains(authURL, param) {
			t.Errorf("AuthURL does not contain expected param: %v", param)
		}
	}
}

// contains はシンプルな文字列検索ヘルパー
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
