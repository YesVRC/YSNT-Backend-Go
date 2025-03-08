package routes

import (
	"encoding/json"
	"fmt"
	"go-backend-discord/modules/database"
	e "go-backend-discord/modules/errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var AuthPath = "/auth"
var AuthRoutes = []Route{
	RegisterLocalRoute,
	LoginDiscordRoute,
	GetUserRoute,
	LoginTwitchRoute,
	GetTwitchRoute,
}

var RegisterLocalRoute = Route{
	Path:       fmt.Sprintf("GET %s/test", AuthPath),
	Handler:    RegisterLocal,
	Middleware: BasicMiddleware,
}

var LoginDiscordRoute = Route{
	Path:       fmt.Sprintf("GET %s/discord/redirect", AuthPath),
	Handler:    LoginDiscord,
	Middleware: BasicMiddleware,
}

var LoginTwitchRoute = Route{
	Path:       fmt.Sprintf("GET %s/twitch/redirect", AuthPath),
	Handler:    LoginTwitch,
	Middleware: AuthMiddleware,
}

var GetUserRoute = Route{
	Path:       fmt.Sprintf("GET %s/@me", AuthPath),
	Handler:    GetCurrentUser,
	Middleware: AuthMiddleware,
}
var GetTwitchRoute = Route{
	Path:       fmt.Sprintf("GET %s/@twitch", AuthPath),
	Handler:    GetCurrentTwitch,
	Middleware: AuthMiddleware,
}

func RegisterLocal(w http.ResponseWriter, r *http.Request) {

}

func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(database.User)
	marshal, _ := json.Marshal(user)
	_, _ = w.Write(marshal)
}

func GetCurrentTwitch(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(database.User)
	var platform database.PlatformConnection
	twitchUser := database.Db.Where(&database.PlatformConnection{
		Platform: "Twitch",
		UserID:   user.ID,
	}).Find(&platform)
	if twitchUser.Error != nil {
		e.AuthError(twitchUser.Error, w)
		return
	}
	marshal, _ := json.Marshal(platform)
	_, _ = w.Write(marshal)
}

func LoginDiscord(w http.ResponseWriter, r *http.Request) {
	var err error
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if os.Getenv("DEBUG") == "true" {
		println("LoginDiscord: " + code)
	}
	info, err := GetDiscordOAuth(code)
	if err != nil {
		e.AuthError(err, w)
		return
	}

	acc, err := GetDiscordAccountFromToken(info.AccessToken)

	if err != nil {
		e.AuthError(err, w)
		return
	}
	search := &database.User{
		DiscordID: acc.Id,
		Username:  acc.Username,
		Email:     acc.Email,
	}

	err = database.Db.FirstOrCreate(search).Error
	if err != nil {
		e.AuthError(err, w)
		return
	}
	if os.Getenv("DEBUG") == "true" {
		println("LoginDiscord: " + code)
		searchData, _ := json.Marshal(search)
		println(string(searchData))
	}

	session, err := CreateSession(search)
	if err != nil {
		e.AuthError(err, w)
		return
	}
	w.Header().Set("Set-Cookie", fmt.Sprintf("Authorization=Session %s; HttpOnly; Expires=%s; Path=%s", session,
		time.Now().Add(time.Hour*24*7).UTC().Format(http.TimeFormat), "/"))
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

func LoginTwitch(w http.ResponseWriter, r *http.Request) {
	var err error
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth, err := GetTwitchOAuth(code)
	if err != nil {
		e.AuthError(err, w)
		return
	}

	twitchUser, err := GetTwitchAccountFromToken(auth.AccessToken)
	if err != nil {
		e.AuthError(err, w)
		return
	}
	user := r.Context().Value("user").(database.User)
	res := database.Db.Save(&database.PlatformConnection{
		Platform:     "Twitch",
		PlatformID:   twitchUser.Login,
		User:         user,
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		ExpiresIn:    auth.ExpiresIn,
	})
	if res.Error != nil {
		e.ServerError(res.Error, w)
		return
	}
	data, _ := json.Marshal(twitchUser)
	_, _ = w.Write(data)
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
