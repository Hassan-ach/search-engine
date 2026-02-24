package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

func GetReq(
	client *http.Client,
	url string,
	maxRetry, delay int,
) ([]byte, int, error) {
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

	var res *http.Response
	for attempt := range maxRetry {
		if attempt > 0 {
			time.Sleep(time.Second * time.Duration(delay))
		}

		res, err = client.Do(req)
		if err != nil {
			continue
		}

		statusCode := res.StatusCode
		if statusCode >= 500 || statusCode == 429 {
			res.Body.Close()
			continue
		}
		if statusCode >= 400 {
			return nil, statusCode, fmt.Errorf("client error: %d", statusCode)
		}

		body, err := io.ReadAll(io.LimitReader(res.Body, 10<<20)) // 10 MB
		res.Body.Close()
		if err != nil {
			err = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		err = nil
		return body, statusCode, nil
	}

	return nil, 0, fmt.Errorf("all %d retries failed: %w", maxRetry, err)
}

var (
	disallowPathPrefixes = []string{
		"/login", "/logout", "/register", "/signup", "/password-reset",
		"/account/", "/cart", "/checkout", "/order/", "/payment/",
		"/search", "/filter/", "/admin/", "/dashboard/", "/settings/",
		"/404", "/error/", "/maintenance", "/test/", "/print/", "/preview/", "/tag/",
	}

	disallowQueryParams = []string{
		"sort", "page", "filter", "q", "search",
	}

	skipFileExtensions = []string{
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".zip", ".rar", ".7z", ".tar", ".gz", ".exe", ".msi", ".dmg", ".apk",
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".svg",
		".mp3", ".wav", ".aac", ".ogg", ".flac",
		".mp4", ".avi", ".mov", ".wmv", ".mkv", ".flv", ".webm",
		".css", ".js", ".ico",
	}

	// Wiki-specific: blocks /Template:Foo/xx/, /Help:Bar/en/, etc.
	wikiLangSubpageRE   = regexp.MustCompile(`(?i)/[a-z]{2,3}(-[a-z]{2,4})?/?$`)
	wikiNoisyNamespaces = []string{"/Template:", "/Help:", "/Manual:", "/Extension:"}
)

func NormalizeUrl(raw string, baseHost string) (string, bool) {
	raw = sanitizeUTF8(raw)
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", false
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	if shouldSkipByPath(u.Path) {
		return "", false
	}

	if shouldSkipByExtension(u.Path) {
		return "", false
	}

	if shouldSkipByQuery(u.Query()) {
		return "", false
	}

	if shouldSkipWikiLanguageSubpage(u.Path) {
		return "", false
	}

	normalizeURLParts(u, baseHost)

	return u.String(), true
}

func shouldSkipByPath(path string) bool {
	for _, prefix := range disallowPathPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func shouldSkipByExtension(path string) bool {
	pathLower := strings.ToLower(path)
	for _, ext := range skipFileExtensions {
		if strings.HasSuffix(pathLower, ext) {
			return true
		}
	}
	return false
}

func shouldSkipByQuery(q url.Values) bool {
	for _, param := range disallowQueryParams {
		if _, exists := q[param]; exists {
			return true
		}
	}
	return false
}

func shouldSkipWikiLanguageSubpage(path string) bool {
	if !wikiLangSubpageRE.MatchString(path) {
		return false
	}
	for _, ns := range wikiNoisyNamespaces {
		if strings.Contains(path, ns) {
			return true
		}
	}
	return false
}

func normalizeURLParts(u *url.URL, baseHost string) {
	// Force HTTPS
	u.Scheme = "https"

	// Fill host if relative URL
	if u.Host == "" && baseHost != "" {
		u.Host = baseHost
	}

	// Clean host
	u.Host = strings.TrimPrefix(strings.ToLower(u.Host), "www.")

	// Remove fragment
	u.Fragment = ""

	// Canonicalize query: sorted keys
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

	// Add trailing slash to directory-like paths (no extension, no query)
	if len(u.Query()) == 0 &&
		u.Path != "" &&
		!strings.HasSuffix(u.Path, "/") &&
		!strings.Contains(u.Path, ".") {
		u.Path += "/"
	}
}

func NormalizeUrls(raws []string, baseHost string) []string {
	result := make([]string, 0, len(raws))
	for _, raw := range raws {
		if norm, ok := NormalizeUrl(raw, baseHost); ok {
			result = append(result, norm)
		}
	}
	return result
}

func sanitizeUTF8(s string) string {
	b := []byte(s)
	if utf8.Valid(b) {
		return s
	}

	var runes []rune
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r == utf8.RuneError && size == 1 {
			b = b[1:] // skip invalid byte
			continue
		}
		runes = append(runes, r)
		b = b[size:]
	}
	return string(runes)
}
