import { useMemo, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { FileText, ChevronDown, ChevronUp, Save, Copy, Check, SlidersHorizontal, ShieldCheck } from "lucide-react";
import type { MediaItem } from "@/types/media";
import { formatFileSize } from "@/lib/media-utils";
import { toast } from "sonner";
import { ReadmeDiffDialog } from "./ReadmeDiffDialog";
import { CommandReferencePanel } from "./CommandReferencePanel";

interface ReadmePreviewProps {
  media: MediaItem[];
}

export interface ReadmeSections {
  overview: boolean;
  topGenres: boolean;
  topRated: boolean;
  allGenres: boolean;
  movies: boolean;
  tv: boolean;
}

const DEFAULT_SECTIONS: ReadmeSections = {
  overview: true,
  topGenres: true,
  topRated: true,
  allGenres: true,
  movies: true,
  tv: true,
};

const SECTION_LABELS: { key: keyof ReadmeSections; label: string; icon: string }[] = [
  { key: "overview", label: "Overview", icon: "📊" },
  { key: "topGenres", label: "Top Genres", icon: "🔥" },
  { key: "topRated", label: "Top Rated", icon: "🏆" },
  { key: "allGenres", label: "All Genres", icon: "🏷️" },
  { key: "movies", label: "Movies", icon: "🎞️" },
  { key: "tv", label: "TV Shows", icon: "📺" },
];

function buildReadmeContent(media: MediaItem[], sections: ReadmeSections): string {
  const movies = media.filter((m) => m.type === "movie");
  const tv = media.filter((m) => m.type === "tv");
  const totalSize = media.reduce((sum, m) => sum + (m.fileSize ?? 0), 0);

  const rated = media.filter((m) => (m.tmdbRating ?? 0) > 0);
  const avgRating =
    rated.length > 0
      ? (rated.reduce((sum, m) => sum + (m.tmdbRating ?? 0), 0) / rated.length).toFixed(1)
      : "0.0";

  const years = media.map((m) => m.year ?? 0).filter((y) => y > 0);
  const minYear = years.length > 0 ? Math.min(...years) : 0;
  const maxYear = years.length > 0 ? Math.max(...years) : 0;

  const genreSet = new Set<string>();
  media.forEach((m) => m.genres?.forEach((g) => genreSet.add(g)));
  const genres = Array.from(genreSet).sort();

  const genreCounts = new Map<string, number>();
  media.forEach((m) =>
    m.genres?.forEach((g) => genreCounts.set(g, (genreCounts.get(g) ?? 0) + 1)),
  );
  const topGenres = Array.from(genreCounts.entries())
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5);

  const topRated = [...rated]
    .sort((a, b) => (b.tmdbRating ?? 0) - (a.tmdbRating ?? 0))
    .slice(0, 10);

  const generatedAt = new Date().toISOString().replace("T", " ").split(".")[0] + " UTC";

  const lines: string[] = [];
  lines.push("# 🎬 Media Library");
  lines.push("");
  lines.push(`> Auto-generated catalogue of the local media collection.`);
  lines.push(`> _Last updated: ${generatedAt}_`);
  lines.push("");
  lines.push("---");
  lines.push("");

  if (sections.overview) {
    lines.push("## 📊 Overview");
    lines.push("");
    lines.push("| Metric | Value |");
    lines.push("| --- | --- |");
    lines.push(`| Total items | **${media.length}** |`);
    lines.push(`| 🎞️ Movies | ${movies.length} |`);
    lines.push(`| 📺 TV Shows | ${tv.length} |`);
    lines.push(`| 💾 Total size | ${formatFileSize(totalSize)} |`);
    lines.push(`| ⭐ Average rating | ${avgRating} / 10 |`);
    lines.push(`| 🏷️ Unique genres | ${genres.length} |`);
    if (minYear > 0) {
      lines.push(`| 📅 Year range | ${minYear} – ${maxYear} |`);
    }
    lines.push("");
  }

  if (sections.topGenres && topGenres.length > 0) {
    lines.push("## 🔥 Top Genres");
    lines.push("");
    topGenres.forEach(([g, count], i) => {
      lines.push(`${i + 1}. **${g}** — ${count} title${count === 1 ? "" : "s"}`);
    });
    lines.push("");
  }

  if (sections.topRated && topRated.length > 0) {
    lines.push("## 🏆 Top Rated");
    lines.push("");
    lines.push("| # | Title | Year | Rating |");
    lines.push("| --- | --- | --- | --- |");
    topRated.forEach((m, i) => {
      lines.push(
        `| ${i + 1} | ${m.title} | ${m.year ?? "—"} | ⭐ ${(m.tmdbRating ?? 0).toFixed(1)} |`,
      );
    });
    lines.push("");
  }

  if (sections.allGenres && genres.length > 0) {
    lines.push("## 🏷️ All Genres");
    lines.push("");
    lines.push(genres.map((g) => `\`${g}\``).join(" · "));
    lines.push("");
  }

  if (sections.movies && movies.length > 0) {
    lines.push("## 🎞️ Movies");
    lines.push("");
    movies
      .slice()
      .sort((a, b) => a.title.localeCompare(b.title))
      .forEach((m) => {
        const rating = m.tmdbRating ? ` — ⭐ ${m.tmdbRating.toFixed(1)}` : "";
        const year = m.year ? ` (${m.year})` : "";
        lines.push(`- **${m.title}**${year}${rating}`);
      });
    lines.push("");
  }

  if (sections.tv && tv.length > 0) {
    lines.push("## 📺 TV Shows");
    lines.push("");
    tv.slice()
      .sort((a, b) => a.title.localeCompare(b.title))
      .forEach((m) => {
        const rating = m.tmdbRating ? ` — ⭐ ${m.tmdbRating.toFixed(1)}` : "";
        const year = m.year ? ` (${m.year})` : "";
        lines.push(`- **${m.title}**${year}${rating}`);
      });
    lines.push("");
  }

  lines.push("---");
  lines.push("");
  lines.push(`_Generated by Media Library · ${media.length} items indexed_`);
  lines.push("");
  return lines.join("\n");
}

