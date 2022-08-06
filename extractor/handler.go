package extractor

import (
	"log"
	"net/http"
)

type Handler struct{}

func (Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	_, err := writer.Write([]byte("hello\n"))
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Println("failed to write response:", err)
	}
}
