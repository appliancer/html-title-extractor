package extractor

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type mockHttpGetter struct{}

func (mockHttpGetter) Get(url string) (*http.Response, error) {
	switch url {
	case "http://example.com/test1":
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("....<title>test\n1</title>.......")),
		}, nil
	case "http://example.com/test2":
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("...\n...<title>\n    test 2\n</title>....\n.....")),
		}, nil
	default:
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       http.NoBody,
		}, nil
	}
}

func TestHandler_ServeHTTP(t *testing.T) {
	type testCase struct {
		urls               []string
		method             string
		query              string
		expectedStatusCode int
		expectedBody       []byte
	}

	testCases := []testCase{
		{
			urls:               []string{"http://example.com/test1", "http://example.com/test2"},
			method:             http.MethodGet,
			query:              "n_concurrent=2",
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte(`{"titles":["test\n1","test 2"],"successful":2,"failed":0}`),
		},
		{
			urls:               []string{"http://example.com/test1", "http://example.com/test2"},
			method:             http.MethodPost,
			query:              "n_concurrent=2",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			urls:               []string{"http://example.com/test1", "http://example.com/test2"},
			method:             http.MethodGet,
			query:              "",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       []byte("missing or invalid n_concurrent parameter"),
		},
		{
			urls:               []string{"invalid"},
			method:             http.MethodGet,
			query:              "n_concurrent=2",
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte(`{"titles":null,"successful":0,"failed":1}`),
		},
		{
			urls:               []string{"http://example.com/test2", "invalid"},
			method:             http.MethodGet,
			query:              "n_concurrent=2",
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte(`{"titles":["test 2"],"successful":1,"failed":1}`),
		},
	}

	for i, tc := range testCases {
		handler := NewHandler(
			mockHttpGetter{},
			tc.urls,
		)

		responseRecorder := httptest.NewRecorder()
		handler.ServeHTTP(responseRecorder, &http.Request{
			Method: tc.method,
			URL:    &url.URL{RawQuery: tc.query},
		})

		response := responseRecorder.Result()
		if response.StatusCode != tc.expectedStatusCode {
			t.Errorf("case %d: unexpected status code: %d", i, response.StatusCode)
		}

		body, _ := io.ReadAll(response.Body)
		if !bytes.Equal(body, tc.expectedBody) {
			t.Errorf("case %d: unexpected body: %s", i, body)
		}
	}
}

func TestClampInt(t *testing.T) {
	if clampInt(6, 8, 14) != 8 {
		t.Error("should not be clamped")
	}
	if clampInt(10, 8, 14) != 10 {
		t.Error("should be clamped to min")
	}
	if clampInt(16, 8, 14) != 14 {
		t.Error("should be clamped to max")
	}

	func() {
		defer func() {
			if recover() == nil {
				t.Error("should panic")
			}
		}()
		clampInt(10, 14, 8)
	}()
}