/**
 * Sanitizes README content by collapsing any accidental `movie movie ` (or
 * repeated `movie movie movie ...`) command prefixes back into a single
 * `movie `. The CLI command tree is FLAT — `movie <cmd>` only — so any
 * nested form is always a bug. See mem://constraints/command-syntax-flat.
 *
 * Returns the cleaned text and a count of how many occurrences were fixed.
 */
export function sanitizeMovieCommandPrefix(input: string): {
  text: string;
  replacements: number;
} {
  let replacements = 0;
  // Collapse one or more leading `movie ` repeats followed by another `movie `
  // into a single `movie `. Word-boundaries prevent matching things like
  // `moviemovie` or `movies`.
  const text = input.replace(/\bmovie(?:\s+movie)+\b/gi, (match) => {
    replacements += 1;
    // Preserve the casing of the first occurrence.
    const first = match.split(/\s+/)[0];
    return first;
  });
  return { text, replacements };
}

export function ReadmePreview({ media }: ReadmePreviewProps) {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [sections, setSections] = useState<ReadmeSections>(DEFAULT_SECTIONS);

  const { content, sanitizedCount } = useMemo(() => {
    const raw = buildReadmeContent(media, sections);
    const { text, replacements } = sanitizeMovieCommandPrefix(raw);
    return { content: text, sanitizedCount: replacements };
  }, [media, sections]);
  const lineCount = content.split("\n").length;
  const charCount = content.length;
  const enabledCount = Object.values(sections).filter(Boolean).length;

  const toggleSection = (key: keyof ReadmeSections) => {
    setSections((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const handleCopy = async () => {
    await navigator.clipboard.writeText(content);
    setCopied(true);
    toast.success("Copied README.md content to clipboard");
    setTimeout(() => setCopied(false), 2000);
  };

  const handleSave = () => {
    // Defense-in-depth: re-sanitize at write time in case content is mutated
    // by future code paths between preview and save.
    const { text: safeContent, replacements } = sanitizeMovieCommandPrefix(content);
    const blob = new Blob([safeContent], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "README.md";
    a.click();
    URL.revokeObjectURL(url);
    const sanitizedNote =
      replacements > 0
        ? ` · auto-fixed ${replacements} \`movie movie\` occurrence${replacements === 1 ? "" : "s"}`
        : "";
    toast.success("README.md saved", {
      description: `${lineCount} lines · ${charCount} characters${sanitizedNote}`,
    });
  };

  return (
    <>
      <Collapsible open={open} onOpenChange={setOpen}>
        <CollapsibleTrigger asChild>
          <Button variant="ghost" size="sm" className="gap-2 text-muted-foreground">
            <FileText className="h-4 w-4" />
            README.md preview
            {open ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
          </Button>
        </CollapsibleTrigger>
        <CollapsibleContent className="mt-2">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
              <div className="space-y-1">
                <CardTitle className="text-base flex items-center gap-2">
                  <FileText className="h-4 w-4 text-primary" />
                  README.md
                </CardTitle>
                <p className="text-xs text-muted-foreground">
                  Exact content that will be written · {lineCount} lines · {charCount} chars
                </p>
                {sanitizedCount > 0 && (
                  <p className="text-xs text-primary flex items-center gap-1">
                    <ShieldCheck className="h-3 w-3" />
                    Auto-fixed {sanitizedCount} <code className="px-1 rounded bg-muted">movie movie</code>{" "}
                    occurrence{sanitizedCount === 1 ? "" : "s"} → <code className="px-1 rounded bg-muted">movie</code>
                  </p>
                )}
              </div>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" className="gap-1.5" onClick={handleCopy}>
                  {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
                  {copied ? "Copied" : "Copy"}
                </Button>
                <Button size="sm" className="gap-1.5" onClick={() => setConfirmOpen(true)}>
                  <Save className="h-3.5 w-3.5" />
                  Review & Save
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="rounded-md border bg-muted/30 p-3">
                <div className="flex items-center gap-2 mb-3">
                  <SlidersHorizontal className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm font-medium">Sections</span>
                  <span className="text-xs text-muted-foreground">
                    {enabledCount} of {SECTION_LABELS.length} enabled
                  </span>
                </div>
                <div className="grid grid-cols-2 gap-x-4 gap-y-2 sm:grid-cols-3">
                  {SECTION_LABELS.map(({ key, label, icon }) => (
                    <div key={key} className="flex items-center justify-between gap-2">
                      <Label
                        htmlFor={`section-${key}`}
                        className="flex items-center gap-1.5 text-sm font-normal cursor-pointer"
                      >
                        <span>{icon}</span>
                        <span>{label}</span>
                      </Label>
                      <Switch
                        id={`section-${key}`}
                        checked={sections[key]}
                        onCheckedChange={() => toggleSection(key)}
                      />
                    </div>
                  ))}
                </div>
              </div>

              <pre className="max-h-96 overflow-auto rounded-md border bg-muted/40 p-4 text-xs font-mono text-foreground whitespace-pre-wrap break-words">
                {content}
              </pre>
            </CardContent>
          </Card>
        </CollapsibleContent>
      </Collapsible>

      <ReadmeDiffDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        newContent={content}
        onConfirm={handleSave}
      />
    </>
  );
}
