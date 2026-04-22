package cmd

import (
	"testing"

	"github.com/alimtvnetwork/movie-cli-v5/db"
)

func TestPathInScope(t *testing.T) {
	scope := normalizeScope("/movies/2024")

	cases := []struct {
		path string
		want bool
	}{
		{"/movies/2024/Inception/film.mkv", true},
		{"/movies/2024", true},
		{"/movies/2025/film.mkv", false},
		{"/movies/2024extra/film.mkv", false}, // prefix collision must NOT match
		{"", false},
	}
	for _, c := range cases {
		if got := pathInScope(c.path, scope); got != c.want {
			t.Errorf("pathInScope(%q) = %v, want %v", c.path, got, c.want)
		}
	}

	if !pathInScope("/anything", "") {
		t.Errorf("empty scope should match anything")
	}
}

func TestActionInScopeReadsSnapshotPaths(t *testing.T) {
	scope := normalizeScope("/movies/2024")
	a := db.ActionRecord{
		MediaSnapshot: `{"original_path":"/movies/2024/Junk","compact_path":"/movies/2024/.temp/Junk"}`,
	}
	if !ActionInScope(a, scope) {
		t.Fatalf("expected snapshot path to register as in-scope")
	}

	out := db.ActionRecord{
		MediaSnapshot: `{"original_path":"/movies/2025/Junk","compact_path":"/movies/2025/.temp/Junk"}`,
	}
	if ActionInScope(out, scope) {
		t.Fatalf("did not expect /movies/2025 snapshot to match /movies/2024 scope")
	}
}

func TestFilterMovesAndActions(t *testing.T) {
	scope := normalizeScope("/movies/2024")
	moves := []db.MoveRecord{
		{ID: 1, FromPath: "/movies/2024/a.mkv", ToPath: "/movies/2024/A/a.mkv"},
		{ID: 2, FromPath: "/movies/2025/b.mkv", ToPath: "/movies/2025/B/b.mkv"},
	}
	if got := FilterMoves(moves, scope); len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("FilterMoves wrong: %#v", got)
	}

	actions := []db.ActionRecord{
		{ActionHistoryId: 10, MediaSnapshot: `{"file_path":"/movies/2024/x.mkv"}`},
		{ActionHistoryId: 11, MediaSnapshot: `{"file_path":"/movies/2025/y.mkv"}`},
	}
	if got := FilterActions(actions, scope); len(got) != 1 || got[0].ActionHistoryId != 10 {
		t.Fatalf("FilterActions wrong: %#v", got)
	}
}