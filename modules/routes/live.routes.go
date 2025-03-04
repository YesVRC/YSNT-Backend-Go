package routes

import (
	"fmt"
	"net/http"
)

const LivePath = "/live"

var LiveRoutes = []Route{
	PublishRoute,
}

var PublishRoute = Route{
	Path:       fmt.Sprintf("POST %s/publish", LivePath),
	Handler:    PublishHandler,
	Middleware: BasicMiddleware,
}

func PublishHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	println(r.Form)
	http.Redirect(w, r, "rtmp://ysnt.live/live/test-redirect", http.StatusFound)
}
