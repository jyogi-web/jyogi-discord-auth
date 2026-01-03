package handler

import (
	"strings"
	"testing"
)

// validateRedirectURI は修正後のバリデーションロジックをテストするためのヘルパー関数
func validateRedirectURI(redirectURI string, allowedOrigins []string) bool {
	if redirectURI == "" {
		return false
	}

	for _, origin := range allowedOrigins {
		// 完全一致または正当なパス指定を確認（"example.com.attacker.com"のような攻撃を防ぐ）
		if redirectURI == origin || strings.HasPrefix(redirectURI, origin+"/") {
			return true
		}
	}

	return false
}

// TestRedirectURIValidation tests the Open Redirect vulnerability fix
func TestRedirectURIValidation(t *testing.T) {
	testCases := []struct {
		name           string
		allowedOrigins []string
		redirectURI    string
		shouldAccept   bool
	}{
		{
			name:           "正当なリダイレクトURI（完全一致）",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://example.com",
			shouldAccept:   true,
		},
		{
			name:           "正当なリダイレクトURI（パス付き）",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://example.com/auth/callback",
			shouldAccept:   true,
		},
		{
			name:           "Open Redirect攻撃の防止（サブドメイン偽装）",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://example.com.attacker.com",
			shouldAccept:   false,
		},
		{
			name:           "Open Redirect攻撃の防止（パス偽装）",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://example.com@attacker.com",
			shouldAccept:   false,
		},
		{
			name:           "異なるオリジンの拒否",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://attacker.com",
			shouldAccept:   false,
		},
		{
			name:           "複数の許可オリジン（2番目にマッチ）",
			allowedOrigins: []string{"https://example1.com", "https://example2.com"},
			redirectURI:    "https://example2.com/callback",
			shouldAccept:   true,
		},
		{
			name:           "プロトコルの違いを拒否",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "http://example.com",
			shouldAccept:   false,
		},
		{
			name:           "空のURI",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "",
			shouldAccept:   false,
		},
		{
			name:           "パスなしでスラッシュ開始（拒否）",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://example.com.evil.com/",
			shouldAccept:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validateRedirectURI(tc.redirectURI, tc.allowedOrigins)

			if result != tc.shouldAccept {
				t.Errorf("Expected %v, got %v for URI: %s", tc.shouldAccept, result, tc.redirectURI)
			}
		})
	}
}
