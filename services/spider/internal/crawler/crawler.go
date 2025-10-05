package crawler

// TODO:
//[] impl DB store for (meta data and html)
//[] clean the code
//[] impl proper logger and metric for monitoring
//[] add comments
import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

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
	CacheClient   *redis.Client
	Ctx           context.Context
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
	pages := 0
	for !c.DiscovedURLs.Empty() {
		u, ok := c.getUrl()
		if !ok {
			continue
		}

		fmt.Printf("Start Crawing URL: %s\n", u)
		data, err := c.process(u)
		if err != nil {
			fmt.Printf("error while crawling URL: %s, ERROR: %v\n", u, err)
			continue
		}

		c.VisitedURLs.Add(u)
		pages++
		c.addUrls(data.Urls)
		fmt.Printf("Processed %d Pages for this Host: %s\n", pages, c.Host)

		if c.MaxPages > 0 && pages >= c.MaxPages {
			fmt.Printf("Max pages (%d) reached; stopping", c.MaxPages)
			break
		}
		time.Sleep(time.Duration(c.Delay) * time.Second)
	}
	if c.DiscovedURLs.Empty() {
		fmt.Println("the set is empty")
	}
}

func (c *Crawler) getUrl() (string, bool) {
	s, ok := c.DiscovedURLs.Pop()
	if !ok {
		return "", false
	}
	if c.VisitedURLs.Contains(s) {
		return "", false
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", false
	}
	for _, notAllowed := range c.NotAllwedUrls {
		if strings.HasPrefix(u.Path, notAllowed) {
			fmt.Println("This is not allowed path for this HOST: %s, PATH: %s", u.Host, u.Path)
			return "", false
		}
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	return u.String(), true
}

func (c *Crawler) addUrls(s []string) {
	for _, u := range s {
		l, err := url.Parse(u)
		if err != nil {
			fmt.Printf("Error while adding URL: %s, ERROR: %v\n", u, err)
			continue
		}
		if l.Host == "" {
			l.Host = c.Host
		}
		if l.Scheme == "" {
			l.Scheme = "https"
		}
		if c.VisitedURLs.Contains(l.String()) {
			continue
		}

		if l.Host == c.Host {
			c.DiscovedURLs.Push(l.String())
		} else {
			c.CacheClient.SAdd(c.Ctx, "DiscovedHosts", l.String())
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
	store.PostHtml(u, body)
	return data, nil
}
