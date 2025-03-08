package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"go-backend-discord/modules/database"
	e "go-backend-discord/modules/errors"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

type Route struct {
	Path       string
	Handler    http.HandlerFunc
	Middleware []func(http.HandlerFunc) http.HandlerFunc
}

var BasicMiddleware = []func(http.HandlerFunc) http.HandlerFunc{
	LoggerMiddleware,
	PanicMiddleware,
}

var AuthMiddleware = []func(http.HandlerFunc) http.HandlerFunc{
	SessionAuthMiddleware,
	LoggerMiddleware,
	PanicMiddleware,
}

func LoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] %s %s %s}\n", time.Now(), r.Method, r.RemoteAddr, r.RequestURI)
		next(w, r)
	}
}

func PanicMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err, string(debug.Stack()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}

func SessionAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session database.Session
		authCookie, _ := r.Cookie("Authorization")
		prefix := strings.TrimPrefix(authCookie.Value, "Session ")
		if prefix == "" {
			e.AuthError(SessionInvalidError, w)
			return
		}
		err := database.Db.Model(&database.Session{}).Preload("User").Where("session_id = ?", prefix).First(&session).Error
		if err != nil {
			e.AuthError(err, w)
			return
		}
		if os.Getenv("DEBUG") == "true" {
			data, _ := json.Marshal(session)
			println(string(data))
		}
		newContext := context.WithValue(r.Context(), "user", session.User)
		next(w, r.WithContext(newContext))
	}
}
