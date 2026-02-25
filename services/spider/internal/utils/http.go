package utils

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func GetReq(
	client *http.Client,
	url string,
	maxRetry, delay int,
) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("request initialization failed: %w", err)
	}
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.118 Safari/537.36",
	)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US")

	var res *http.Response
	for attempt := range maxRetry {
		if attempt > 0 {
			time.Sleep(time.Second * time.Duration(delay))
		}

		res, err = client.Do(req)
		if err != nil {
			continue
		}

		statusCode := res.StatusCode
		if statusCode >= 500 || statusCode == 429 {
			res.Body.Close()
			continue
		}
		if statusCode >= 400 {
			return nil, statusCode, fmt.Errorf("client error: %d", statusCode)
		}

		body, err := io.ReadAll(io.LimitReader(res.Body, 10<<20)) // 10 MB
		res.Body.Close()
		if err != nil {
			err = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		err = nil
		return body, statusCode, nil
	}

	return nil, 0, fmt.Errorf("all %d retries failed: %w", maxRetry, err)
}
