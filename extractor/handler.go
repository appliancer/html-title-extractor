package extractor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
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

	response, err := handler.run(nConcurrent)
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

type result struct {
	title string
	err   error
}

func (handler *Handler) run(nConcurrent int) ([]byte, error) {
	responseObj := struct {
		Titles     []string `json:"titles"`
		Successful int      `json:"successful"`
		Failed     int      `json:"failed"`
	}{}

	urls := genChan(handler.urls...)
	for result := range extractTitles(handler.httpGetter, urls, nConcurrent) {
		if result.err != nil {
			log.Println("failed to extract title:", result.err)
			responseObj.Failed++
		} else {
			responseObj.Titles = append(responseObj.Titles, result.title)
			responseObj.Successful++
		}
	}

	sort.Strings(responseObj.Titles) // ensure deterministic response

	response, err := json.Marshal(responseObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	return response, nil
}

func extractTitles(httpGetter HttpGetter, urls <-chan string, nConcurrent int) <-chan result {
	results := make(chan result)

	var wg sync.WaitGroup
	wg.Add(nConcurrent)

	for i := 0; i < nConcurrent; i++ {
		go func() {
			for url := range urls {
				title, err := extractTitle(httpGetter, url)
				results <- result{title: title, err: err}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func genChan[T any](items ...T) chan T {
	c := make(chan T)
	go func() {
		for _, item := range items {
			c <- item
		}
		close(c)
	}()
	return c
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
