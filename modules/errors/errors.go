package AppError

import "net/http"

func ServerError(err error, w http.ResponseWriter) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func InputError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}

func AuthError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(err.Error()))
}
