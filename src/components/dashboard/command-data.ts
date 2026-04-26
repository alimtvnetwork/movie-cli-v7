/**
 * Single source of truth for every `movie` CLI command surfaced in the
 * dashboard. Used by:
 *   - CommandReferencePanel (grouped reference list)
 *   - CommandPalette       (Cmd/Ctrl+K search)
 *
 * The CLI is FLAT — every command starts with `movie ` and never `movie movie`.
 * See mem://constraints/command-syntax-flat.
 */

export interface CommandEntry {
  cmd: string;
  desc: string;
  group: string;
  /** README anchor for "more details" deep-link. */
  anchor: string;
  /** Keywords to widen fuzzy-search recall. */
  keywords?: string[];
}

export const COMMAND_ENTRIES: CommandEntry[] = [
  // Setup & Info ----------------------------------------------------------
  { cmd: "movie hello", desc: "Greeting with version", group: "Setup & Info", anchor: "configuration--system" },
  { cmd: "movie version", desc: "Show version, commit, build date", group: "Setup & Info", anchor: "configuration--system", keywords: ["build", "commit"] },
  { cmd: "movie changelog", desc: "Display the changelog", group: "Setup & Info", anchor: "configuration--system" },
  { cmd: "movie update", desc: "Self-update to latest release", group: "Setup & Info", anchor: "configuration--system", keywords: ["upgrade"] },
  { cmd: "movie config", desc: "View configuration", group: "Setup & Info", anchor: "configuration--system" },
  { cmd: "movie config set tmdb_api_key YOUR_KEY", desc: "Set TMDb API key", group: "Setup & Info", anchor: "configuration--system", keywords: ["api", "key", "tmdb"] },
  { cmd: "movie config set source_folder /path/to/media", desc: "Set scan source folder", group: "Setup & Info", anchor: "configuration--system", keywords: ["folder", "path"] },

  // Library Management ---------------------------------------------------
  { cmd: "movie scan /path/to/media", desc: "Scan folder → DB + TMDb metadata", group: "Library Management", anchor: "scanning--library", keywords: ["index", "import"] },
  { cmd: "movie ls", desc: "Paginated library list", group: "Library Management", anchor: "scanning--library", keywords: ["list", "browse"] },
  { cmd: "movie info 123", desc: "Show details for media ID 123", group: "Library Management", anchor: "scanning--library", keywords: ["details"] },
  { cmd: "movie search inception", desc: "Live TMDb search → save", group: "Library Management", anchor: "discovery--organization", keywords: ["find", "tmdb"] },
  { cmd: "movie suggest", desc: "Recommendations & trending", group: "Library Management", anchor: "discovery--organization", keywords: ["recommend"] },
  { cmd: "movie stats", desc: "Library statistics + sizes", group: "Library Management", anchor: "maintenance--debugging", keywords: ["report"] },
  { cmd: "movie export", desc: "Export library data", group: "Library Management", anchor: "maintenance--debugging", keywords: ["dump", "csv", "json"] },

  // File Operations ------------------------------------------------------
  { cmd: "movie move", desc: "Browse & move files (interactive)", group: "File Operations", anchor: "file-management" },
  { cmd: "movie move --all", desc: "Batch move all (cross-drive safe)", group: "File Operations", anchor: "file-management", keywords: ["batch"] },
  { cmd: "movie rename", desc: "Batch clean rename", group: "File Operations", anchor: "file-management" },
  { cmd: "movie undo", desc: "Revert last move/rename (with confirm)", group: "File Operations", anchor: "history--undo", keywords: ["revert", "rollback"] },
  { cmd: "movie play 123", desc: "Open in default player", group: "File Operations", anchor: "file-management", keywords: ["watch", "open"] },
  { cmd: "movie duplicates", desc: "Detect duplicates", group: "File Operations", anchor: "maintenance--debugging", keywords: ["dupes"] },
  { cmd: "movie cleanup", desc: "Remove stale DB entries", group: "File Operations", anchor: "maintenance--debugging", keywords: ["clean", "stale"] },

  // Tags & Watchlist -----------------------------------------------------
  { cmd: "movie tag add 1 favorite", desc: "Add tag to media", group: "Tags & Watchlist", anchor: "discovery--organization", keywords: ["label"] },
  { cmd: "movie tag remove 1 favorite", desc: "Remove tag", group: "Tags & Watchlist", anchor: "discovery--organization" },
  { cmd: "movie tag list 1", desc: "List tags for a media item", group: "Tags & Watchlist", anchor: "discovery--organization" },
  { cmd: "movie watch add 123", desc: "Add to watchlist", group: "Tags & Watchlist", anchor: "discovery--organization" },
  { cmd: "movie watch list", desc: "Show watchlist", group: "Tags & Watchlist", anchor: "discovery--organization" },
];

/** Stable list of group names in display order. */
export const COMMAND_GROUPS = Array.from(
  new Set(COMMAND_ENTRIES.map((e) => e.group)),
);