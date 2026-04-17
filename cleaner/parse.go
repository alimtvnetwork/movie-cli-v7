// Package cleaner handles filename cleaning and metadata extraction from file names.
package cleaner

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// video file extensions
var videoExtensions = map[string]bool{
	".mkv": true, ".mp4": true, ".avi": true, ".mov": true,
	".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
	".ts": true, ".vob": true, ".ogv": true, ".mpg": true,
	".mpeg": true, ".3gp": true,
}

// patterns to remove from file names
var junkPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(1080p|720p|480p|2160p|4k|uhd|fhd|hd|sd)\b`),
	regexp.MustCompile(`(?i)\b(bluray|blu-ray|bdrip|brrip|bdremux|remux|webrip|web-dl|web\.dl|webdl|web|hdrip|dvdrip|dvdscr|dvd|hdtv|pdtv|hdcam|cam|ts|tc|telesync|telecine|workprint|screener)\b`),
	regexp.MustCompile(`(?i)\b(x264|x265|h\.?264|h\.?265|hevc|avc|xvid|divx|aac|ac3|dts(?:-?hd)?|dts-?ma|mp3|flac|atmos|ddp?|dd[px]?\s*[257]\s*[01]|ddp?5\.?1|eac3|truehd|lpcm|opus|vorbis|mp2)\b`),
	regexp.MustCompile(`(?i)\b(rarbg|yts|yify|eztv|ettv|sparks|geckos|fgt|evo|ion10|cmrg|ntb|megusta|immortal|msubs?|pahe|mkvcinemas?|tamilrockers|yts\.mx|jhs|ds4k|ds4k\.jhs|hq|hqcam|psa|qxr|kogi|deflate|ntg|edith|amiable|chd|wiki|fraternity|publichd|silence|tigole|joy|qx|axxo|dimension|lol|killers|fov|ds4k\b)\b`),
	regexp.MustCompile(`(?i)\b(extended|unrated|directors\.?cut|directors|remastered|proper|rerip|internal|repack|theatrical|imax|criterion\.?collection|anniversary|edition|complete|final\.?cut)\b`),
	regexp.MustCompile(`(?i)\b(multi|dual|eng|english|hindi|hin|spanish|spa|french|fra|german|ger|ita|italian|tamil|tam|telugu|tel|malayalam|mal|korean|kor|japanese|jap|chinese|chi|arabic|ara|bengali|ben|kannada|kan|marathi|mar|punjabi|pun|urdu|urd|portuguese|por|russian|rus|turkish|tur|vietnamese|vie|thai|polish|dutch|swedish|norwegian|danish|finnish)\b`),
	regexp.MustCompile(`(?i)\b(sub|subs|subbed|subtitle|subtitles|esub|esubs|e-sub|e-subs|msub|msubs|softcoded|hardcoded|softsub|hardsub)\b`),
	regexp.MustCompile(`(?i)\b(10bit|8bit|12bit|10-bit|8-bit|hdr|hdr10|hdr10\+?|hdr10plus|dv|dovi|dolby\.?vision|sdr|imax\.enhanced)\b`),
	regexp.MustCompile(`(?i)\b(amzn|nf|netflix|dsnp|disney|disney\+?|hmax|hbomax|hbo|hulu|atvp|apple|appletv|pcok|peacock|pmtp|paramount|paramount\+?|stan|crav|criterion|mubi|sho|showtime|starz|max)\b`),
	regexp.MustCompile(`(?i)\b([57]\s*1|[27]\s*0|5\.1|7\.1|2\.0)\b`),
	regexp.MustCompile(`\[.*?\]`),
	regexp.MustCompile(`\([^)]*\)`),
	regexp.MustCompile(`\{[^}]*\}`),
}

var yearInParens = regexp.MustCompile(`\((\d{4})\)`)
var yearPattern = regexp.MustCompile(`\b((?:19|20)\d{2})\b`)
var tvPattern = regexp.MustCompile(`(?i)S\d{1,2}E\d{1,2}|Season\s*\d+|Episode\s*\d+`)
var multiSpace = regexp.MustCompile(`\s{2,}`)
var trailingJunk = regexp.MustCompile(`\s*[-–—]+\s*$`)
var dashSeparator = regexp.MustCompile(`\s+[-–—]+\s+`)
var duplicateYear = regexp.MustCompile(`\b((?:19|20)\d{2})(?:\s+\1)+\b`)

// Result holds the cleaned information extracted from a filename.
type Result struct {
	CleanTitle string
	Extension  string
	Type       string // "movie" or "tv"
	Year       int
}

// IsVideoFile checks if the file has a video extension.
func IsVideoFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return videoExtensions[ext]
}

// Clean takes a raw filename and returns cleaned metadata.
func Clean(filename string) Result {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// Detect TV show
	mediaType := "movie"
	if tvPattern.MatchString(name) {
		mediaType = "tv"
	}

	// Extract year (prefer parenthesized year, fall back to bare year)
	year := 0
	if matches := yearInParens.FindStringSubmatch(name); len(matches) > 1 {
		year, _ = strconv.Atoi(matches[1])
	}
	if year == 0 {
		if matches := yearPattern.FindStringSubmatch(name); len(matches) > 1 {
			year, _ = strconv.Atoi(matches[1])
		}
	}

	// Replace dots and underscores with spaces
	cleaned := strings.ReplaceAll(name, ".", " ")
	cleaned = strings.ReplaceAll(cleaned, "_", " ")

	// Replace year in parens with bare year before junk removal
	cleaned = yearInParens.ReplaceAllString(cleaned, "$1")

	// Cut at first " - " separator if right side is mostly junk/release-group
	if parts := dashSeparator.Split(cleaned, 2); len(parts) == 2 {
		cleaned = cutAtDashKeepingTitle(parts[0], parts[1], year)
	}

	// Collapse duplicate years ("2025 2025" -> "2025")
	cleaned = duplicateYear.ReplaceAllString(cleaned, "$1")

	// Remove junk patterns (two passes to catch tokens revealed after first pass)
	for pass := 0; pass < 2; pass++ {
		for _, p := range junkPatterns {
			cleaned = p.ReplaceAllString(cleaned, "")
		}
	}

	// Clean up extra spaces
	cleaned = multiSpace.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	// If year exists, truncate everything after the year
	if year > 0 {
		yearStr := strconv.Itoa(year)
		if idx := strings.Index(cleaned, yearStr); idx > 0 {
			cleaned = strings.TrimSpace(cleaned[:idx])
		}
	}

	// Remove trailing dashes/hyphens left over
	cleaned = trailingJunk.ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)

	return Result{
		CleanTitle: cleaned,
		Year:       year,
		Extension:  strings.ToLower(ext),
		Type:       mediaType,
	}
}

// ToSlug converts a clean title to a slug (e.g., "Scream 2022" → "scream-2022").
func ToSlug(title string) string {
	s := strings.ToLower(title)
	s = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(s, "")
	s = multiSpace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

// ToCleanFileName creates a clean filename: "Scream (2022).mkv"
func ToCleanFileName(title string, year int, ext string) string {
	if year > 0 {
		return title + " (" + strconv.Itoa(year) + ")" + ext
	}
	return title + ext
}

// cutAtDashKeepingTitle decides whether to keep just the left side of a
// "Title - Suffix" split. If the suffix only contains junk tokens or just a
// year, drop it. Otherwise keep the original (left + " " + right) so we don't
// truncate legitimate subtitles like "Mad Max - Fury Road".
func cutAtDashKeepingTitle(left, right string, year int) string {
	stripped := right
	for _, p := range junkPatterns {
		stripped = p.ReplaceAllString(stripped, "")
	}
	if year > 0 {
		stripped = strings.ReplaceAll(stripped, strconv.Itoa(year), "")
	}
	stripped = multiSpace.ReplaceAllString(stripped, " ")
	stripped = strings.TrimSpace(stripped)
	if stripped == "" {
		return strings.TrimSpace(left)
	}
	return strings.TrimSpace(left) + " " + strings.TrimSpace(right)
}
