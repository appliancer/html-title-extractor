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

var urls = []string{
	"https://www.result.si/projekti/",
	"https://www.result.si/o-nas/",
	"https://www.result.si/kariera/",
	"https://www.result.si/blog/",
}

func main() {
	handler := extractor.NewHandler(http.DefaultClient, urls)
	http.Handle("/extract-titles", handler)

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
