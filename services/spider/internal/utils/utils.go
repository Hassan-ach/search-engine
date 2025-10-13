package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func GetReq(url string, maxRetry, delay int) (body []byte, statusCode int, err error) {
	start := time.Now()
	log := Log.Network().With("url", url, "operation", "GetReq")

	defer func() {
		log = log.With("execTime", time.Since(start))
		if err != nil {
			log.Error("GET Request failed", "error", err)
		} else {
			log.Info("GET Request succeeded")
		}
	}()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("request initialization failed: %w", err)
	}
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.118 Safari/537.36",
	)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US")

	client := &http.Client{}
	var res *http.Response
	for attempt := range maxRetry {
		if attempt > 0 {
			log.Debug("Retrying request", "attempt", attempt+1)
			time.Sleep(time.Second * time.Duration(delay))
		}

		res, err = client.Do(req)
		if err != nil {
			continue
		}

		statusCode = res.StatusCode
		if statusCode >= 500 || statusCode == 429 {
			log.Warn("Server error, retrying", "statusCode", statusCode)
			res.Body.Close()
			continue
		}

		body, err = io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			err = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		err = nil
		return
	}

	return nil, 0, fmt.Errorf("all %d retries failed: %w", maxRetry, err)
}

// NormalizeUrl canonicalizes a URL: normalizes scheme/host/query/path, filters disallowed paths/queries/extensions.
// Returns (normalized string, true) if valid for crawling; ("", false) if skipped or invalid.
func NormalizeUrl(raw, baseHost string) (string, bool) {
	disallowPaths := []string{
		"/login", "/logout", "/register", "/signup", "/password-reset",
		"/account/", "/cart", "/checkout", "/order/", "/payment/",
		"/search", "/filter/", "/admin/", "/dashboard/", "/settings/",
		"/404", "/error/", "/maintenance", "/test/", "/print/", "/preview/", "/tag/",
	}
	disallowQueries := []string{
		"sort", "page", "filter", "q", "search",
	}
	skipExtensions := []string{
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".zip", ".rar", ".7z", ".tar", ".gz", ".exe", ".msi", ".dmg", ".apk",
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".svg",
		".mp3", ".wav", ".aac", ".ogg", ".flac",
		".mp4", ".avi", ".mov", ".wmv", ".mkv", ".flv", ".webm",
		".css", ".js", ".ico",
	}
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	pathLower := strings.ToLower(u.Path)
	for _, ext := range skipExtensions {
		if strings.HasSuffix(pathLower, ext) {
			return "", false
		}
	}

	for _, dis := range disallowPaths {
		if u.Path == dis || strings.HasPrefix(u.Path, dis) {
			return "", false
		}
	}

	for _, param := range disallowQueries {
		if _, ok := u.Query()[param]; ok {
			return "", false
		}
	}

	u.Scheme = "https"
	baseHost = strings.ToLower(strings.TrimPrefix(baseHost, "www."))
	if u.Host == "" && baseHost != "" {
		u.Host = baseHost
	}
	if u.Host == "" {
		return "", false
	}
	u.Host = strings.TrimPrefix(u.Host, "www.")
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""
	if u.RawQuery != "" {
		q := u.Query()
		keys := make([]string, 0, len(q))
		for k := range q {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sorted := url.Values{}
		for _, k := range keys {
			sorted[k] = q[k]
		}
		u.RawQuery = sorted.Encode()
	}
	if len(u.Query()) == 0 && u.Path != "" && !strings.HasSuffix(u.Path, "/") &&
		!strings.Contains(u.Path, ".") {
		u.Path += "/"
	}

	return u.String(), true
}
