package main

import (
	"fmt"

	"query-engine/internal"
)

func main() {
	pages, err := internal.Run("math engine")
	if err != nil {
		fmt.Println("Error running query engine:", err)
		return
	}

	for _, page := range pages {
		fmt.Printf("URL: %s, Global Score: %f\n", page.URL, page.GlobalScore)
	}
}
