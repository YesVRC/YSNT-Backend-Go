package routes

import (
	"errors"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"go-backend-discord/modules/database"
	"net/http"
	"strings"
	"time"
)

type DiscordAuth struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type DiscordUser struct {
	Id                   string `json:"id"`
	Username             string `json:"username"`
	Avatar               string `json:"avatar"`
	Discriminator        string `json:"discriminator"`
	PublicFlags          int64  `json:"public_flags"`
	Flags                int64  `json:"flags"`
	Banner               string `json:"banner"`
	AccentColor          string `json:"accent_color"`
	GlobalName           string `json:"global_name"`
	AvatarDecorationData string `json:"avatar_decoration_data"`
	BannerColor          string `json:"banner_color"`
	ClanName             string `json:"clan"`
	PrimaryGuildId       string `json:"primary_guild"`
	MfaEnabled           bool   `json:"mfa_enabled"`
	Locale               string `json:"locale"`
	PremiumType          int    `json:"premium_type"`
	Email                string `json:"email"`
	Verified             bool   `json:"verified"`
}

type TwitchAuth struct {
	AccessToken  string   `json:"access_token"`
	ExpiresIn    uint     `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

type TwitchUser struct {
	Id              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"display_name"`
	Type            string    `json:"type"`
	BroadcasterType string    `json:"broadcaster_type"`
	Description     string    `json:"description"`
	ProfileImageUrl string    `json:"profile_image_url"`
	OfflineImageUrl string    `json:"offline_image_url"`
	ViewCount       int64     `json:"view_count"`
	Email           string    `json:"email"`
	CreatedAt       time.Time `json:"created_at"`
}

type TwitchUserData struct {
	Data []TwitchUser `json:"data"`
}

type TwitchVerify struct {
	ClientId  string   `json:"client_id"`
	Login     string   `json:"login"`
	Scopes    []string `json:"scopes"`
	UserId    string   `json:"user_id"`
	ExpiresIn int64    `json:"expires_in"`
}

var SessionExpiredError = errors.New("session expired")
var SessionInvalidError = errors.New("session invalid")

func GenerateID() (string, error) {
	sessionId, err := gonanoid.Generate("ABCDEFGHIJKLMMNOPQRSTUVWXYZ-_", 10)
	if err != nil {
		return "", err
	}
	return sessionId, nil
}

func SessionFromAuthHeader(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", SessionInvalidError
	}

	return strings.TrimPrefix(header, "Session "), nil
}
func CreateSession(user *database.User) (string, error) {
	sessionId, err := GenerateID()
	if err != nil {
		return "", err
	}

	session := &database.Session{
		User:      *user,
		SessionID: "session_" + sessionId,
	}

	res := database.Db.Create(session)
	if res.Error != nil {
		return "", res.Error
	}
	return "session_" + sessionId, nil
}

func GetSession(session string) (*database.User, error) {
	var sessionData database.Session
	res := database.Db.Where("session_id = ?", session).First(&sessionData)
	if res.Error != nil {
		return nil, res.Error
	}
	return &sessionData.User, nil
}
