package parser

import (
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/net/html"

	"spider/internal/utils"
)

// TODO: use the new struct Page to store the data and meta data of each url crawled

// Data is struct that holed the data parsed from HTML page
type Data struct {
	Urls   []string
	Images []string
	// Content string
}

type MetaData struct {
	Url         string
	Title       string
	Description string
	Type        string
	SiteName    string
	Local       string
	Keywords    []string
	Icons       []string
	CrawledAt   time.Time
}

type Page struct {
	MetaData
	StatusCode int
	HTML       []byte
	// Text       string
	Images []string
	Links  []string
}

// Html id function tacks an HTML document and will return a
// pointer for Data struct content
func Html(r io.Reader) (*Page, error) {
	fmt.Println("Start parsing HTML content")
	doc, err := html.Parse(r)
	if err != nil {
		fmt.Println("in Html, fail to pars HTML content\n Error: ", err)
		return nil, err
	}
	urls, imags, desc, metaData := processNodes(doc)

	metaData.CrawledAt = time.Now()
	if len(metaData.Description) == 0 {
		metaData.Description = desc
	}
	return &Page{
		MetaData: metaData,
		Links:    urls,
		Images:   imags,
	}, nil
}

func processNodes(node *html.Node) (urls []string, imgs []string, desc string, metaData MetaData) {
	if node == nil {
		return
	}
	if node.Type == html.ElementNode && (node.Data == "script" || node.Data == "style") {
		return
	}
	if node.Type == html.ElementNode {
		switch node.Data {
		case "meta":
			metaData = getMetaData(node, metaData)
		case "link":
			metaData.Icons = append(metaData.Icons, getIcon(node))
		case "a":
			u, ok := utils.NormalizeUrl(getAttr(node, "href"), "")
			if ok {
				urls = append(urls, u)
			}
		case "img":
			s := getAttr(node, "src")
			if s != "" {
				imgs = append(imgs, s)
			}
		}
	}
	if len(desc) < 200 && node.Type == html.TextNode {
		desc += "\n" + strings.ToLower(strings.TrimSpace(node.Data))
	}
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		u, i, d, m := processNodes(n)
		metaData = m
		urls = append(urls, u...)
		imgs = append(imgs, i...)
		if len(d) > 200 {
			desc += d
		}
	}
	return
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func getIcon(n *html.Node) string {
	for _, v := range n.Attr {
		if v.Key == "rel" && v.Val == "icon" {
			return getAttr(n, "href")
		}
	}
	return ""
}

func getMetaProperty(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "property" {
			if strings.HasPrefix(a.Val, "og:") {
				return a.Val[3:]
			}
		}
		if a.Key == "name" {
			return a.Val
		}
	}
	return ""
}

func getMetaContent(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "content" {
			return a.Val
		}
	}

	return ""
}

func getMetaData(n *html.Node, m MetaData) MetaData {
	//
	switch getMetaProperty(n) {
	case "url":
		m.Url = getMetaContent(n)
	case "title":
		m.Title = getMetaContent(n)
	case "description":
		m.Description = getMetaContent(n)
	case "type":
		m.Type = getMetaContent(n)
	case "site_name":
		m.SiteName = getMetaContent(n)
	case "loca":
		m.Local = getMetaContent(n)
	case "keywords":
		m.Keywords = append(m.Keywords, getMetaContent(n))
	}
	return m
}
