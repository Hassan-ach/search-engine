package store

import (
	"fmt"
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

var x int = 0

func PostHtml(u string, b []byte) {
	// TODO: Store HTML content in db to process later
	f, err := os.OpenFile(fmt.Sprintf("%d.html", x), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	x++
	defer f.Close()
	f.Write(b)
}

func NewHost() (string, bool) {
	// TODO: retrive new url from cache
	return "spaceappschallenge.org", true
}
