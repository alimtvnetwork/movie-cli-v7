// fallback.go — search fallbacks: progressive query trimming and IMDb-via-web lookup.
package tmdb

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FindResponse mirrors TMDb /find/{external_id} payload.
type FindResponse struct {
	MovieResults []SearchResult `json:"movie_results"`
	TVResults    []SearchResult `json:"tv_results"`
}

// SearchWithFallback tries SearchMulti first, then progressively trims trailing
// tokens from the title, and finally falls back to a web search → IMDb id →
// TMDb /find lookup. Returns the first non-empty result set or nil.
func (c *Client) SearchWithFallback(title string, year int) ([]SearchResult, error) {
	query := title
	if year > 0 {
		query = title + " " + strconv.Itoa(year)
	}
	if results, err := c.SearchMulti(query); err == nil && len(results) > 0 {
		return results, nil
	} else if err != nil && !isEmptyResultErr(err) {
		return nil, err
	}

	if results := c.tryProgressiveTrim(title, year); len(results) > 0 {
		return results, nil
	}

	if results := c.tryIMDbViaWeb(title, year); len(results) > 0 {
		return results, nil
	}

	return nil, nil
}

func isEmptyResultErr(err error) bool {
	// network / auth errors should bubble up; only "no results" is treated as empty.
	return false
}

// tryProgressiveTrim drops the last word from the title repeatedly until a
// match is found or the title is too short.
func (c *Client) tryProgressiveTrim(title string, year int) []SearchResult {
	words := strings.Fields(title)
	for n := len(words) - 1; n >= 2; n-- {
		shorter := strings.Join(words[:n], " ")
		query := shorter
		if year > 0 {
			query = shorter + " " + strconv.Itoa(year)
		}
		if results, err := c.SearchMulti(query); err == nil && len(results) > 0 {
			return results
		}
		if year > 0 {
			if results, err := c.SearchMulti(shorter); err == nil && len(results) > 0 {
				return results
			}
		}
	}
	return nil
}

var imdbIDPattern = regexp.MustCompile(`tt\d{7,10}`)

// tryIMDbViaWeb does a DuckDuckGo HTML search for "<title> <year> imdb",
// extracts the first IMDb id, and resolves it via TMDb /find. Cached when
// a Client.IMDbCache is attached.
func (c *Client) tryIMDbViaWeb(title string, year int) []SearchResult {
	imdbID := c.findIMDbIDViaWeb(title, year)
	if imdbID == "" {
		return nil
	}
	return c.lookupByIMDbID(imdbID)
}

// findIMDbIDViaWeb returns the cached id when available, otherwise fetches
// from DuckDuckGo and writes the result (hit or miss) back to the cache.
func (c *Client) findIMDbIDViaWeb(title string, year int) string {
	if c.IMDbCache != nil {
		if id, _, found := c.IMDbCache.Look(title, year); found {
			return id
		}
	}

	id := c.fetchIMDbIDFromDuckDuckGo(title, year)

	if c.IMDbCache != nil {
		_ = c.IMDbCache.Store(title, year, id)
	}
	return id
}

// fetchIMDbIDFromDuckDuckGo performs the actual HTTP scrape. Always hits the
// network; callers should consult the cache via findIMDbIDViaWeb instead.
func (c *Client) fetchIMDbIDFromDuckDuckGo(title string, year int) string {
	query := title + " imdb"
	if year > 0 {
		query = title + " " + strconv.Itoa(year) + " imdb"
	}
	searchURL := "https://duckduckgo.com/html/?q=" + url.QueryEscape(query)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	req, reqErr := http.NewRequest(http.MethodGet, searchURL, nil)
	if reqErr != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; movie-cli/1.0)")
	resp, getErr := httpClient.Do(req)
	if getErr != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if readErr != nil {
		return ""
	}
	return imdbIDPattern.FindString(string(body))
}

func (c *Client) lookupByIMDbID(imdbID string) []SearchResult {
	if !c.HasAuth() {
		return nil
	}
	params := url.Values{}
	params.Set("external_source", "imdb_id")
	var resp FindResponse
	if err := c.get(c.buildURL("/find/"+imdbID, params), &resp); err != nil {
		return nil
	}
	out := make([]SearchResult, 0, len(resp.MovieResults)+len(resp.TVResults))
	for i := range resp.MovieResults {
		resp.MovieResults[i].MediaType = "movie"
		out = append(out, resp.MovieResults[i])
	}
	for i := range resp.TVResults {
		resp.TVResults[i].MediaType = "tv"
		out = append(out, resp.TVResults[i])
	}
	return out
}
