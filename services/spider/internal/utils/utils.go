package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func UrlClean(url string) *url.URL {
	return nil
}

func GetReq(url string, MaxRetry uint8) (body []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.118 Safari/537.36",
	)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US")

	if err != nil {
		fmt.Printf("Request initialization failed\n%v", err)
		return nil, err
	}
	for attempt := 0; attempt < int(MaxRetry) || err != nil; attempt++ {
		if attempt != 0 {
			time.Sleep(5 * time.Second)
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("retry %d failed\nURL: %s", attempt+1, url)
			continue
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if res.StatusCode > 299 {
			fmt.Printf(
				"Response failed with status code: %d and\n body: %s\n",
				res.StatusCode,
				body,
			)
			continue
		}
		if err != nil {
			fmt.Printf("In GetReq error while reading the response body:  %v", err)
			continue
		}
		return body, nil
	}
	return nil, err
}

func Filter[T any](s []T, f func(T) bool) []T {
	var r []T
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}
