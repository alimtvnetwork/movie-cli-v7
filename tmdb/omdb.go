// omdb.go — OMDb (omdbapi.com) fallback search.
//
// Used by SearchWithFallback as a secondary source when TMDb returns no
// results. OMDb maps a title (and optional year) to an IMDb id, which we
// then resolve back to a TMDb record via the existing /find endpoint.
//
// Auth: requires the OMDB_API_KEY environment variable. If unset, the
// fallback is silently skipped — no secrets are stored in the repo.
package tmdb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// omdbAPIKeyEnv is the environment variable that holds the user's OMDb key.
// Users obtain a free key at https://www.omdbapi.com/apikey.aspx and export
// it before running the CLI:  export OMDB_API_KEY=xxxxxxxx
const omdbAPIKeyEnv = "OMDB_API_KEY"

// omdbBaseURL is the OMDb HTTPS endpoint.
const omdbBaseURL = "https://www.omdbapi.com/"

// omdbHTTPTimeout caps OMDb requests so a slow upstream cannot stall scans.
const omdbHTTPTimeout = 8 * time.Second

// omdbResponse mirrors the subset of the OMDb payload we consume.
type omdbResponse struct {
	Response string `json:"Response"` // "True" or "False"
	Error    string `json:"Error,omitempty"`
	ImdbID   string `json:"imdbID"`
	Type     string `json:"Type"` // "movie" | "series" | "episode"
}

// omdbAPIKey reads the OMDb key from the environment. Empty string means
// the fallback is disabled.
// SHARED: used by tryOmdbFallback and HasOmdb so the key source stays in
// one place.
func omdbAPIKey() string {
	return os.Getenv(omdbAPIKeyEnv)
}

// HasOmdb reports whether an OMDB_API_KEY is configured. Exported so callers
// (e.g. `movie doctor` or status output) can advertise the fallback state
// without poking env vars themselves.
func HasOmdb() bool {
	return omdbAPIKey() != ""
}

// tryOmdbFallback queries OMDb for the title (+ optional year) and, if it
// returns an IMDb id, resolves it back to TMDb via /find. Returns nil on
// any failure — OMDb is best-effort, never fatal.
func (c *Client) tryOmdbFallback(title string, year int) []SearchResult {
	key := omdbAPIKey()
	if key == "" {
		return nil
	}
	imdbID := fetchOmdbImdbID(key, title, year)
	if imdbID == "" {
		return nil
	}
	results := c.lookupByImdbId(imdbID)
	if len(results) > 0 {
		c.storeImdbCache(title, year, imdbID, results[0].ID, results[0].MediaType)
		return results
	}
	// IMDb id known but TMDb /find could not resolve it — cache the partial
	// hit so a future run can retry the /find without re-scraping.
	c.storeImdbCache(title, year, imdbID, 0, "")
	return nil
}

// fetchOmdbImdbID performs the actual OMDb HTTP lookup. Pure function (no
// receiver) so it can be unit-tested without a Client.
func fetchOmdbImdbID(apiKey, title string, year int) string {
	params := url.Values{}
	params.Set("apikey", apiKey)
	params.Set("t", title)
	params.Set("type", "")
	if year > 0 {
		params.Set("y", strconv.Itoa(year))
	}
	reqURL := omdbBaseURL + "?" + params.Encode()

	httpClient := &http.Client{Timeout: omdbHTTPTimeout}
	req, reqErr := http.NewRequest(http.MethodGet, reqURL, nil)
	if reqErr != nil {
		return ""
	}
	req.Header.Set("User-Agent", "movie-cli/omdb-fallback")
	resp, doErr := httpClient.Do(req)
	if doErr != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if readErr != nil {
		return ""
	}
	var parsed omdbResponse
	if jsonErr := json.Unmarshal(body, &parsed); jsonErr != nil {
		return ""
	}
	if parsed.Response != "True" || parsed.ImdbID == "" {
		return ""
	}
	return parsed.ImdbID
}