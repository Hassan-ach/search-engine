package store

import (
	"log"
	"os"
)

func Write(cnt string) {
	f, err := os.OpenFile("file.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.WriteString("##############################################################\n")
	f.WriteString(cnt)
}

func PostHtml(u string, b []byte) {
	// TODO: Store HTML content in db to process later
}

func NewUrl() (string, bool) {
	// TODO: retrive new url from cache
	return "https://www.spaceappschallenge.org", true
}
