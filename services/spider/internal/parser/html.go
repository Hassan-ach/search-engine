package parser

import (
	"fmt"
	"io"

	"golang.org/x/net/html"
)

// Data is struct that holed the data parsed from HTML page
type Data struct {
	Urls   []string
	Images []string
	// Content string
}

// Html id function tacks an HTML document and will return a
// pointer for Data struct content
func Html(r io.Reader) (*Data, error) {
	fmt.Println("Start parsing HTML content")
	doc, err := html.Parse(r)
	if err != nil {
		fmt.Println("in Html, fail to pars HTML content\n Error: ", err)
		return nil, err
	}
	urlNodes, _, _ := getNodes(doc)
	var urls []string = make([]string, len(urlNodes))
	for i, node := range urlNodes {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				if attr.Val[0] != '#' {
					urls[i] = attr.Val
				}
			}
		}
	}
	// content := strings.Join(textContent, " ")
	// content = strings.TrimSpace(
	// 	content,
	// )
	// content = strings.ReplaceAll(content, "\n", " ")

	return &Data{
		Urls:   urls,
		Images: nil,
		// Content: content,
	}, nil
}

func getNodes(node *html.Node) (urls []*html.Node, imgs []*html.Node, cnt []string) {
	if node == nil {
		return
	}
	if node.Type == html.ElementNode && (node.Data == "script" || node.Data == "style") {
		return
	}
	if node.Type == html.ElementNode {
		switch node.Data {
		case "a":
			urls = append(urls, node)
		case "img":
			imgs = append(imgs, node)
		}
	}
	if node.Type == html.TextNode {
		cnt = append(cnt, node.Data)
	}
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		u, i, c := getNodes(n)
		urls = append(urls, u...)
		imgs = append(imgs, i...)
		cnt = append(cnt, c...)
	}
	return
}
