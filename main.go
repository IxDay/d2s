package main

import (
	"log"
	"net/http"

	"github.com/douglasmakey/oauth2-example/handlers"
)

func main() {
	// We create a simple server using http.Server and run.
	server := &http.Server{
		Addr:    ":8090",
		Handler: handlers.New(),
	}

	log.Printf("Starting HTTP Server. Listening at %q", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed!")
	}
}
