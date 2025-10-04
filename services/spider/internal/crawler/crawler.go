package crawler

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"spider/internal/parser"
	"spider/internal/store"
	"spider/internal/utils"
)

// Crawler is type that contain all the necessary informations about
// how Crawler can crawl the Domain
type Crawler struct {
	MaxRetry      int
	Delay         int
	MaxPages      int
	Host          string
	DiscovedURLs  *utils.SetQueu[string]
	VisitedURLs   *utils.Set[string]
	AllowedUrls   []string
	NotAllwedUrls []string
}

type MetaData struct {
	Url         string
	Title       string
	Description string
}

type Domain struct {
	name string
}

func (c *Crawler) Crawl() {
	attempt := 0
	for !c.DiscovedURLs.Empty() {
		u, ok := c.getUrl()
		if !ok {
			continue
		}

		fmt.Printf("Start Crawing URL: %s\n", u.String())
		if ok {
			data, err := c.process(u.String())
			if err != nil {
				fmt.Printf("error while crawling URL: %s, ERROR: %v\n", u.String(), err)
				continue
			}

			c.VisitedURLs.Add(u.String())
			c.addUrls(data.Urls)
			fmt.Println("VisitedURLs:")
			c.VisitedURLs.Print()

			attempt++
			if c.MaxPages > 0 && attempt > int(c.MaxPages) {
				break
			}
		}
	}
}

func (c *Crawler) getUrl() (*url.URL, bool) {
	s, ok := c.DiscovedURLs.Pop()
	if !ok {
		return nil, false
	}
	if c.VisitedURLs.Contains(s) {
		return nil, false
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, false
	}
	for _, notAllowed := range c.NotAllwedUrls {
		if strings.HasPrefix(u.Path, notAllowed) {
			fmt.Println("This is not allowed path, PATH: %s", u.Path)
			return nil, false
		}
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	return u, true
}

func (c *Crawler) addUrls(s []string) {
	for _, u := range s {
		l, err := url.Parse(u)
		if l.Host == "" {
			l.Host = c.Host
		}
		if l.Scheme == "" {
			l.Scheme = "https"
		}
		if err != nil {
			fmt.Printf("Error while adding URL: %s, ERROR: %v\n", u, err)
			continue
		}
		if c.VisitedURLs.Contains(l.String()) {
			continue
		}

		if l.Host == c.Host {
			c.DiscovedURLs.Push(l.String())
		} else {
			store.AddNewHost(l.String())
		}
	}
}

func (c *Crawler) process(u string) (*parser.Data, error) {
	body, err := utils.GetReq(u, c.MaxRetry)
	if err != nil {
		fmt.Printf("Error while sending GET Request for\n\t URL: %s\n", u)
		return nil, err
	}
	data, err := parser.Html(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error while Parsing HTML content:\n\t URL: %s\n", u)
		return nil, err
	}
	go store.PostHtml(u, body)
	return data, nil
}
