// migrate_v2.go — V2: ImdbLookupCache table for caching DuckDuckGo→IMDb lookups.
//
// The fallback chain in tmdb.SearchWithFallback ends with a DuckDuckGo HTML
// search to find the IMDb id for a title. That lookup is slow, network-bound,
// and rate-limited by the search engine. We persist every result (hit OR miss)
// keyed by lowercase clean title + year so that repeat runs of `movie scan`
// and `movie rescan-failed` reuse the cached id instead of re-hitting the web.
package db

// migrateV2 creates the ImdbLookupCache table.
func migrateV2(d *DB) error {
	_, err := d.Exec(`
	CREATE TABLE IF NOT EXISTS ImdbLookupCache (
		ImdbLookupCacheId INTEGER PRIMARY KEY AUTOINCREMENT,
		LookupKey         TEXT NOT NULL UNIQUE,
		CleanTitle        TEXT NOT NULL,
		Year              INTEGER NOT NULL DEFAULT 0,
		ImdbId            TEXT NOT NULL DEFAULT '',
		IsHit             BOOLEAN NOT NULL DEFAULT 0,
		LookedUpAt        TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_imdb_lookup_cache_lookedup
		ON ImdbLookupCache(LookedUpAt);
	`)
	return err
}
