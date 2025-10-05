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

func UrlClean(url string) *url.URL {
	return nil
}

func GetReq(url string, MaxRetry int) (body []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.118 Safari/537.36",
	)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US")

	if err != nil {
		fmt.Printf("Request initialization failed\n ERROR: %v\n", err)
		return nil, err
	}
	for attempt := 0; attempt < MaxRetry || err != nil; attempt++ {
		if attempt != 0 {
			time.Sleep(5 * time.Second)
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("retry %d failed\nURL: %s\nERROR: %v\n", attempt+1, url, err)
			continue
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if res.StatusCode > 299 {
			fmt.Printf(
				"Response failed with status code: %d and\n body: %s\n",
				res.StatusCode,
				body,
			)
			continue
		}
		if err != nil {
			fmt.Printf("In GetReq error while reading the response body:  %v\n", err)
			continue
		}
		return body, nil
	}
	return nil, err
}

func Filter[T any](s []T, f func(T) bool) []T {
	var r []T
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

// NormalizeUrl canonicalizes a URL: normalizes scheme/host/query/path, filters disallowed paths/queries/extensions.
// Returns (normalized string, true) if valid for crawling; ("", false) if skipped or invalid.
func NormalizeUrl(raw, baseHost string) (string, bool) {
	disallowPaths := []string{ // Consider making []string param for config
		"/login", "/logout", "/register", "/signup", "/password-reset",
		"/account/", "/cart", "/checkout", "/order/", "/payment/",
		"/search", "/filter/", "/admin/", "/dashboard/", "/settings/",
		"/404", "/error/", "/maintenance", "/test/", "/print/", "/preview/", "/tag/",
	}
	disallowQueries := []string{ // Param names only
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

	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	// Skip extensions
	pathLower := strings.ToLower(u.Path)
	for _, ext := range skipExtensions {
		if strings.HasSuffix(pathLower, ext) {
			return "", false
		}
	}

	// Skip disallowed paths (prefix match only for flexibility)
	for _, dis := range disallowPaths {
		if u.Path == dis || strings.HasPrefix(u.Path, dis) {
			return "", false
		}
	}

	// Skip disallowed query params
	for _, param := range disallowQueries {
		if _, ok := u.Query()[param]; ok {
			return "", false
		}
	}

	// Defaults and normalization
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" && baseHost != "" {
		baseHost = strings.ToLower(strings.TrimPrefix(baseHost, "www."))
		u.Host = baseHost
	}
	if u.Host == "" { // Still empty? Invalid
		return "", false
	}
	if strings.HasPrefix(u.Host, "www.") {
		u.Host = u.Host[4:]
	}
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
			sorted[k] = q[k] // Preserves multi-values
		}
		u.RawQuery = sorted.Encode()
	}
	// Trailing / heuristic
	if u.Path != "" && !strings.HasSuffix(u.Path, "/") && !strings.Contains(u.Path, ".") {
		u.Path += "/"
	}

	return u.String(), true
}
