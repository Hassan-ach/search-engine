package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"spider/internal/parser"
)

func main() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://en.wikipedia.org/wiki/CPP", nil)
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.118 Safari/537.36",
	)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US")
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}
	reader := bytes.NewReader(body)
	data, err := parser.Html(reader)
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}
