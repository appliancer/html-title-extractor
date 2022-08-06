package extractor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const (
	concurrentMin = 1
	concurrentMax = 4
)

type HttpGetter interface {
	Get(url string) (*http.Response, error)
}

type Handler struct {
	httpGetter HttpGetter
	urls       []string
}

func NewHandler(httpGetter HttpGetter, urls []string) *Handler {
	return &Handler{httpGetter: httpGetter, urls: urls}
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	nConcurrentStr := request.URL.Query().Get("n_concurrent")
	nConcurrent, err := strconv.Atoi(nConcurrentStr)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		if _, err := writer.Write([]byte("missing or invalid n_concurrent parameter")); err != nil {
			log.Println("failed to write response:", err)
		}
		return
	}
	nConcurrent = clampInt(nConcurrent, concurrentMin, concurrentMax)

	response, err := handler.run()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Println("failed to extract titles:", err)
		return
	}

	if _, err := writer.Write(response); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Println("failed to write response:", err)
	}
}

func (handler *Handler) run() ([]byte, error) {
	titles := make([]string, 0, len(handler.urls))

	for _, url := range handler.urls {
		body, err := download(handler.httpGetter, url)
		if err != nil {
			return nil, fmt.Errorf("failed to download from url %s: %w", url, err)
		}

		title, err := extractTitleFromHtml(body)
		if err != nil {
			return nil, fmt.Errorf("failed to extract title from html: %w", err)
		}

		titles = append(titles, title)
	}

	titlesJson, err := json.Marshal(titles)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	return titlesJson, nil
}

func clampInt(val, min, max int) int {
	if min > max {
		panic(fmt.Sprintf("min %d is greater than max %d", min, max))
	}
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
