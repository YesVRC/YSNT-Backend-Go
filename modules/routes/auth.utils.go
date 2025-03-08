package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"go-backend-discord/modules/database"
	"io"
	"net/http"
	"os"
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

func GetDiscordOAuth(code string) (*DiscordAuth, error) {
	httpClient := &http.Client{}
	body := strings.NewReader(fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", code, os.Getenv("DISCORD_AUTH_REDIRECT_URL")))
	req, err := http.NewRequest(http.MethodPost, "https://discord.com/api/v10/oauth2/token", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(os.Getenv("DISCORD_CLIENT_ID"), os.Getenv("DISCORD_CLIENT_SECRET"))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord OAuth response code %d: %s", resp.StatusCode, string(respBody))
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	data, _ := io.ReadAll(resp.Body)
	if os.Getenv("DEBUG") == "true" {
		println("GetDiscordOAuth: " + string(data))
	}

	out := &DiscordAuth{}
	err = json.Unmarshal(data, &out)

	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetDiscordAccountFromToken(token string) (*DiscordUser, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(httpResp.Body)
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	out := &DiscordUser{}
	err = json.Unmarshal(body, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetTwitchOAuth(code string) (*TwitchAuth, error) {
	httpClient := &http.Client{}
	body := strings.NewReader(fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s&client_id=%s&client_secret=%s",
		code,
		os.Getenv("TWITCH_AUTH_REDIRECT_URL"),
		os.Getenv("TWITCH_CLIENT_ID"),
		os.Getenv("TWITCH_CLIENT_SECRET")))

	req, err := http.NewRequest(http.MethodPost, "https://id.twitch.tv/oauth2/token", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	data, _ := io.ReadAll(resp.Body)
	out := &TwitchAuth{}
	err = json.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetTwitchAccountFromToken(token string) (*TwitchUser, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Client-ID", os.Getenv("TWITCH_CLIENT_ID"))
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf(httpResp.Status)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(httpResp.Body)
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	out := &TwitchUserData{}
	err = json.Unmarshal(body, &out)
	if err != nil {
		return nil, err
	}
	return &out.Data[0], nil
}
