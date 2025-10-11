package parser

import (
	"io"
	"strings"
	"time"

	"golang.org/x/net/html"

	"spider/internal/entity"
	"spider/internal/utils"
)

// Html parses an HTML document from the given reader and returns a populated Page struct.
// It extracts metadata, links, images, and generates a short text description.
// Logs parsing start, completion, and execution time.
func Html(r io.Reader) (*entity.Page, error) {
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
		"execTime", time.Since(start),
		"linksFound", u.Len(),
		"imagesFound", i.Len(),
	)

	return &entity.Page{
		MetaData: metaData,
		Links:    u,
		Images:   i,
	}, nil
}

// processNodes recursively traverses an HTML node tree and extracts:
// - href URLs
// - image sources
// - a short text description
// - metadata from <meta> and <link> tags
func processNodes(
	node *html.Node,
) (urls []string, imgs []string, desc string, metaData entity.MetaData) {
	if node == nil {
		return
	}

	// Skip script and style nodes
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

	// Aggregate text content for description
	if node.Type == html.TextNode {
		text := strings.ToLower(strings.TrimSpace(node.Data))
		if text != "" {
			desc += "\n" + text
		}
	}

	// Recursively process child nodes
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		u, i, d, m := processNodes(n)
		metaData = mergeMetaData(metaData, m)
		urls = append(urls, u...)
		imgs = append(imgs, i...)
		if d != "" {
			desc += strings.TrimSpace(d)
		}
	}

	// Limit description length
	if len(desc) > 200 {
		desc = desc[:200]
	}
	desc = strings.TrimSpace(desc)
	return
}

func getMetaData(n *html.Node) entity.MetaData {
	prop := getMetaProperty(n)
	content := getMetaContent(n)
	if content == "" {
		return entity.MetaData{}
	}
	var m entity.MetaData
	switch prop {
	case "url":
		m.URL = content
	case "title":
		m.Title = content
	case "description":
		m.Description = content
	case "type":
		m.Type = content
	case "site_name":
		m.SiteName = content
	case "locale":
		m.Locale = content
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

func mergeMetaData(base, other entity.MetaData) entity.MetaData {
	if other.URL != "" {
		base.URL = other.URL
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
	if other.Locale != "" {
		base.Locale = other.Locale
	}
	base.Keywords = append(base.Keywords, other.Keywords...)
	base.Icons = append(base.Icons, other.Icons...)
	return base
}
