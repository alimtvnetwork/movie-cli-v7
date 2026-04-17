// imdb_lookup_cache_admin.go — admin-facing helpers for the ImdbLookupCache.
//
// Used by the `movie cache imdb` command to inspect and invalidate cache rows
// without opening the SQLite file directly.
package db

import "database/sql"

// ImdbCacheEntry is one row from the ImdbLookupCache table, in display form.
type ImdbCacheEntry struct {
	LookupKey  string
	CleanTitle string
	Year       int
	ImdbID     string
	IsHit      bool
	LookedUpAt string
}

// ListImdbLookups returns every cached entry ordered by most recent first.
// Pass limit <= 0 for "all rows".
func (d *DB) ListImdbLookups(limit int) ([]ImdbCacheEntry, error) {
	rows, err := queryImdbCache(d, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanImdbCacheRows(rows)
}

func queryImdbCache(d *DB, limit int) (*sql.Rows, error) {
	base := `SELECT LookupKey, CleanTitle, Year, ImdbId, IsHit, LookedUpAt
	         FROM ImdbLookupCache
	         ORDER BY LookedUpAt DESC`
	if limit > 0 {
		return d.Query(base+" LIMIT ?", limit)
	}
	return d.Query(base)
}

func scanImdbCacheRows(rows *sql.Rows) ([]ImdbCacheEntry, error) {
	var out []ImdbCacheEntry
	for rows.Next() {
		var e ImdbCacheEntry
		if err := rows.Scan(&e.LookupKey, &e.CleanTitle, &e.Year, &e.ImdbID, &e.IsHit, &e.LookedUpAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CountImdbLookups returns (totalRows, hitRows). missRows = total - hits.
func (d *DB) CountImdbLookups() (int, int, error) {
	var total, hits int
	row := d.QueryRow(`SELECT COUNT(*), COALESCE(SUM(CASE WHEN IsHit THEN 1 ELSE 0 END), 0) FROM ImdbLookupCache`)
	if err := row.Scan(&total, &hits); err != nil {
		return 0, 0, err
	}
	return total, hits, nil
}

// ClearImdbLookups deletes every row from ImdbLookupCache. Returns row count removed.
func (d *DB) ClearImdbLookups() (int64, error) {
	res, err := d.Exec(`DELETE FROM ImdbLookupCache`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ClearImdbLookupMisses deletes only rows where IsHit = 0 (negative cache).
// Useful when you want to retry titles that previously failed without losing
// the long-lived hit cache.
func (d *DB) ClearImdbLookupMisses() (int64, error) {
	res, err := d.Exec(`DELETE FROM ImdbLookupCache WHERE IsHit = 0`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
