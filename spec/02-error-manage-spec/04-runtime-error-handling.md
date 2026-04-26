# Runtime Error Handling — TMDb, Database, Network

**Version:** 1.1.0  
**Updated:** 2026-04-26  
**Format:** Error scenarios with handling strategy and GIVEN/WHEN/THEN criteria

---

## Purpose

Define how the CLI handles runtime errors from external dependencies (TMDb API, SQLite database, filesystem, network). This spec ensures graceful degradation, clear user messaging, and consistent error recovery across all commands.

---

## 1. TMDb API Errors

### 1.1 Rate Limiting (HTTP 429)

TMDb enforces rate limits (~40 requests per 10 seconds for API v3).

**Current behavior:** No retry logic. The HTTP client returns the error and the command prints it.

**Required behavior:**

| Scenario | Strategy |
|----------|----------|
| Single 429 response | Retry after `Retry-After` header value (or 2s default) |
| 3 consecutive 429s | Abort with clear message: "TMDb rate limit exceeded — try again in X seconds" |
| Batch operations (scan) | Add 250ms delay between requests to stay under limits |

**Implementation pattern:**

```go
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
    maxRetries := 3
    for attempt := 0; attempt < maxRetries; attempt++ {
        resp, err := c.httpClient.Do(req)
        if err != nil {
            return nil, err
        }
        if resp.StatusCode != 429 {
            return resp, nil
        }
        resp.Body.Close()

        retryAfter := resp.Header.Get("Retry-After")
        delay := 2 * time.Second
        if secs, err := strconv.Atoi(retryAfter); err == nil {
            delay = time.Duration(secs) * time.Second
        }
        time.Sleep(delay)
    }
    return nil, fmt.Errorf("TMDb rate limit exceeded after %d retries", maxRetries)
}
```

**Acceptance Criteria:**

- GIVEN a 429 response from TMDb WHEN a request is made THEN the client retries after `Retry-After` seconds
- GIVEN 3 consecutive 429 responses WHEN retries are exhausted THEN a clear error message is shown
- GIVEN a scan of 50 files WHEN TMDb requests are made THEN a 250ms delay is inserted between requests

---

### 1.2 Authentication Errors (HTTP 401)

| Scenario | Message |
|----------|---------|
| Invalid API key | `❌ TMDb API key is invalid. Run: movie config set tmdb_api_key YOUR_KEY` |
| Missing API key | `❌ No TMDb API key configured. Run: movie config set tmdb_api_key YOUR_KEY` |

**Acceptance Criteria:**

- GIVEN an invalid API key WHEN any TMDb request returns 401 THEN the error message includes the fix command
- GIVEN no API key configured WHEN a TMDb-dependent command runs THEN the error is shown before any network request

---

### 1.3 Server Errors (HTTP 5xx)

| Scenario | Strategy |
|----------|----------|
| 500 Internal Server Error | Retry once after 3s, then fail |
| 502/503 Bad Gateway | Retry once after 5s, then fail |
| 504 Gateway Timeout | Retry once after 5s, then fail |

**Message:** `⚠️ TMDb is temporarily unavailable. Try again later.`

---

### 1.4 Network Timeout

The HTTP client has a 15-second timeout (`tmdb/client.go:32`).

| Scenario | Strategy |
|----------|----------|
| Timeout on search/details | Show: `⚠️ TMDb request timed out. Check your internet connection.` |
| Timeout on poster download | Skip poster, continue with metadata: `⚠️ Poster download timed out — skipping` |

**Acceptance Criteria:**

- GIVEN no internet connection WHEN `movie search` runs THEN a network error message is shown (not a panic)
- GIVEN a poster download times out WHEN `movie scan` runs THEN the scan continues without the poster

---

## 2. SQLite Database Errors

### 2.1 Database Locked (SQLITE_BUSY)

SQLite allows only one writer at a time. WAL mode helps but doesn't eliminate contention.

| Scenario | Strategy |
|----------|----------|
| Write blocked by another connection | Retry with exponential backoff: 100ms, 200ms, 400ms, 800ms, 1.6s |
| 5 retries exhausted | Fail with: `❌ Database is busy — another process may be using it. Try again.` |

