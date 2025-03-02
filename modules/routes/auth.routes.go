package routes

import (
	"encoding/json"
	"fmt"
	"go-backend-discord/modules/database"
	e "go-backend-discord/modules/errors"
	"io"
	"log"
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

var GetUserRoute = Route{
	Path:       fmt.Sprintf("GET %s/@me", AuthPath),
	Handler:    GetCurrentUser,
	Middleware: AuthMiddleware,
}

func RegisterLocal(w http.ResponseWriter, r *http.Request) {

}

func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(database.User)
	marshal, _ := json.Marshal(user)
	w.Write(marshal)
}

func LoginDiscord(w http.ResponseWriter, r *http.Request) {
	var err error
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
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
	println(acc)
	var user database.User
	search := &database.User{
		DiscordID: acc.Id,
		Username:  acc.Username,
		Email:     acc.Email,
	}

	err = database.Db.FirstOrCreate(&user, search).Error
	if err != nil {
		e.AuthError(err, w)
		return
	}

	log.Println(user)

	session, err := CreateSession(user)
	if err != nil {
		e.AuthError(err, w)
		return
	}
	w.Header().Set("Set-Cookie", fmt.Sprintf("Authorization=Session %s; HttpOnly; Expires=%s; Path=%s", session,
		time.Now().Add(time.Hour*24*7).UTC().Format(http.TimeFormat), "/"))
}

func GetDiscordOAuth(code string) (*DiscordAuth, error) {
	httpClient := &http.Client{}
	body := strings.NewReader(fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", code, "http://localhost:8080/auth/discord/redirect"))
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
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
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
	defer httpResp.Body.Close()
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
