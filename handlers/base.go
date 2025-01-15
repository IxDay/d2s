package handlers

import (
	"net/http"
)

func New() http.Handler {
	mux := http.NewServeMux()
	// Root
	mux.Handle("/", http.FileServer(http.Dir("templates/")))

	// OauthGoogle
	mux.HandleFunc("/auth/login", login)
	mux.HandleFunc("/auth/callback", callback)

	return mux
}