**Implementation pattern:**

```go
// In db.Open() — set busy_timeout pragma
db.Exec("PRAGMA busy_timeout = 5000")  // 5 second busy timeout
```

SQLite's built-in `busy_timeout` pragma handles most contention automatically. The 5-second timeout covers typical concurrent access scenarios.

**Acceptance Criteria:**

- GIVEN two CLI instances running simultaneously WHEN both write to the DB THEN the second waits up to 5 seconds
- GIVEN the DB is locked for >5 seconds WHEN a write is attempted THEN a clear error message is shown

---

### 2.2 Database Corruption

| Scenario | Strategy |
|----------|----------|
| `SQLITE_CORRUPT` error | Show: `❌ Database appears corrupted. Run: movie db repair` (future command) |
| Missing database file | Auto-create with migrations (current behavior in `db.Open()`) |
| Read-only filesystem | Show: `❌ Cannot write to database — check file permissions` |

---

### 2.3 Migration Failures

| Scenario | Strategy |
|----------|----------|
| New migration fails | Roll back the transaction, show the error, continue with old schema |
| Schema version mismatch | Log warning, attempt migration, fail gracefully if incompatible |

**Acceptance Criteria:**

- GIVEN a new version with schema changes WHEN migration fails THEN the CLI continues with the existing schema and warns the user

---

## 3. Filesystem Errors

### 3.1 File Not Found

| Command | Scenario | Message |
|---------|----------|---------|
| `movie play <id>` | `current_file_path` doesn't exist | `❌ File not found: /path/to/file.mkv` |
| `movie undo` | Source file at `to_path` is missing | `❌ Cannot undo — file no longer exists at: /path/to/file.mkv` |
| `movie scan <dir>` | Directory doesn't exist | `❌ Directory not found: /path/to/dir` |

---

### 3.2 Permission Denied

| Scenario | Message |
|----------|---------|
| Cannot read scan directory | `❌ Permission denied: /path/to/dir` |
| Cannot write to destination | `❌ Cannot write to destination — check permissions: /path/to/dir` |
| Cannot create thumbnails dir | `⚠️ Cannot create thumbnail dir — skipping poster download` |

---

### 3.3 Disk Full

| Scenario | Strategy |
|----------|----------|
| Copy fails mid-transfer | Delete partial file, keep source intact, show: `❌ Disk full — move aborted, source file preserved` |
| Poster download fails | Skip poster, continue scan |
| DB write fails | Show error, suggest freeing disk space |

**Acceptance Criteria:**

- GIVEN a cross-device move WHEN the copy fails mid-transfer THEN the source file is NOT deleted and the partial destination is cleaned up

---

## 4. Offline Mode / Graceful Degradation

When the network is unavailable, commands should degrade gracefully:

| Command | Online Behavior | Offline Behavior |
|---------|----------------|-------------------|
| `movie scan` | Fetch TMDb metadata + poster | Scan files, insert with cleaned filename only, skip metadata |
| `movie search` | Search TMDb | `❌ Network required for TMDb search` |
| `movie info <title>` | DB lookup → TMDb fallback | DB lookup only, skip TMDb fallback |
| `movie suggest` | TMDb recommendations + trending | `❌ Network required for suggestions` |
| `movie ls` | List from DB | Works fully offline ✅ |
| `movie move` | Works locally | Works fully offline ✅ |
| `movie rename` | Works locally | Works fully offline ✅ |
| `movie undo` | Works locally | Works fully offline ✅ |
| `movie play` | Opens local file | Works fully offline ✅ |
| `movie stats` | Reads from DB | Works fully offline ✅ |
| `movie tag` | Reads/writes DB | Works fully offline ✅ |
| `movie export` | Reads from DB | Works fully offline ✅ |
| `movie config` | Reads/writes DB | Works fully offline ✅ |
| `movie update` | `git pull` + rebuild | `❌ Network required for update` |

**Detection pattern:**

```go
func isNetworkError(err error) bool {
    var netErr net.Error
    if errors.As(err, &netErr) {
        return true
    }
    var dnsErr *net.DNSError
    return errors.As(err, &dnsErr)
}
```

**Acceptance Criteria:**

