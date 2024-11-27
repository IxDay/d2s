package server

import "net/http"

type Handler interface {
	Handle(r *http.Request) error
}
