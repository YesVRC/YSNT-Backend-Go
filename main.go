package main

import "net/http"

func main() {
	s := http.NewServeMux()

	s.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/plain")
	})

	panic(http.ListenAndServe(":8080", s))
}