- GIVEN no internet WHEN `movie scan ~/dir` runs THEN files are scanned and inserted with local data only (no metadata)
- GIVEN no internet WHEN `movie ls` runs THEN the library is displayed normally
- GIVEN no internet WHEN `movie search` runs THEN a clear "network required" message is shown

---

## 5. OMDb Fallback Errors

OMDb (`omdbapi.com`) is the **second** fallback tier in `SearchWithFallback`,
sitting between TMDb's progressive query trim and the IMDb-via-web scrape.
It is a best-effort enrichment source — failures here must NEVER block a
scan or surface as a user-facing error.

**Activation:** only runs when the `OMDB_API_KEY` environment variable is
set. When unset, the tier is silently skipped (this is the default state
for fresh installs and is not an error).

### 5.1 Missing API Key

| Scenario | Strategy |
|----------|----------|
| `OMDB_API_KEY` unset | Skip the OMDb tier silently — do NOT log, warn, or count as a failure |
| `OMDB_API_KEY` empty string | Treated identically to unset |
| `movie doctor` invoked | Report `OMDb fallback: disabled (set OMDB_API_KEY to enable)` as INFO, not WARN |

**Rationale:** OMDb is opt-in. Most users will never set the key, and the
TMDb + IMDb-scrape chain works fine without it.

### 5.2 Network / HTTP Errors

| Scenario | Strategy |
|----------|----------|
| DNS / TCP failure | Return `nil` from `tryOmdbFallback`; chain continues to IMDb-via-web |
| HTTP timeout (>8s) | Same — fail-soft, no retry, no log spam |
| HTTP 4xx (401/403 invalid key) | Return `nil`; emit a single `⚠️ OMDb auth failed — check OMDB_API_KEY` warning per process (deduplicated) |
| HTTP 5xx | Return `nil`; OMDb is unreliable, no retry — fall through to next tier |
| HTTP 429 (rate limit) | Return `nil`; OMDb's free tier is 1k/day — back off entirely for this process run |

**No retry policy.** Unlike TMDb, OMDb failures do NOT trigger retries.
The whole point of the OMDb tier is "cheap, fast, optional" — a slow retry
would defeat its purpose. The next fallback tier (IMDb scrape) takes over.

### 5.3 Empty / Invalid Response

| Scenario | Strategy |
|----------|----------|
| `Response: "False"` (no match) | Return `nil` — normal "no results", chain continues |
| Missing `imdbID` field | Return `nil` — treated as no result |
| Malformed JSON | Return `nil` — log at DEBUG only, never user-facing |
| `imdbID` present but TMDb `/find` returns nothing | Cache the partial hit (IMDb id, TmdbId=0) so the next run can retry `/find` without re-hitting OMDb |

**Acceptance Criteria:**

- GIVEN `OMDB_API_KEY` is unset WHEN `movie scan` runs THEN the OMDb tier is skipped without any log line
- GIVEN OMDb returns `Response: "False"` WHEN the fallback runs THEN the next tier (IMDb scrape) is invoked
- GIVEN OMDb returns HTTP 401 WHEN the fallback runs THEN exactly one `⚠️ OMDb auth failed` warning is printed per process
- GIVEN OMDb returns HTTP 5xx or times out WHEN the fallback runs THEN no retry is attempted and the chain continues
- GIVEN OMDb returns an `imdbID` that TMDb `/find` cannot resolve WHEN the fallback completes THEN the IMDb id is cached with `TmdbId=0` for future retry

---

## 6. IMDb-Scrape Fallback Errors

The IMDb-scrape tier (`tryImdbViaWeb` → `fetchImdbIdFromDuckDuckGo`) is the
**last-resort** fallback. It performs a DuckDuckGo HTML search for
`"<title> <year> imdb"`, extracts an `tt\d{7,10}` IMDb id from the result
page, then resolves it via TMDb `/find`. It is fragile by nature (HTML
layout can change, DuckDuckGo can rate-limit, results may be wrong) and
therefore has the strictest fail-soft policy of any tier.

### 6.1 Cache-First Behavior

