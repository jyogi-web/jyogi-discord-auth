package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims はJWTクレームを表します
type Claims struct {
	UserID    string `json:"user_id"`
	DiscordID string `json:"discord_id"`
	Username  string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken はJWTトークンを生成します
// userID: ユーザーID
// discordID: Discord ID
// username: ユーザー名
// secret: JWT署名用のシークレットキー
// duration: トークンの有効期間
func GenerateToken(userID, discordID, username, secret string, duration time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	// クレームを作成
	claims := &Claims{
		UserID:    userID,
		DiscordID: discordID,
		Username:  username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "jyogi-auth",
		},
	}

	// HS256アルゴリズムでトークンを生成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// トークンに署名
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken はJWTトークンを検証し、クレームを返します
// tokenString: 検証するJWTトークン
// secret: JWT署名検証用のシークレットキー
func ValidateToken(tokenString, secret string) (*Claims, error) {
	// トークンをパースして検証
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名アルゴリズムが期待通りか確認
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// クレームを取得
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// トークンが有効か確認
	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}
