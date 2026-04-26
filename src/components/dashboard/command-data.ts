/**
 * Single source of truth for every `movie` CLI command surfaced in the
 * dashboard. Used by:
 *   - CommandReferencePanel (grouped reference list)
 *   - CommandPalette       (Cmd/Ctrl+K search)
 *
 * The CLI is FLAT — every command starts with `movie ` and never `movie movie`.
 * See mem://constraints/command-syntax-flat.
 */

/** A complete, runnable invocation that demonstrates a command in context. */
export interface CommandExampleInvocation {
  cmd: string;
  note: string;
}

export interface CommandEntry {
  cmd: string;
  desc: string;
  group: string;
  /** README anchor for "more details" deep-link. */
  anchor: string;
  /** Keywords to widen fuzzy-search recall. */
  keywords?: string[];
  /** Real-world invocations the user can copy and run as-is. */
  examples?: CommandExampleInvocation[];
}

export const COMMAND_ENTRIES: CommandEntry[] = [
  // Setup & Info ----------------------------------------------------------
  { cmd: "movie hello", desc: "Greeting with version", group: "Setup & Info", anchor: "configuration--system" },
  { cmd: "movie version", desc: "Show version, commit, build date", group: "Setup & Info", anchor: "configuration--system", keywords: ["build", "commit"] },
  { cmd: "movie changelog", desc: "Display the changelog", group: "Setup & Info", anchor: "configuration--system" },
  { cmd: "movie update", desc: "Self-update to latest release", group: "Setup & Info", anchor: "configuration--system", keywords: ["upgrade"] },
  { cmd: "movie config", desc: "View configuration", group: "Setup & Info", anchor: "configuration--system" },
  { cmd: "movie config set tmdb_api_key YOUR_KEY", desc: "Set TMDb API key", group: "Setup & Info", anchor: "configuration--system", keywords: ["api", "key", "tmdb"],
    examples: [
      { cmd: "movie config set tmdb_api_key abcd1234efgh5678", note: "Replace with your real TMDb v3 API key" },
      { cmd: "movie config get tmdb_api_key", note: "Verify the key was stored" },
    ] },
  { cmd: "movie config set source_folder /path/to/media", desc: "Set scan source folder", group: "Setup & Info", anchor: "configuration--system", keywords: ["folder", "path"],
    examples: [
      { cmd: "movie config set source_folder /mnt/storage/Movies", note: "Linux / macOS absolute path" },
      { cmd: "movie config set source_folder D:\\Media\\Movies", note: "Windows path (escape backslashes in PowerShell)" },
    ] },

  // Library Management ---------------------------------------------------
  { cmd: "movie scan /path/to/media", desc: "Scan folder → DB + TMDb metadata", group: "Library Management", anchor: "scanning--library", keywords: ["index", "import"],
    examples: [
      { cmd: "movie scan /mnt/storage/Movies", note: "Scan a single root folder" },
      { cmd: "movie scan /mnt/storage/Movies --refresh", note: "Re-fetch TMDb metadata for existing entries" },
      { cmd: "movie scan . --dry-run", note: "Preview what would be added without writing" },
    ] },
  { cmd: "movie ls", desc: "Paginated library list", group: "Library Management", anchor: "scanning--library", keywords: ["list", "browse"],
    examples: [
      { cmd: "movie ls --genre Action", note: "Filter by a single genre" },
      { cmd: "movie ls --year 2020 --sort rating", note: "Combine filters and sort" },
      { cmd: "movie ls --limit 5", note: "Show only the top 5 rows" },
    ] },
  { cmd: "movie info 123", desc: "Show details for media ID 123", group: "Library Management", anchor: "scanning--library", keywords: ["details"],
    examples: [
      { cmd: "movie info 123 --json", note: "Machine-readable output for piping" },
    ] },
  { cmd: "movie search inception", desc: "Live TMDb search → save", group: "Library Management", anchor: "discovery--organization", keywords: ["find", "tmdb"],
    examples: [
      { cmd: "movie search \"the matrix\"", note: "Quote multi-word titles" },
      { cmd: "movie search inception --year 2010", note: "Disambiguate with release year" },
    ] },
  { cmd: "movie suggest", desc: "Recommendations & trending", group: "Library Management", anchor: "discovery--organization", keywords: ["recommend"],
    examples: [
      { cmd: "movie suggest --genre Drama --limit 10", note: "Pick 10 drama recommendations" },
    ] },
  { cmd: "movie stats", desc: "Library statistics + sizes", group: "Library Management", anchor: "maintenance--debugging", keywords: ["report"],
    examples: [
      { cmd: "movie stats --by genre", note: "Break down counts and size by genre" },
    ] },
  { cmd: "movie export", desc: "Export library data", group: "Library Management", anchor: "maintenance--debugging", keywords: ["dump", "csv", "json"],
    examples: [
      { cmd: "movie export --format csv --out library.csv", note: "Export the whole library to CSV" },
      { cmd: "movie export --format json --out library.json", note: "Export to JSON for tooling" },
    ] },

  // File Operations ------------------------------------------------------
  { cmd: "movie move", desc: "Browse & move files (interactive)", group: "File Operations", anchor: "file-management",
    examples: [
      { cmd: "movie move 123 --to /mnt/storage/Action", note: "Move a specific media ID to a target folder" },
    ] },
  { cmd: "movie move --all", desc: "Batch move all (cross-drive safe)", group: "File Operations", anchor: "file-management", keywords: ["batch"],
    examples: [
      { cmd: "movie move --all --to /mnt/storage/Sorted --dry-run", note: "Preview the full batch first" },
      { cmd: "movie move --all --to /mnt/storage/Sorted", note: "Execute the batch move" },
    ] },
  { cmd: "movie rename", desc: "Batch clean rename", group: "File Operations", anchor: "file-management",
    examples: [
      { cmd: "movie rename --all --pattern \"{title} ({year})\"", note: "Apply a TMDb-driven naming pattern" },
      { cmd: "movie rename 123 --dry-run", note: "Preview a single rename" },
    ] },
  { cmd: "movie undo", desc: "Revert last move/rename (with confirm)", group: "File Operations", anchor: "history--undo", keywords: ["revert", "rollback"],
    examples: [
      { cmd: "movie undo --list", note: "Show recent reversible operations" },
      { cmd: "movie undo --id 42", note: "Revert a specific history entry" },
    ] },
  { cmd: "movie play 123", desc: "Open in default player", group: "File Operations", anchor: "file-management", keywords: ["watch", "open"],
    examples: [
      { cmd: "movie play 123 --player vlc", note: "Force a specific player binary" },
    ] },
  { cmd: "movie duplicates", desc: "Detect duplicates", group: "File Operations", anchor: "maintenance--debugging", keywords: ["dupes"],
    examples: [
      { cmd: "movie duplicates --by hash", note: "Compare by file hash instead of title" },
    ] },
  { cmd: "movie cleanup", desc: "Remove stale DB entries", group: "File Operations", anchor: "maintenance--debugging", keywords: ["clean", "stale"],
    examples: [
      { cmd: "movie cleanup --dry-run", note: "List entries that would be removed" },
      { cmd: "movie cleanup --yes", note: "Skip confirmation prompt" },
    ] },

  // Tags & Watchlist -----------------------------------------------------
  { cmd: "movie tag add 1 favorite", desc: "Add tag to media", group: "Tags & Watchlist", anchor: "discovery--organization", keywords: ["label"],
    examples: [
      { cmd: "movie tag add 1 favorite rewatch", note: "Add multiple tags in one call" },
    ] },
  { cmd: "movie tag remove 1 favorite", desc: "Remove tag", group: "Tags & Watchlist", anchor: "discovery--organization",
    examples: [
      { cmd: "movie tag remove 1 --all", note: "Strip every tag from a media item" },
    ] },
  { cmd: "movie tag list 1", desc: "List tags for a media item", group: "Tags & Watchlist", anchor: "discovery--organization",
    examples: [
      { cmd: "movie tag list --all", note: "List every tag across the library" },
    ] },
  { cmd: "movie watch add 123", desc: "Add to watchlist", group: "Tags & Watchlist", anchor: "discovery--organization",
    examples: [
      { cmd: "movie watch add 123 --priority high", note: "Pin to the top of the watchlist" },
    ] },
  { cmd: "movie watch list", desc: "Show watchlist", group: "Tags & Watchlist", anchor: "discovery--organization",
    examples: [
      { cmd: "movie watch list --sort priority", note: "Order by priority instead of date added" },
    ] },
];

/** Stable list of group names in display order. */
export const COMMAND_GROUPS = Array.from(
  new Set(COMMAND_ENTRIES.map((e) => e.group)),
);