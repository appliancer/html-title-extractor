package main

import (
	"html-title-extractor/extractor"
	"log"
	"net/http"
	"strconv"
)

const (
	port = 9000
)

func main() {
	http.Handle("/extract-titles", extractor.Handler{})

	server := http.Server{
		Addr: ":" + strconv.Itoa(port),
	}

	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		log.Println("main: Server closed")
	} else if err != nil {
		log.Println("main: Server closed with error:", err)
	}
}
