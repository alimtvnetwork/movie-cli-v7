// migrate_v3.go — V3: extend ImdbLookupCache with TmdbId + MediaType so a
// cached IMDb hit can short-circuit the TMDb /find call entirely.
//
// Background
// ----------
// V2 stored only the IMDb id resolved from DuckDuckGo. On every cache hit
// the fallback chain still hit TMDb /find?external_source=imdb_id to turn
// that IMDb id into a TMDb id + media type. By caching those two fields
// alongside the IMDb id we skip the TMDb network round-trip too: a fully
// "warm" cache hit returns a synthetic SearchResult{ID, MediaType} and lets
// the caller fetch full details from /movie/{id} or /tv/{id} as usual.
//
// SQLite ALTER TABLE only supports ADD COLUMN, which is exactly what we need
// here. NULL defaults keep the migration safe for existing rows — those rows
// will fall back to the V2 behaviour (one /find call) until they are
// re-resolved and the new columns are populated.
package db

// migrateV3 adds TmdbId (INTEGER) and MediaType (TEXT) to ImdbLookupCache.
func migrateV3(d *DB) error {
	stmts := []string{
		`ALTER TABLE ImdbLookupCache ADD COLUMN TmdbId INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE ImdbLookupCache ADD COLUMN MediaType TEXT NOT NULL DEFAULT ''`,
	}
	for _, stmt := range stmts {
		if _, err := d.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
