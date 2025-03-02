package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	var id = r.URL.Query().Get("id")
	if id == "" {
		id = "179031614683217920"
	}
	user, err := dg.User(id)
	if err != nil {
		fmt.Println("error getting user,", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	data, _ := json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
