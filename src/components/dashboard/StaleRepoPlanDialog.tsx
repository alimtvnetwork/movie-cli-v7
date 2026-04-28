import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ClipboardList, ShieldCheck, Copy, Check } from "lucide-react";
import { toast } from "sonner";

const REMOTE = "origin";
const BRANCH = "main";

const PLACEHOLDER_BACKUP_BRANCH = "backup/stale-repo-<YYYYMMDD-HHMMSS>";
const PLACEHOLDER_STASH_MSG = "stale-repo backup <UTC-ISO-TIMESTAMP>";

const PLANNED_COMMANDS: Array<{ cmd: string; note: string; destructive: boolean }> = [
  {
    cmd: `git branch ${PLACEHOLDER_BACKUP_BRANCH} <LOCAL_SHA>`,
    note: "Safety: snapshot current HEAD as a recovery branch (only if you have local commits / changes).",
    destructive: false,
  },
  {
    cmd: `git stash push --include-untracked -m "${PLACEHOLDER_STASH_MSG}"`,
    note: "Safety: stash dirty files + untracked files so they can be restored later (only if worktree is dirty).",
    destructive: false,
  },
  {
    cmd: `git fetch ${REMOTE}`,
    note: "Read-only: download the latest refs from the remote. No working-tree changes.",
    destructive: false,
  },
  {
    cmd: `git reset --hard ${REMOTE}/${BRANCH}`,
    note: "DESTRUCTIVE: discards <AHEAD> local commit(s) and any uncommitted changes on the current branch.",
    destructive: true,
  },
  {
    cmd: "git clean -fd",
    note: "DESTRUCTIVE: deletes <UNTRACKED_COUNT> untracked file(s) and empty directories permanently.",
    destructive: true,
  },
];

const SUMMARY_BLOCK = `────────────────────────────────────────────────────────────────────────
              FINAL CONFIRMATION — review before applying
────────────────────────────────────────────────────────────────────────
  Repository       : <REPO_ROOT>
  Branch           : <CURRENT_BRANCH>  →  ${REMOTE}/${BRANCH}
  Local  HEAD      : <LOCAL_SHA[:10]>   <LOCAL_HEAD_MSG>
  Remote HEAD      : <REMOTE_SHA[:10]>  <REMOTE_HEAD_MSG>
  Behind remote    : <BEHIND> commit(s)    [will be pulled in]
  Ahead of remote  : <AHEAD> commit(s)     [will be DISCARDED]
  Dirty worktree   : <yes|no>              [uncommitted changes will be LOST]
  Untracked files  : <UNTRACKED_COUNT>     [will be DELETED by clean -fd]

  Backup plan:
      Backup branch  : ${PLACEHOLDER_BACKUP_BRANCH}  →  <LOCAL_SHA[:10]>
      Stash entry    : "${PLACEHOLDER_STASH_MSG}"
      Recover with   : git checkout ${PLACEHOLDER_BACKUP_BRANCH} && git stash pop
────────────────────────────────────────────────────────────────────────`;

export function StaleRepoPlanDialog() {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const text = [
      "# Dry-run plan — scripts/check-stale-repo.sh (no --apply)",
      "# Run this to see the live numbers (read-only, no mutations):",
      "bash scripts/check-stale-repo.sh",
      "",
      "# Commands the script would run on --apply (for reference only):",
      ...PLANNED_COMMANDS.map((c) => c.cmd),
    ].join("\n");
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      toast.success("Plan copied to clipboard");
      setTimeout(() => setCopied(false), 1500);
    } catch {
      toast.error("Could not copy to clipboard");
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <ClipboardList className="h-4 w-4" />
          Preview planned commands
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-primary" />
            Dry-run: what the stale-repo helper would do
          </DialogTitle>
          <DialogDescription>
            Read-only preview. Nothing here runs git, touches your worktree, or commits anything.
            Placeholders like <code className="font-mono text-xs">&lt;LOCAL_SHA&gt;</code> are filled in
            with live values when you actually run{" "}
            <code className="font-mono text-xs">bash scripts/check-stale-repo.sh</code> in your terminal.
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="max-h-[55vh] pr-3">
          <div className="space-y-4">
            <section>
              <h3 className="mb-2 text-sm font-semibold">Planned commands (in order)</h3>
              <ol className="space-y-2">
                {PLANNED_COMMANDS.map((c, i) => (
                  <li
                    key={c.cmd}
                    className="rounded-md border bg-muted/40 p-3"
                  >
                    <div className="flex items-start gap-2">
                      <Badge
                        variant={c.destructive ? "destructive" : "secondary"}
                        className="shrink-0"
                      >
                        {i + 1}
                      </Badge>
                      <div className="min-w-0 flex-1">
                        <code className="block break-all font-mono text-xs text-foreground">
                          {c.cmd}
                        </code>
                        <p className="mt-1 text-xs text-muted-foreground">{c.note}</p>
                      </div>
                    </div>
                  </li>
                ))}
              </ol>
            </section>

            <section>
              <h3 className="mb-2 text-sm font-semibold">
                Final confirmation summary (template)
              </h3>
              <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 text-[11px] leading-relaxed text-foreground">
                {SUMMARY_BLOCK}
              </pre>
            </section>

            <section className="rounded-md border border-primary/30 bg-primary/5 p-3 text-xs text-foreground">
              <p className="font-semibold">To see real values for your repo:</p>
              <code className="mt-1 block font-mono text-[11px]">
                bash scripts/check-stale-repo.sh
              </code>
              <p className="mt-1 text-muted-foreground">
                That invocation is read-only (no <code className="font-mono">--apply</code>) and prints
                this same summary with live SHAs, behind/ahead counts, and the actual backup branch
                name it would create.
              </p>
            </section>
          </div>
        </ScrollArea>

        <DialogFooter className="gap-2 sm:gap-2">
          <Button variant="outline" size="sm" onClick={handleCopy} className="gap-2">
            {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
            {copied ? "Copied" : "Copy plan"}
          </Button>
          <Button size="sm" onClick={() => setOpen(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
