package extractor

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var titleRegexp = regexp.MustCompilePOSIX("<title>((\n|.)*?)</title>")

func extractTitle(httpGetter HttpGetter, url string) (string, error) {
	body, err := download(httpGetter, url)
	if err != nil {
		return "", fmt.Errorf("failed to download from url %s: %w", url, err)
	}

	title, err := extractTitleFromHtml(body)
	if err != nil {
		return "", fmt.Errorf("failed to extract title from html: %w", err)
	}

	return title, nil
}

func download(httpGetter HttpGetter, url string) (body []byte, err error) {
	response, err := httpGetter.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println("failed to close response body:", err)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unexpected response status code: %d", response.StatusCode)
	}

	body, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func extractTitleFromHtml(body []byte) (string, error) {
	matches := titleRegexp.FindSubmatch(body)
	if len(matches) != 3 {
		return "", fmt.Errorf("invalid number of matches: %d", len(matches))
	}

	title := string(matches[1])
	title = html.UnescapeString(title)
	title = strings.TrimSpace(title)
	return title, nil
}
