package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
	"github.com/jyogi-web/jyogi-discord-auth/pkg/discord"
)

// AuthService は認証サービスを表します
type AuthService struct {
	discordClient *discord.Client
	userRepo      repository.UserRepository
	sessionRepo   repository.SessionRepository
	profileRepo   repository.ProfileRepository
	guildID       string
}

// NewAuthService は新しい認証サービスを作成します
func NewAuthService(
	discordClient *discord.Client,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	profileRepo repository.ProfileRepository,
	guildID string,
) *AuthService {
	return &AuthService{
		discordClient: discordClient,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		profileRepo:   profileRepo,
		guildID:       guildID,
	}
}

// GenerateState はCSRF攻撃を防ぐためのランダムなstate文字列を生成します
func (s *AuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL は認証URLを取得します
func (s *AuthService) GetAuthURL(state string) string {
	return s.discordClient.GetAuthURL(state)
}

// HandleCallback はDiscord OAuth2コールバックを処理します
// 認証成功時にセッショントークンを返します
func (s *AuthService) HandleCallback(ctx context.Context, code string) (string, error) {
	// 1. 認可コードをアクセストークンに交換
	token, err := s.discordClient.ExchangeCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// 2. ユーザー情報を取得
	discordUser, err := s.discordClient.GetUser(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	// 3. じょぎサーバーメンバーシップを確認
	isMember, err := s.discordClient.IsMemberOfGuild(ctx, token, s.guildID, discordUser.ID)
	if err != nil {
		return "", fmt.Errorf("failed to check guild membership: %w", err)
	}

	if !isMember {
		return "", domain.ErrNotGuildMember
	}

	// 4. ユーザーをデータベースに保存または更新
	user, err := s.upsertUser(ctx, discordUser, token)
	if err != nil {
		return "", fmt.Errorf("failed to upsert user: %w", err)
	}

	// 5. セッションを作成
	sessionToken, err := s.createSession(ctx, user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

// upsertUser はユーザーを作成または更新します
func (s *AuthService) upsertUser(ctx context.Context, discordUser *discord.User, token *oauth2.Token) (*domain.User, error) {
	// 既存のユーザーを検索
	existingUser, err := s.userRepo.GetByDiscordID(ctx, discordUser.ID)
	// エラーが発生した場合、ユーザーが見つからない場合以外はエラーを返す
	// 今回の変更でGetByDiscordIDはユーザーが見つからない場合にエラーを返すようになったので、
	// エラーの内容を確認する必要があるが、簡易的にエラーがあれば新規作成とみなす実装にはできない。
	// エラーの種類を判別するか、エラーメッセージで判断する必要があるが、
	// GORMの実装では fmt.Errorf("user not found: %s", discordID) を返している。
	// 厳密にはエラータイプを定義すべきだが、ここではエラーがあれば新規ユーザー作成フローに進むようにする（要検討）
	// ただし、DB接続エラーなども含まれるため、本来はエラータイプを確認すべき。
	// ここでは、"user not found" を含むエラーの場合は新規作成、それ以外はエラーとする。
	if err != nil {
		if err.Error() == fmt.Sprintf("user not found: %s", discordUser.ID) {
			existingUser = nil
		} else {
			return nil, fmt.Errorf("failed to get user by discord_id: %w", err)
		}
	}

	now := time.Now()
	avatarURL := ""
	if discordUser.Avatar != "" {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUser.ID, discordUser.Avatar)
	}

	if existingUser != nil {
		// 既存ユーザーを更新
		existingUser.Username = discordUser.Username
		existingUser.AvatarURL = avatarURL
		existingUser.LastLoginAt = &now
		existingUser.UpdatedAt = now

		if err := s.userRepo.Update(ctx, existingUser); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		return existingUser, nil
	}

	// 新規ユーザーを作成
	user := &domain.User{
		ID:          uuid.New().String(),
		DiscordID:   discordUser.ID,
		Username:    discordUser.Username,
		AvatarURL:   avatarURL,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: &now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// createSession はセッションを作成します
func (s *AuthService) createSession(ctx context.Context, userID string) (string, error) {
	sessionToken, err := s.GenerateState() // ランダムなトークンを生成
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	session := &domain.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7日間有効
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

// GetUserBySessionToken はセッショントークンからユーザーを取得します
func (s *AuthService) GetUserBySessionToken(ctx context.Context, sessionToken string) (*domain.User, error) {
	// セッションを取得
	session, err := s.sessionRepo.GetByToken(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// session is already checked for error above, if GetByToken returns error for not found,
	// err check handles it. If it returns valid session, session is not nil.
	// So we don't need explicit nil check if we trust GetByToken to return error on not found.
	// But let's keep it safe.
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// セッションが期限切れかチェック
	if session.IsExpired() {
		return nil, fmt.Errorf("session expired")
	}

	// ユーザーを取得
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Logout はセッションを削除してログアウトします
func (s *AuthService) Logout(ctx context.Context, sessionToken string) error {
	if err := s.sessionRepo.DeleteByToken(ctx, sessionToken); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// MemberWithProfile はメンバー情報とプロフィール情報を結合した構造体
type MemberWithProfile struct {
	User    *domain.User
	Profile *domain.Profile
}

// GetAllMembers は全てのメンバーを取得します（Deprecated: GetAllMembersWithProfilesを使用してください）
func (s *AuthService) GetAllMembers(ctx context.Context) ([]*domain.User, error) {
	// 全ユーザーを取得
	members, err := s.userRepo.GetAll(ctx, 500, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get all members: %w", err)
	}

	return members, nil
}

// GetAllMembersWithProfiles は全てのメンバーとそのプロフィール情報を取得します
func (s *AuthService) GetAllMembersWithProfiles(ctx context.Context) ([]*MemberWithProfile, error) {
	// 全ユーザーを取得
	users, err := s.userRepo.GetAll(ctx, 500, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	// 全プロフィールを一括取得（N+1問題を回避）
	allProfiles, err := s.profileRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all profiles: %w", err)
	}

	// プロフィールをマップ化（UserID -> Profile）
	profileMap := make(map[string]*domain.Profile, len(allProfiles))
	for _, profile := range allProfiles {
		profileMap[profile.UserID] = profile
	}

	// ユーザーとプロフィールを結合
	result := make([]*MemberWithProfile, 0, len(users))
	for _, user := range users {
		result = append(result, &MemberWithProfile{
			User:    user,
			Profile: profileMap[user.ID], // マップから取得（存在しなければnil）
		})
	}

	return result, nil
}
