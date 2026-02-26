package internal

import (
	"fmt"

	"github.com/trustmaster/go-aspell"
)

func Run(query string) ([]*Page, error) {
	speller, _ := aspell.NewSpeller(map[string]string{
		"lang": "en_US", // American English
	})

	q := get_words_from_query(&speller, query)

	fmt.Printf("Original query: %s\n", query)
	fmt.Printf("Processed query: %v\n", q)

	db := dbSetup()
	pages, err := ranking(db, q)
	if err != nil {
		fmt.Errorf("Error running query engine: %v\n", err)
	}

	return pages, nil
}
