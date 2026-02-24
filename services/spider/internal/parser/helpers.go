package parser

import (
	"strings"

	"spider/internal/entity"

	"golang.org/x/net/html"
)

func extrantMeta(n *html.Node) entity.MetaData {
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

func traverse(n *html.Node, visit func(*html.Node)) {
	visit(n)
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		traverse(child, visit)
	}
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func isIconLink(n *html.Node) bool {
	for _, v := range n.Attr {
		if v.Key == "rel" && v.Val == "icon" {
			return true
		}
	}
	return false
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
