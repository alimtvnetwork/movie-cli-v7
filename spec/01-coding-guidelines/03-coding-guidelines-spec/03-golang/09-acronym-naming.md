# 09 — Acronym Naming (MixedCaps Rule)

> **TL;DR**: In Go identifiers we write multi-letter acronyms as **MixedCaps**,
> not as their original all-caps form. `ImdbCache`, not `IMDbCache`. `ApiKey`,
> not `APIKey`. `HttpClient`, not `HTTPClient`.

This spec is the project-specific override of Effective Go's "initialism"
recommendation. We chose the *strict MixedCaps* form because it removes
ambiguity at rename time, satisfies `golangci-lint` consistency checks, and
matches the rest of our codebase that grew organically toward this style
(`db.SetImdbLookup`, `cmd.attachImdbCacheUnless`, etc.).

---

## 1. The Rule

For every identifier (type, field, method, function, constant, variable,
test name) where an acronym appears **at the start or in the middle** of a
CamelCase name:

| Acronym | ❌ Wrong | ✅ Right |
|---------|---------|---------|
| IMDb    | `IMDbCache`, `IMDbID` (mid-word), `LookupByIMDbID` | `ImdbCache`, `ImdbId`, `LookupByImdbId` |
| TMDb    | `TMDbClient`, `initTMDbClient` | `TmdbClient`, `initTmdbClient` |
| API     | `APIKey`, `APIClient`, `getAPIToken` | `ApiKey`, `ApiClient`, `getApiToken` |
| HTTP    | `HTTPClient`, `HTTPError`, `parseHTTPHeader` | `HttpClient`, `HttpError`, `parseHttpHeader` |
| URL     | `URLPattern`, `BaseURL` (mid-word) | `UrlPattern`, `BaseUrl` |
| JSON    | `JSONDecoder`, `parseJSONBody` | `JsonDecoder`, `parseJsonBody` |
| SQL     | `SQLBuilder`, `runSQLQuery` | `SqlBuilder`, `runSqlQuery` |
| HTML    | `HTMLEscape`, `renderHTMLPage` | `HtmlEscape`, `renderHtmlPage` |
| XML     | `XMLParser` | `XmlParser` |

The same rule applies whether the identifier is **exported** (`ApiKey`) or
**unexported** (`apiKey`).

---

## 2. Intentional Exceptions

The following are **NOT** violations and must NOT be renamed:

### 2.1 Trailing-initialism short locals

When an acronym is the **very last token** of a short-lived **unexported
local variable**, keep it in original all-caps form. This matches Go's
classic idiom and is easier on the eye:

```go
// ✅ ok
imdbID := "tt0123456"
tmdbID := 12345
imgURL := "https://..."
reqURL := buildURL(...)
posterURL := poster.Path

// ❌ wrong — name is exported / mid-word, not trailing-only
type Movie struct {
    IMDbID int  // should be ImdbId
}
```

### 2.2 Database column names

Already MixedCaps and used inside SQL string literals. Do **not** alias them
in Go code:

```go
// SQL string — column is the source of truth
db.Query(`SELECT ImdbId, TmdbId FROM Media`)
```

### 2.3 External contracts

Environment variables, config keys, JSON field tags, HTTP headers, and
user-facing CLI flags are external contracts and follow the upstream
convention:

```go
os.Getenv("TMDB_API_KEY")           // env vars: SCREAMING_SNAKE
db.GetConfig("TmdbApiKey")           // config keys: MixedCaps
req.Header.Set("X-API-Key", key)     // HTTP headers: per RFC
```

### 2.4 Prose

Comments, doc strings, README, and Markdown files refer to the **product**
name, not the identifier:

```go
// ✅ ok — prose
// HttpClient wraps Go's net/http client to add retries for the TMDb API.

// ❌ wrong — prose says one thing, identifier says another
// HTTPClient wraps Go's net/http client …
```

---

## 3. Enforcement

### 3.1 At write time

When you create a new identifier, scan it for these substrings and lower
the trailing letters:

```
IMDb → Imdb     TMDb → Tmdb     API  → Api      HTTP → Http
URL  → Url      JSON → Json     SQL  → Sql      HTML → Html
XML  → Xml
```

### 3.2 At review time

Run a one-liner sweep before committing Go changes:

```bash
grep -rn -E '\b(IMDb|TMDb|API|HTTP|URL|JSON|SQL|HTML|XML)[A-Z]' \
  --include='*.go' . \
  | grep -vE '\b(imdbID|tmdbID|imgURL|reqURL|posterURL)\b'
```

Any non-empty output (excluding the trailing-initialism allowlist in section
2.1) is a violation.

### 3.3 At rename time

Use `sed -i 's/\bOldName\b/NewName/g'` on every `*.go` file. Word boundaries
matter — without them you will corrupt comments and string literals.

---

## 4. History

| Date / Version | Change |
|---|---|
| v2.107.0 | First piecemeal rename: `attachIMDbCacheUnless` → `attachImdbCacheUnless` |
| v2.115.0 | Wholesale sweep across 17 Go files; this spec authored |

---

## 5. Acceptance Criteria

- GIVEN any new Go file WHEN linted THEN no exported or unexported identifier
  contains the substring `IMDb`, `TMDb`, `API`, `HTTP`, `URL`, `JSON`, `SQL`,
  `HTML`, or `XML` followed by another uppercase letter.
- GIVEN a local variable that ends in `ID` or `URL` AND is short-lived AND
  unexported THEN the trailing all-caps form is permitted.
- GIVEN a comment or doc string THEN the human-readable product name is
  preferred (`TMDb API`, `IMDb id`).

---

*Authored: 2026-04-18 — see also `mem://constraints/acronym-mixedcaps`.*
