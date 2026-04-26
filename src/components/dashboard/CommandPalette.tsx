import { useEffect, useState } from "react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import { Button } from "@/components/ui/button";
import { Search, Copy, Check } from "lucide-react";
import { toast } from "sonner";
import {
  COMMAND_ENTRIES,
  COMMAND_GROUPS,
  type CommandEntry,
} from "./command-data";

/**
 * Global command palette. Opens with Cmd/Ctrl+K or via the trigger button.
 * Selecting an item copies the exact `movie ...` command string to the
 * clipboard so it can be pasted straight into a terminal.
 */

function buildSearchValue(entry: CommandEntry): string {
  // cmdk filters by the `value` prop. Combine cmd + desc + keywords so
  // typing partial words ("scan", "tmdb", "undo") still matches.
  return [entry.cmd, entry.desc, entry.group, ...(entry.keywords ?? [])]
    .join(" ")
    .toLowerCase();
}

async function copyCommand(cmd: string) {
  await navigator.clipboard.writeText(cmd);
  toast.success("Copied", { description: cmd });
}

function PaletteItem({
  entry,
  onSelect,
}: {
  entry: CommandEntry;
  onSelect: () => void;
}) {
  const [copied, setCopied] = useState(false);

  const handleSelect = async () => {
    await copyCommand(entry.cmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 1200);
    onSelect();
  };

  return (
    <CommandItem
      value={buildSearchValue(entry)}
      onSelect={handleSelect}
      className="flex items-start gap-3"
    >
      <div className="mt-0.5 shrink-0">
        {copied ? (
          <Check className="h-4 w-4 text-primary" />
        ) : (
          <Copy className="h-4 w-4 text-muted-foreground" />
        )}
      </div>
      <div className="min-w-0 flex-1 space-y-0.5">
        <code className="block font-mono text-xs text-foreground break-all">
          {entry.cmd}
        </code>
        <p className="text-xs text-muted-foreground">{entry.desc}</p>
      </div>
    </CommandItem>
  );
}

export function CommandPalette() {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const isMeta = e.metaKey || e.ctrlKey;
      if (isMeta && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setOpen((prev) => !prev);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        className="gap-2 text-muted-foreground"
        onClick={() => setOpen(true)}
        aria-label="Open command palette"
      >
        <Search className="h-3.5 w-3.5" />
        <span className="hidden sm:inline">Search commands…</span>
        <span className="sm:hidden">Search</span>
        <kbd className="hidden sm:inline rounded border border-border bg-background px-1.5 py-0.5 font-mono text-[10px] uppercase">
          ⌘K
        </kbd>
      </Button>

      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Type a command or keyword (scan, move, undo…)" />
        <CommandList>
          <CommandEmpty>No matching commands.</CommandEmpty>
          {COMMAND_GROUPS.map((group, idx) => {
            const items = COMMAND_ENTRIES.filter((e) => e.group === group);
            return (
              <div key={group}>
                {idx > 0 && <CommandSeparator />}
                <CommandGroup heading={group}>
                  {items.map((entry) => (
                    <PaletteItem
                      key={entry.cmd}
                      entry={entry}
                      onSelect={() => setOpen(false)}
                    />
                  ))}
                </CommandGroup>
              </div>
            );
          })}
        </CommandList>
        <div className="flex items-center justify-between border-t px-3 py-2 text-[10px] text-muted-foreground">
          <span>Enter or click to copy</span>
          <span>Esc to close · ⌘K to toggle</span>
        </div>
      </CommandDialog>
    </>
  );
}