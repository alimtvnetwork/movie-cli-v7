import { useMemo, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Terminal,
  Copy,
  Check,
  AlertTriangle,
  ShieldCheck,
  ChevronDown,
  ChevronRight,
} from "lucide-react";
import { toast } from "sonner";
import {
  COMMAND_ENTRIES,
  COMMAND_GROUPS,
  type CommandExampleInvocation,
} from "./command-data";

/**
 * Command reference panel for the `movie` CLI.
 *
 * The CLI command tree is FLAT — every command is a direct subcommand of
 * `movie`. There is NO `movie movie <cmd>` form. This panel exists so docs
 * and README authors always have the correct syntax in view.
 *
 * See mem://constraints/command-syntax-flat for the full rule.
 */

interface CommandExample {
  cmd: string;
  desc: string;
  examples?: CommandExampleInvocation[];
}

interface CommandGroup {
  title: string;
  icon: string;
  commands: CommandExample[];
}

const GROUP_ICONS: Record<string, string> = {
  "Setup & Info": "⚙️",
  "Library Management": "📚",
  "File Operations": "🗂️",
  "Tags & Watchlist": "🏷️",
};

async function copyToClipboard(text: string) {
  await navigator.clipboard.writeText(text);
  toast.success("Copied", { description: text });
}

function ExampleRow({ cmd, note }: CommandExampleInvocation) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await copyToClipboard(cmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className="flex items-start justify-between gap-2 rounded border border-dashed border-border/60 bg-background/40 px-2 py-1.5">
      <div className="min-w-0 flex-1 space-y-0.5">
        <code className="block font-mono text-[11px] text-foreground break-all">
          {cmd}
        </code>
        <p className="text-[10px] text-muted-foreground">{note}</p>
      </div>
      <Button
        variant="ghost"
        size="icon"
        className="h-6 w-6 shrink-0"
        onClick={handleCopy}
        aria-label={`Copy example: ${cmd}`}
      >
        {copied ? (
          <Check className="h-3 w-3 text-primary" />
        ) : (
          <Copy className="h-3 w-3" />
        )}
      </Button>
    </div>
  );
}

function CommandRow({ cmd, desc, examples }: CommandExample) {
  const [copied, setCopied] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const hasExamples = !!examples && examples.length > 0;

  const handleCopy = async () => {
    await copyToClipboard(cmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className="group rounded-md border border-transparent px-2 py-1.5 hover:border-border hover:bg-muted/40">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1 space-y-0.5">
          <code className="block font-mono text-xs text-foreground break-all">
            {cmd}
          </code>
          <p className="text-xs text-muted-foreground">{desc}</p>
        </div>
        <div className="flex shrink-0 items-center gap-1">
          {hasExamples && (
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => setExpanded((v) => !v)}
              aria-label={
                expanded
                  ? `Hide examples for ${cmd}`
                  : `Show ${examples!.length} example${examples!.length === 1 ? "" : "s"} for ${cmd}`
              }
              aria-expanded={expanded}
            >
              {expanded ? (
                <ChevronDown className="h-3.5 w-3.5" />
              ) : (
                <ChevronRight className="h-3.5 w-3.5" />
              )}
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 opacity-0 group-hover:opacity-100"
            onClick={handleCopy}
            aria-label={`Copy ${cmd}`}
          >
            {copied ? (
              <Check className="h-3.5 w-3.5 text-primary" />
            ) : (
              <Copy className="h-3.5 w-3.5" />
            )}
          </Button>
        </div>
      </div>
      {hasExamples && (
        <div className="mt-1 flex items-center gap-1 pl-1">
          <Badge variant="secondary" className="h-4 px-1.5 text-[9px]">
            {examples!.length} example{examples!.length === 1 ? "" : "s"}
          </Badge>
        </div>
      )}
      {hasExamples && expanded && (
        <div className="mt-2 space-y-1.5 pl-3">
          {examples!.map((ex) => (
            <ExampleRow key={ex.cmd} {...ex} />
          ))}
        </div>
      )}
    </div>
  );
}

export function CommandReferencePanel() {
  const groups = useMemo<CommandGroup[]>(
    () =>
      COMMAND_GROUPS.map((title) => ({
        title,
        icon: GROUP_ICONS[title] ?? "📦",
        commands: COMMAND_ENTRIES.filter((e) => e.group === title).map((e) => ({
          cmd: e.cmd,
          desc: e.desc,
          examples: e.examples,
        })),
      })),
    [],
  );

  return (
    <Card className="border-primary/20">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-2">
          <CardTitle className="text-base flex items-center gap-2">
            <Terminal className="h-4 w-4 text-primary" />
            Command reference
          </CardTitle>
          <Badge variant="outline" className="gap-1 text-xs">
            <ShieldCheck className="h-3 w-3" />
            Flat syntax
          </Badge>
        </div>
        <p className="text-xs text-muted-foreground">
          Use these exact commands when writing docs or README content. Click any line to copy.
        </p>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3">
          <div className="flex items-start gap-2">
            <AlertTriangle className="h-4 w-4 shrink-0 text-destructive mt-0.5" />
            <div className="space-y-1 text-xs">
              <p className="font-medium text-foreground">
                Always <code className="rounded bg-muted px-1">movie &lt;cmd&gt;</code> — never{" "}
                <code className="rounded bg-muted px-1">movie movie &lt;cmd&gt;</code>
              </p>
              <p className="text-muted-foreground">
                The CLI is flat. Any nested form is a bug and gets auto-fixed by the README sanitizer.
              </p>
            </div>
          </div>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          {groups.map((group) => (
            <div key={group.title} className="space-y-1">
              <div className="flex items-center gap-1.5 px-2 py-1">
                <span>{group.icon}</span>
                <h4 className="text-sm font-medium">{group.title}</h4>
                <span className="text-xs text-muted-foreground">
                  ({group.commands.length})
                </span>
              </div>
              <div className="space-y-0.5">
                {group.commands.map((c) => (
                  <CommandRow key={c.cmd} {...c} />
                ))}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
