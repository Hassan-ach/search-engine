package crawler

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Crawler is type that contain all the necessary informations about
// how Crawler can crawl the Domain
type Crawler struct {
	MaxRetry     uint8
	Delay        uint8
	MaxPages     uint8
	StartUrl     string
	DiscovedURLs map[string]bool
	VisitedURLs  map[string]bool
}

type MetaData struct {
	url         string
	title       string
	description string
}

type Domain struct {
	name string
	logo string
}

func (c *Crawler) Run() {
	body, err := getReq(c.StartUrl, c.MaxRetry)
	if err != nil {
		fmt.Println("Fail to Crawl this Url: %s", c.StartUrl)
	}
}

func getReq(url string, MaxRetry uint8) (body []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.118 Safari/537.36",
	)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US")

	if err != nil {
		fmt.Printf("Request initialization failed", err)
		return nil, err
	}
	for attempt := 0; attempt < int(MaxRetry); attempt++ {
		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("retry %d failed\nURL: %s", attempt+1, url)
			continue
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if res.StatusCode > 299 {
			fmt.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
			continue
		}
		if err != nil {
			fmt.Printf("In getHtml error while reading the response body:  %v", err)
			continue
		}
		time.Sleep(5 * time.Second)
	}
	return body, err
}
