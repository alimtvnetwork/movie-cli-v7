import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Compass, Copy, Check, ExternalLink } from "lucide-react";
import { toast } from "sonner";

/**
 * Mirror of the README "Jump to a command" table, with per-command copy
 * buttons. The README itself can't render JS, so this lives in the dashboard.
 *
 * Anchors point at the rendered README sections (GitHub-style slugs).
 */

interface JumpCommand {
  cmd: string;
  label: string;
}

interface JumpSection {
  title: string;
  icon: string;
  anchor: string;
  commands: JumpCommand[];
}

/**
 * Vertical offset (px) reserved at the top of the viewport when scrolling to
 * an anchor, so the heading isn't tucked under a sticky header / mobile bar.
 * Mobile gets a larger buffer because mobile browsers often render a sticky
 * URL bar that eats the top of the viewport.
 */
const SCROLL_OFFSET_DESKTOP = 80;
const SCROLL_OFFSET_MOBILE = 96;
const MOBILE_BREAKPOINT_PX = 640;

function scrollToAnchor(anchor: string) {
  const target = document.getElementById(anchor);
  if (!target) {
    window.location.hash = anchor;
    return;
  }
  const isMobile = window.innerWidth < MOBILE_BREAKPOINT_PX;
  const offset = isMobile ? SCROLL_OFFSET_MOBILE : SCROLL_OFFSET_DESKTOP;
  const top = target.getBoundingClientRect().top + window.scrollY - offset;
  window.scrollTo({ top, behavior: "smooth" });
  history.replaceState(null, "", `#${anchor}`);
}

const SECTIONS: JumpSection[] = [
  {
    title: "Scanning & Library",
    icon: "📚",
    anchor: "scanning--library",
    commands: [
      { cmd: "movie scan /path/to/media", label: "Scan a folder" },
      { cmd: "movie ls", label: "List library" },
      { cmd: "movie info 123", label: "Show details" },
    ],
  },
  {
    title: "File Management",
    icon: "🗂️",
    anchor: "file-management",
    commands: [
      { cmd: "movie move", label: "Interactive move" },
      { cmd: "movie move --all", label: "Batch move" },
      { cmd: "movie rename", label: "Batch rename" },
    ],
  },
  {
    title: "History & Undo",
    icon: "⏪",
    anchor: "history--undo",
    commands: [
      { cmd: "movie undo", label: "Revert last op" },
      { cmd: "movie history", label: "Show history" },
    ],
  },
  {
    title: "Discovery & Organization",
    icon: "🔍",
    anchor: "discovery--organization",
    commands: [
      { cmd: "movie search inception", label: "TMDb search" },
      { cmd: "movie suggest", label: "Recommendations" },
      { cmd: "movie tag add 1 favorite", label: "Tag a media" },
    ],
  },
  {
    title: "Maintenance & Debugging",
    icon: "🛠️",
    anchor: "maintenance--debugging",
    commands: [
      { cmd: "movie duplicates", label: "Find duplicates" },
      { cmd: "movie cleanup", label: "Remove stale entries" },
      { cmd: "movie stats", label: "Library stats" },
    ],
  },
  {
    title: "Configuration & System",
    icon: "⚙️",
    anchor: "configuration--system",
    commands: [
      { cmd: "movie config", label: "View config" },
      { cmd: "movie config set tmdb_api_key YOUR_KEY", label: "Set TMDb key" },
      { cmd: "movie version", label: "Show version" },
    ],
  },
];

function CopyButton({ cmd }: { cmd: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(cmd);
    setCopied(true);
    toast.success("Copied", { description: cmd });
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <Button
      variant="ghost"
      size="icon"
      className="h-7 w-7 shrink-0"
      onClick={handleCopy}
      aria-label={`Copy ${cmd}`}
    >
      {copied ? (
        <Check className="h-3.5 w-3.5 text-primary" />
      ) : (
        <Copy className="h-3.5 w-3.5" />
      )}
    </Button>
  );
}

export function JumpToCommandTable() {
  return (
    <Card className="border-primary/20">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-2">
          <CardTitle className="text-base flex items-center gap-2">
            <Compass className="h-4 w-4 text-primary" />
            Jump to a command
          </CardTitle>
          <Badge variant="outline" className="text-xs">
            README mirror
          </Badge>
        </div>
        <p className="text-xs text-muted-foreground">
          Top commands per section. Copy directly, or jump to the README
          subsection for full reference.
        </p>
      </CardHeader>
      <CardContent>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {SECTIONS.map((section) => (
            <div
              key={section.anchor}
              className="rounded-md border border-border bg-muted/20 p-3 space-y-2"
            >
              <a
                href={`#${section.anchor}`}
                onClick={(e) => {
                  e.preventDefault();
                  scrollToAnchor(section.anchor);
                }}
                aria-label={`Jump to ${section.title} section (${section.commands.length} commands)`}
                style={{ scrollMarginTop: SCROLL_OFFSET_MOBILE }}
                className="flex items-center justify-between gap-2 text-sm font-medium hover:text-primary transition-colors"
              >
                <span className="flex items-center gap-1.5">
                  <span aria-hidden="true">{section.icon}</span>
                  {section.title}
                </span>
                <ExternalLink className="h-3 w-3 opacity-60" aria-hidden="true" />
              </a>
              <div className="space-y-0.5">
                {section.commands.map((c) => (
                  <div
                    key={c.cmd}
                    className="group flex items-center justify-between gap-2 rounded px-1.5 py-1 hover:bg-background"
                  >
                    <div className="min-w-0 flex-1">
                      <code className="block font-mono text-xs text-foreground truncate">
                        {c.cmd}
                      </code>
                      <p className="text-[10px] text-muted-foreground">
                        {c.label}
                      </p>
                    </div>
                    <CopyButton cmd={c.cmd} />
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}