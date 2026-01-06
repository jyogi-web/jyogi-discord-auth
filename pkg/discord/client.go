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
	ID            string  `json:"id"`
	Username      string  `json:"username"`
	Discriminator string  `json:"discriminator"`
	GlobalName    *string `json:"global_name"` // Display name
	Avatar        *string `json:"avatar"`      // Avatar hash
}

// GetAvatarURL はアバターのURLを返します
// アバターが設定されていない場合は空文字列を返します
func (u *User) GetAvatarURL() string {
	if u.Avatar == nil || *u.Avatar == "" {
		return ""
	}
	// Discord CDNのアバターURL
	// https://cdn.discordapp.com/avatars/{user_id}/{avatar_hash}.png
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", u.ID, *u.Avatar)
}

// GetDisplayName は表示名を返します
// GlobalNameが設定されていない場合はUsernameを返します
func (u *User) GetDisplayName() string {
	if u.GlobalName != nil && *u.GlobalName != "" {
		return *u.GlobalName
	}
	return u.Username
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
	User     User     `json:"user"`
	Roles    []string `json:"roles"`
	JoinedAt string   `json:"joined_at"`
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

// Message はDiscordメッセージを表します
type Message struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	Author    User   `json:"author"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// GetChannelMessages はチャンネルのメッセージを取得します（ページネーション非対応）
// botTokenはBot認証用のトークンです
func GetChannelMessages(ctx context.Context, botToken, channelID string, limit int) ([]*Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 100 // Discord APIの上限
	}

	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages?limit=%d", channelID, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", botToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel messages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord API returned status %d: %s", resp.StatusCode, string(body))
	}

	var messages []*Message
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}

	return messages, nil
}

// GetAllChannelMessages はチャンネルのすべてのメッセージを取得します（ページネーション対応）
// maxMessagesで取得する最大メッセージ数を指定できます（0の場合は制限なし）
func GetAllChannelMessages(ctx context.Context, botToken, channelID string, maxMessages int) ([]*Message, error) {
	var allMessages []*Message
	var beforeID string
	batchSize := 100 // Discord APIの1回あたりの最大取得数

	for {
		// URLを構築
		url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages?limit=%d", channelID, batchSize)
		if beforeID != "" {
			url += fmt.Sprintf("&before=%s", beforeID)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bot %s", botToken))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get channel messages: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("discord API returned status %d: %s", resp.StatusCode, string(body))
		}

		var messages []*Message
		if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode messages: %w", err)
		}
		resp.Body.Close()

		// メッセージがない場合は終了
		if len(messages) == 0 {
			break
		}

		// メッセージを追加
		allMessages = append(allMessages, messages...)

		// 最大数に達した場合は終了
		if maxMessages > 0 && len(allMessages) >= maxMessages {
			allMessages = allMessages[:maxMessages]
			break
		}

		// 100件未満の場合は、これ以上メッセージがないので終了
		if len(messages) < batchSize {
			break
		}

		// 次のページのために最後のメッセージIDを保存
		beforeID = messages[len(messages)-1].ID
	}

	return allMessages, nil
}
