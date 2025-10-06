package crawler

// TODO:
//[] impl DB store for (meta data and html)
//[] impl proper logger and metric for monitoring
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
// how Crawler can crawl it Host
type Crawler struct {
	Host
	CacheClient *redis.Client
	Ctx         context.Context
}

// Host is type that contain all the roles for a specific host
type Host struct {
	MaxRetry      int
	MaxPages      int
	Delay         int
	Name          string
	AllowedUrls   []string
	NotAllwedPaths []string
	DiscovedURLs  *utils.SetQueu[string]
	VisitedURLs   *utils.Set[string]
}

// Crawl is entry point for the crawl can start working
func (c *Crawler) Crawl() {
	defer func() {
		err := c.CacheClient.HDel(c.Ctx, "inProg", c.Host.Name).Err()
		if err != nil {
			fmt.Println(err)
		}
		err = c.CacheClient.SAdd(c.Ctx, "complHost", c.Host.Name).Err()
		if err != nil {
			fmt.Println(err)
		}
	}()
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
		c.addUrls(data.Links.GetAll())
		fmt.Printf("Processed %d Pages for this Host: %s\n", pages, c.Host.Name)

		if c.MaxPages > 0 && pages >= c.MaxPages {
			fmt.Printf("Max pages (%d) reached; stopping\n", c.MaxPages)
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
	for _, notAllowed := range c.NotAllwedPaths {
		if strings.HasPrefix(u.Path, notAllowed) {
			fmt.Printf("This is not allowed path for this HOST: %s, PATH: %s\n", u.Host, u.Path)
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
			l.Host = c.Host.Name
		}
		if l.Scheme == "" {
			l.Scheme = "https"
		}
		if c.VisitedURLs.Contains(l.String()) {
			continue
		}
		if l.Host == c.Host.Name {
			c.DiscovedURLs.Push(l.String())
		} else {
			err := c.CacheClient.SAdd(c.Ctx, l.Host, l.String()).Err()
			if err != nil {
				fmt.Printf("Error While Adding URLs to Redis %v\n", err)
				continue
			}
			err = c.CacheClient.SAdd(c.Ctx, "newHost", l.Host).Err()
			if err != nil {
				fmt.Printf("Error While Adding Host to Redis %v\n", err)
				continue
			}
		}
	}
}

func (c *Crawler) process(u string) (*parser.Page, error) {
	body, statusCode, err := utils.GetReq(u, c.MaxRetry, c.Delay)
	if err != nil {
		fmt.Printf("Error while sending GET Request for\n\t URL: %s\n", u)
		return nil, err
	}
	page, err := parser.Html(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error while Parsing HTML content:\n\t URL: %s\n", u)
		return nil, err
	}
	if page.Url == "" {
		page.Url = u
	}
	page.StatusCode = statusCode
	page.HTML = body

	store.PostHtml(u, body)
	return page, nil
}
