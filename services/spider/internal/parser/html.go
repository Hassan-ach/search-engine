package parser

import (
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/net/html"

	"spider/internal/utils"
)

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
	Images *utils.Set[string]
	Links  *utils.Set[string]
}

func (p *Page) String() {
	fmt.Println("===== Page Info =====")
	fmt.Printf("URL: %s\n", p.Url)
	fmt.Printf("Status Code: %d\n", p.StatusCode)
	fmt.Printf("Crawled At: %s\n", p.CrawledAt.Format(time.RFC3339))
	fmt.Println()

	fmt.Println("----- Meta Data -----")
	fmt.Printf("Title: %s\n", p.Title)
	fmt.Printf("Description: %s\n", p.Description)
	fmt.Printf("Type: %s\n", p.Type)
	fmt.Printf("Site Name: %s\n", p.SiteName)
	fmt.Printf("Locale: %s\n", p.Local)
	if len(p.Keywords) > 0 {
		fmt.Printf("Keywords: %s\n", strings.Join(p.Keywords, ", "))
	}
	if len(p.Icons) > 0 {
		fmt.Printf("Icons: %s\n", strings.Join(p.Icons, ", "))
	}
	fmt.Println()

	fmt.Println("----- Links -----")
	for _, l := range p.Links.GetAll() {
		fmt.Printf("  - %s\n", l)
	}

	fmt.Println("----- Images -----")
	for _, img := range p.Images.GetAll() {
		fmt.Printf("  - %s\n", img)
	}

	fmt.Println("=====================")
}

// Html id function tacks an HTML document and will return a
// pointer for Data struct content
func Html(r io.Reader) (*Page, error) {
	log := utils.Log.Parsing()
	log.Info("Starting HTML content parsing")
	start := time.Now()

	doc, err := html.Parse(r)
	if err != nil {
		log.Error("Failed to parse HTML content", "error", err)
		return nil, err
	}

	urls, imags, desc, metaData := processNodes(doc)

	metaData.CrawledAt = time.Now()
	if len(metaData.Description) == 0 {
		metaData.Description = desc
	}
	u := utils.NewSet[string]()
	u.BatchAdd(urls...)
	i := utils.NewSet[string]()
	i.BatchAdd(imags...)

	log.Info(
		"HTML parsing complete",
		"execTime", time.Since(start).Seconds(),
		"linksFound", u.Len(),
		"imagesFound", i.Len(),
	)

	return &Page{
		MetaData: metaData,
		Links:    u,
		Images:   i,
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
			metaData = mergeMetaData(metaData, getMetaData(node))
		case "link":
			icon := getIcon(node)
			if icon != "" {
				metaData.Icons = append(metaData.Icons, icon)
			}
		case "a":
			rawURL := getAttr(node, "href")

			u, ok := utils.NormalizeUrl(rawURL, "")
			if ok {
				urls = append(urls, u)
			} else {
				utils.Log.Parsing().Debug("Skipping invalid or unhandled URL", "rawURL", rawURL)
			}
		case "img":
			s := getAttr(node, "src")
			if s != "" {
				imgs = append(imgs, s)
			}
		case "title":
			if node.FirstChild != nil && node.FirstChild.Type == html.TextNode {
				metaData.Title = strings.TrimSpace(node.FirstChild.Data)
			}
		}
	}

	if node.Type == html.TextNode {
		text := strings.ToLower(strings.TrimSpace(node.Data))
		if text != "" {
			desc += "\n" + text
		}
	}

	for n := node.FirstChild; n != nil; n = n.NextSibling {
		u, i, d, m := processNodes(n)
		metaData = mergeMetaData(metaData, m)
		urls = append(urls, u...)
		imgs = append(imgs, i...)
		if d != "" {
			desc += strings.TrimSpace(d)
		}
	}
	if len(desc) > 200 {
		desc = desc[:200]
	}
	desc = strings.TrimSpace(desc)
	return
}

func getMetaData(n *html.Node) MetaData {
	prop := getMetaProperty(n)
	content := getMetaContent(n)
	if content == "" {
		return MetaData{}
	}
	var m MetaData
	switch prop {
	case "url":
		m.Url = content
	case "title":
		m.Title = content
	case "description":
		m.Description = content
	case "type":
		m.Type = content
	case "site_name":
		m.SiteName = content
	case "locale":
		m.Local = content
	case "keywords":
		m.Keywords = strings.Split(content, ",")
		for i := range m.Keywords {
			m.Keywords[i] = strings.TrimSpace(m.Keywords[i])
		}
	}
	return m
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

func mergeMetaData(base, other MetaData) MetaData {
	if other.Url != "" {
		base.Url = other.Url
	}
	if other.Title != "" {
		base.Title = other.Title
	}
	if other.Description != "" {
		base.Description = other.Description
	}
	if other.Type != "" {
		base.Type = other.Type
	}
	if other.SiteName != "" {
		base.SiteName = other.SiteName
	}
	if other.Local != "" {
		base.Local = other.Local
	}
	base.Keywords = append(base.Keywords, other.Keywords...)
	base.Icons = append(base.Icons, other.Icons...)
	return base
}
