// movie_rest_export.go — CSV/JSON export endpoints for the dashboard.
//
// Endpoints:
//   GET /api/dashboard/export?format=csv|json&<dashboard list filters>
//   GET /api/media/{id}/similar/export?format=csv|json
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v6/db"
	"github.com/alimtvnetwork/movie-cli-v6/errlog"
	"github.com/alimtvnetwork/movie-cli-v6/tmdb"
)

const (
	exportFormatCSV  = "csv"
	exportFormatJSON = "json"
)

// ----- /api/dashboard/export ------------------------------------------------

func handleDashboardExport(w http.ResponseWriter, r *http.Request, database *db.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cards, err := loadFilteredCards(database, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeExport(w, r, cards, "library")
}

func loadFilteredCards(database *db.DB, r *http.Request) ([]dashboardCard, error) {
	q := parseDashboardListQuery(r)
	items, err := database.ListAllMedia()
	if err != nil {
		return nil, err
	}
	cards := buildDashboardCards(database, items)
	cards = applyDashboardFilters(cards, q)
	sortDashboardCards(cards, q.Sort)
	return cards, nil
}

func writeExport(w http.ResponseWriter, r *http.Request, cards []dashboardCard, baseName string) {
	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "" {
		format = exportFormatCSV
	}
	switch format {
	case exportFormatJSON:
		writeJSONExport(w, cards, baseName)
	case exportFormatCSV:
		writeCSVExport(w, cards, baseName)
	default:
		http.Error(w, "format must be csv or json", http.StatusBadRequest)
	}
}

func exportFilename(base, ext string) string {
	return fmt.Sprintf("%s-%s.%s", base, time.Now().UTC().Format("20060102-150405"), ext)
}

func writeJSONExport(w http.ResponseWriter, cards []dashboardCard, baseName string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="`+exportFilename(baseName, "json")+`"`)
	payload := map[string]interface{}{
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"count":       len(cards),
		"items":       cards,
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		errlog.Error("export json encode error: %v", err)
	}
}

func writeCSVExport(w http.ResponseWriter, cards []dashboardCard, baseName string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+exportFilename(baseName, "csv")+`"`)
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write(dashboardCSVHeader()); err != nil {
		errlog.Error("export csv header error: %v", err)
		return
	}
	for i := range cards {
		if err := cw.Write(dashboardCSVRow(&cards[i])); err != nil {
			errlog.Error("export csv row error: %v", err)
			return
		}
	}
}

func dashboardCSVHeader() []string {
	return []string{
		"id", "title", "year", "type", "runtime",
		"tmdb_id", "tmdb_rating", "genre", "director",
		"cast", "tagline", "description", "tags", "watched",
	}
}

func dashboardCSVRow(c *dashboardCard) []string {
	return []string{
		strconv.FormatInt(c.ID, 10), c.Title, strconv.Itoa(c.Year), c.Type,
		strconv.Itoa(c.Runtime), strconv.Itoa(c.TmdbID),
		strconv.FormatFloat(c.TmdbRating, 'f', 1, 64),
		c.Genre, c.Director, c.CastList, c.Tagline, c.Description,
		strings.Join(c.Tags, "|"), strconv.FormatBool(c.Watched),
	}
}

// ----- /api/media/{id}/similar/export --------------------------------------

func handleSimilarExport(w http.ResponseWriter, r *http.Request, database *db.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := parseSimilarExportID(r.URL.Path)
	if id <= 0 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	m, getErr := database.GetMediaByID(id)
	if getErr != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	results := loadSimilarOrEmpty(database, m)
	writeSimilarExport(w, r, m, results)
}

func parseSimilarExportID(urlPath string) int64 {
	// /api/media/{id}/similar/export
	path := strings.TrimPrefix(urlPath, "/api/media/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "similar" || parts[2] != "export" {
		return 0
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func loadSimilarOrEmpty(database *db.DB, m *db.Media) []tmdb.SearchResult {
	if m.TmdbID == 0 {
		return nil
	}
	return fetchSimilarFromTMDb(database, m)
}

func writeSimilarExport(w http.ResponseWriter, r *http.Request, m *db.Media, results []tmdb.SearchResult) {
	base := "similar-" + sanitizeFilename(m.Title)
	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "" {
		format = exportFormatCSV
	}
	switch format {
	case exportFormatJSON:
		writeSimilarJSON(w, m, results, base)
	case exportFormatCSV:
		writeSimilarCSV(w, results, base)
	default:
		http.Error(w, "format must be csv or json", http.StatusBadRequest)
	}
}

func writeSimilarJSON(w http.ResponseWriter, m *db.Media, results []tmdb.SearchResult, base string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="`+exportFilename(base, "json")+`"`)
	payload := map[string]interface{}{
		"source_id":    m.ID,
		"source_title": m.Title,
		"exported_at":  time.Now().UTC().Format(time.RFC3339),
		"count":        len(results),
		"items":        results,
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		errlog.Error("similar json encode error: %v", err)
	}
}

func writeSimilarCSV(w http.ResponseWriter, results []tmdb.SearchResult, base string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+exportFilename(base, "csv")+`"`)
	cw := csv.NewWriter(w)
	defer cw.Flush()
	_ = cw.Write([]string{"tmdb_id", "title", "year", "rating", "overview", "poster_path"})
	for i := range results {
		_ = cw.Write(similarCSVRow(&results[i]))
	}
}

func similarCSVRow(s *tmdb.SearchResult) []string {
	title := s.Title
	if title == "" {
		title = s.Name
	}
	year := firstFour(s.ReleaseDate)
	if year == "" {
		year = firstFour(s.FirstAir)
	}
	return []string{
		strconv.Itoa(s.ID), title, year,
		strconv.FormatFloat(s.VoteAvg, 'f', 1, 64),
		s.Overview, s.PosterPath,
	}
}

func firstFour(s string) string {
	if len(s) < 4 {
		return ""
	}
	return s[:4]
}

func sanitizeFilename(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out = append(out, r)
		case r == ' ', r == '-', r == '_':
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "item"
	}
	return string(out)
}
