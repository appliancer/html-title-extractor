package extractor

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const (
	concurrentMin = 1
	concurrentMax = 4
)

type Handler struct{}

func (Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
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

	if _, err := writer.Write([]byte("hello\n")); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Println("failed to write response:", err)
	}
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
