package routes

import (
	"encoding/json"
	"fmt"
	"go-backend-discord/modules/database"
	e "go-backend-discord/modules/errors"
	"net/http"
	"os"
	"time"
)

var AuthPath = "/auth"
var AuthRoutes = []Route{
	LoginDiscordRoute,
	GetUserRoute,
	LoginTwitchRoute,
	GetTwitchRoute,
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
