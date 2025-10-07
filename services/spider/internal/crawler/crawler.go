package crawler

// TODO:
import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"spider/internal/parser"
	"spider/internal/store"
	"spider/internal/utils"
)

// Crawler is type that contain all the necessary informations about
// how Crawler can crawl it Host
type Crawler struct {
	Host
	CacheClient *store.CacheClient
	Ctx         context.Context
}

// Host is type that contain all the roles for a specific host
type Host struct {
	MaxRetry       int
	MaxPages       int
	Delay          int
	Name           string
	AllowedUrls    []string
	NotAllwedPaths []string
	DiscovedURLs   *utils.SetQueu[string]
	VisitedURLs    *utils.Set[string]
}

func (c *Crawler) String() string {
	return fmt.Sprintf(
		"Crawler Info\n"+
			"  Host: %s\n"+
			"  MaxRetry: %d\n"+
			"  MaxPages: %d\n"+
			"  Delay: %ds\n"+
			"  Allowed URLs: %v\n"+
			"  Disallowed Paths: %v\n"+
			"  Discovered URLs: %d\n"+
			"  Visited URLs: %d\n"+
			"  Cache Connected: %v\n",
		c.Name,
		c.MaxRetry,
		c.MaxPages,
		c.Delay,
		c.AllowedUrls,
		c.NotAllwedPaths,
		c.DiscovedURLs.Len(),
		c.VisitedURLs.Len(),
		c.CacheClient != nil,
	)
}

// Crawl is entry point for the crawl can start working
func (c *Crawler) Crawl() {
	log := utils.Log.General().With("host", c.Host.Name, "operation", "Crawl")
	log.Info("Starting crawl process for host")

	defer func() {
		log.Info("Crawl process finished. Performing cleanup.")
		err := store.Completed(c.Host.Name)
		if err != nil {
			log.Error("Cleanup Failed", "host", c.Host.Name, "error", err)
		}
	}()

	pages := 0
	for !c.DiscovedURLs.Empty() {
		u, ok := c.getUrl()
		if !ok {
			continue
		}

		log.Info("Processing URL", "url", u)
		data, err := c.process(u)
		if err != nil {
			log.Warn("Failed to process URL", "url", u, "error", err)
			continue
		}

		c.VisitedURLs.Add(u)
		pages++
		c.addUrls(data.Links.GetAll())
		log.Info("Page processed successfully", "host", c.Host.Name, "processedPages", pages)

		if c.MaxPages > 0 && pages >= c.MaxPages {
			log.Info("Max pages reached; stopping crawl for host", "maxPages", c.MaxPages)
			break
		}
		time.Sleep(time.Duration(c.Delay) * time.Second)
	}
	if c.DiscovedURLs.Empty() {
		log.Info("Discovered URLs queue is empty; crawl concluded.")
	}
}

func (c *Crawler) getUrl() (string, bool) {
	log := utils.Log.General().With("host", c.Host.Name, "operation", "getUrl")

	s, ok := c.DiscovedURLs.Pop()
	if !ok {
		log.Debug("Discovered URLs queue is empty.")
		return "", false
	}
	if c.VisitedURLs.Contains(s) {
		log.Debug("URL already visited, skipping.", "url", s)
		return "", false
	}

	u, err := url.Parse(s)
	if err != nil {
		log.Warn("Failed to parse URL, skipping.", "url", s, "error", err)
		return "", false
	}

	for _, notAllowed := range c.NotAllwedPaths {
		if strings.HasPrefix(u.Path, notAllowed) {
			log.Debug(
				"URL path is disallowed, skipping.",
				"url",
				s,
				"path",
				u.Path,
				"rule",
				notAllowed,
			)
			return "", false
		}
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	return u.String(), true
}

func (c *Crawler) addUrls(s []string) {
	log := utils.Log.General().With("host", c.Host.Name, "operation", "addUrls")
	addedToQueue := 0
	addedToDiscovery := 0

	for _, u := range s {
		l, err := url.Parse(u)
		if err != nil {
			log.Warn("Could not parse discovered URL", "url", u, "error", err)
			continue
		}
		if l.Host == "" {
			l.Host = c.Host.Name
		}
		if l.Scheme == "" {
			l.Scheme = "https"
		}
		finalURL := l.String()

		if c.VisitedURLs.Contains(finalURL) {
			continue
		}
		if l.Host == c.Host.Name {
			c.DiscovedURLs.Push(l.String())
			addedToQueue++

		} else {
			if err := store.AddLink(l.Host, finalURL); err != nil {
				utils.Log.Cache().Debug("Failed to add URL", "url", finalURL, "host", l.Host, "error", err)
				continue
			}
			if err := store.AddHost(l.Host); err != nil {
				utils.Log.Cache().Debug("Failed to add host", "host", l.Host, "error", err)
				continue
			}
			addedToDiscovery++
		}
	}
	log.Debug(
		"URL addition complete",
		"addedToCurrentQueue",
		addedToQueue,
		"addedToGlobalDiscovery",
		addedToDiscovery,
	)
}

func (c *Crawler) process(u string) (*parser.Page, error) {
	body, statusCode, err := utils.GetReq(u, c.MaxRetry, c.Delay)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}

	page, err := parser.Html(bytes.NewReader(body))
	if err != nil {
		utils.Log.Parsing().Error("Failed to parse HTML content", "url", u, "error", err)
		return nil, fmt.Errorf("HTML parsing failed: %w", err)
	}

	if page.Url == "" {
		page.Url = u
	}
	page.StatusCode = statusCode
	page.HTML = body

	// store.PostHtml(u, body)
	return page, nil
}