| Scenario | Strategy |
|----------|----------|
| Title+year cached as a hit | Return cached TMDb id immediately — no network call |
| Title+year cached as a miss (`imdbID == ""`, `found == true`) | Return `nil` immediately — do NOT re-scrape |
| Title+year not in cache | Proceed to network scrape |

**Rationale:** the IMDb cache (`c.ImdbCache`) prevents repeated scrapes for
titles that are genuinely untrackable. Cached misses are honored for the
lifetime of the cache file.

### 6.2 Network / HTTP Errors

| Scenario | Strategy |
|----------|----------|
| DNS / TCP failure | Return `nil`; do NOT cache (transient, retry next run) |
| HTTP timeout (>10s) | Return `nil`; do NOT cache |
| HTTP non-200 (rate limit, captcha, block) | Return `nil`; do NOT cache so a future retry can succeed |
| Body read truncated at 512 KiB | Continue with partial body — IMDb id pattern usually appears early |

**No retry policy.** A single attempt per scan. Retry pressure on
DuckDuckGo would risk an IP ban that breaks the fallback for all users.

### 6.3 Parse / Resolution Errors

| Scenario | Strategy |
|----------|----------|
| No `tt\d{7,10}` match in HTML | Cache as miss (`imdbID=""`, `TmdbId=0`); return `nil` |
| IMDb id extracted but TMDb `/find` returns no results | Cache the partial hit (`imdbID=tt...`, `TmdbId=0`); return `nil` so a future `/find` retry can succeed |
| TMDb `/find` returns multiple results | Pick the first; cache the chosen `(TmdbId, MediaType)` |
| TMDb auth missing (no API key) | Return `nil` immediately — `lookupByImdbId` short-circuits via `HasAuth()` |

### 6.4 No-Results End-State

When **all** tiers (SearchMulti → progressive trim → OMDb → IMDb scrape)
return empty, the caller (`movie scan` / `movie search`) treats the title
as **unmatched**:

- Scan: file is recorded with `TmdbId=NULL`; user can re-run `movie rescan-failed` later
- Search: prints `📭 No matches found — try a shorter query or check spelling`
- Exit code: `0` (no-match is not an error condition for scan; search exits `0` too)

**Acceptance Criteria:**

- GIVEN a title is cached as a miss WHEN the IMDb-scrape tier runs THEN no HTTP request is made
- GIVEN DuckDuckGo returns HTTP 429 or a captcha page WHEN the scrape runs THEN no cache entry is written and the chain returns no results
- GIVEN the HTML contains no `tt\d{7,10}` pattern WHEN parsing completes THEN the title is cached as a miss
- GIVEN an IMDb id is found but TMDb `/find` returns nothing WHEN the chain completes THEN a partial cache entry (`imdbID`, `TmdbId=0`) is written
- GIVEN all four fallback tiers return no results WHEN `movie scan` finishes processing the file THEN the file is recorded with `TmdbId=NULL` and exit code is `0`
- GIVEN all four fallback tiers return no results WHEN `movie search` finishes THEN `📭 No matches found` is printed and exit code is `0`

---

## 7. Error Message Standards

All error messages follow these rules:

| Rule | Example |
|------|---------|
| Start with emoji indicator | `❌` error, `⚠️` warning, `📭` empty |
| Include the fix when possible | `❌ No API key. Run: movie config set tmdb_api_key YOUR_KEY` |
| Never expose stack traces to users | Log to debug, show human message |
| Print to stderr for errors | `fmt.Fprintf(os.Stderr, ...)` |
| Print to stdout for warnings | `fmt.Printf("⚠️ ...")` |

---

## Cross-References

- [Error Management Overview](./00-overview.md)
- [Error Architecture](./02-error-architecture/)
- [Error Code Registry](./03-error-code-registry/)
- [Acceptance Criteria](./97-acceptance-criteria.md)
- [TMDb Client](../../tmdb/client.go) — HTTP timeout, request logic
- [OMDb Fallback](../../tmdb/omdb.go) — OMDB_API_KEY, fail-soft policy
- [IMDb Scrape Fallback](../../tmdb/fallback.go) — DuckDuckGo → IMDb id → TMDb /find
- [DB Layer](../../db/db.go) — Open(), migrations, pragmas

---

*Runtime error handling spec — updated: 2026-04-26*
