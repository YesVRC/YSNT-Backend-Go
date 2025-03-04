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
	r.ParseForm()
	for k, v := range r.Form {
		fmt.Printf("%s: %s \n", k, v)
	}
	http.Redirect(w, r, "test-redirect", http.StatusFound)
}
