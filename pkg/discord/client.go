package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// Client はDiscord OAuth2クライアントを表します
type Client struct {
	config *oauth2.Config
}

// DiscordエンドポイントのURL
var discordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

// NewClient は新しいDiscord OAuth2クライアントを作成します
func NewClient(clientID, clientSecret, redirectURI string) *Client {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"identify", "guilds.members.read"},
		Endpoint:     discordEndpoint,
	}

	return &Client{
		config: config,
	}
}

// GetAuthURL は認証URLを生成します
// stateパラメータはCSRF攻撃を防ぐために使用されます
func (c *Client) GetAuthURL(state string) string {
	return c.config.AuthCodeURL(state)
}

// ExchangeCode は認可コードをアクセストークンに交換します
func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

// User はDiscordユーザー情報を表します
type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

// GetUser はアクセストークンを使用してユーザー情報を取得します
func (c *Client) GetUser(ctx context.Context, token *oauth2.Token) (*User, error) {
	client := c.config.Client(ctx, token)

	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord API returned status %d: %s", resp.StatusCode, string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &user, nil
}

// GuildMember はギルドメンバー情報を表します
type GuildMember struct {
	User   User     `json:"user"`
	Roles  []string `json:"roles"`
	JoinedAt string `json:"joined_at"`
}

// IsMemberOfGuild はユーザーが指定されたギルドのメンバーかどうかを確認します
func (c *Client) IsMemberOfGuild(ctx context.Context, token *oauth2.Token, guildID, userID string) (bool, error) {
	client := c.config.Client(ctx, token)

	url := fmt.Sprintf("https://discord.com/api/users/@me/guilds/%s/member", guildID)
	resp, err := client.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to check guild membership: %w", err)
	}
	defer resp.Body.Close()

	// 404はメンバーではないことを示す
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("discord API returned status %d: %s", resp.StatusCode, string(body))
	}

	return true, nil
}
