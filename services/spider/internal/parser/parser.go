package parser

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"spider/internal/entity"
	"spider/internal/utils"

	"golang.org/x/net/html"
)

type Parser struct {
	Client *http.Client
	log    *slog.Logger
}

func NewParser(client *http.Client, logger *utils.Logger) *Parser {
	return &Parser{
		Client: client,
		log:    logger.With("component", "parser"),
	}
}

// func (p *Parser) FetchAndParseNewHostMetaDta(
//
//	ctx context.Context,
//	raw string,
//
//	) (host *entity.Host, err error) {
//		if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
//			raw = "https://" + raw
//		}
//
//		u, err := url.Parse(strings.TrimPrefix(raw, "www."))
//		if err != nil {
//			return
//		}
//
//		r, err := NewRobots(p.Client, u)
//
//		sitemaps := fetchSitemaps(p.Client, r.SiteMaps, u)
//
//		// create Host object
//		host = &entity.Host{
//			MaxRetry:       5,
//			Delay:          r.CrawlDelay,
//			MaxPages:       10,
//			PagesCrawled:   0,
//			Name:           u.Host,
//			AllowedUrls:    r.Allow,
//			NotAllwedPaths: r.Disallow,
//		}
//
//		// persist in store
//		p.Cache.AddHostMetaData(ctx, host.Name, host)
//		p.Cache.AddUrls(ctx, sitemaps)
//
//		return host, nil
//	}
func (p *Parser) ParseHTML(r io.Reader, baseURL string) (*entity.Page, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(baseURL)

	c := newHtmlCollector(u)
	traverse(doc, c.Visit)

	desc := strings.TrimSpace(c.TextBuffer.String())
	if len(desc) > 300 {
		desc = desc[:300]
	}

	if c.Meta.Description == "" {
		c.Meta.Description = desc
	}
	c.Meta.CrawledAt = time.Now()

	return &entity.Page{
		MetaData: c.Meta,
		Links: utils.NewSetFromSlice(
			utils.NormalizeUrls(c.Links, u)).GetAll(),
		Images: utils.NewSetFromSlice(c.Imags).GetAll(),
	}, nil
}
