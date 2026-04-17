// imdb_cache_adapter.go — adapts *db.DB to the tmdb.IMDbCache interface.
//
// Lives in cmd/ so the db package stays independent of tmdb and vice-versa.
// Wire it up at every TMDb client construction site that has a *db.DB in scope:
//
//	client := tmdb.NewClientWithToken(...)
//	client.SetIMDbCache(newIMDbCacheAdapter(database))
package cmd

import (
	"github.com/alimtvnetwork/movie-cli-v5/db"
	"github.com/alimtvnetwork/movie-cli-v5/errlog"
)

// imdbCacheAdapter wraps *db.DB so it satisfies tmdb.IMDbCache.
type imdbCacheAdapter struct {
	database *db.DB
}

func newIMDbCacheAdapter(database *db.DB) *imdbCacheAdapter {
	if database == nil {
		return nil
	}
	return &imdbCacheAdapter{database: database}
}

// Look returns the cached lookup, swallowing DB errors so a broken cache
// degrades gracefully into a fresh web call rather than failing the search.
func (a *imdbCacheAdapter) Look(cleanTitle string, year int) (string, bool, bool) {
	if a == nil || a.database == nil {
		return "", false, false
	}
	res, err := a.database.GetImdbLookup(cleanTitle, year)
	if err != nil {
		errlog.Warn("imdb cache lookup failed for '%s' (%d): %v", cleanTitle, year, err)
		return "", false, false
	}
	return res.ImdbID, res.IsHit, res.Found
}

// Store records a hit or miss; errors are logged and swallowed.
func (a *imdbCacheAdapter) Store(cleanTitle string, year int, imdbID string) error {
	if a == nil || a.database == nil {
		return nil
	}
	if err := a.database.SetImdbLookup(cleanTitle, year, imdbID); err != nil {
		errlog.Warn("imdb cache store failed for '%s' (%d): %v", cleanTitle, year, err)
		return err
	}
	return nil
}
