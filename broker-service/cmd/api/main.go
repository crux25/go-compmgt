package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	webPort = "9091"
)

type Config struct{}

func main() {
	app := &Config{}

	log.Printf("starting broker service on port %s \n", webPort)

	// Define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// Start the http server
	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
