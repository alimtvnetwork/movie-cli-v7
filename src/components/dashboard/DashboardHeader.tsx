import type { MediaItem } from "@/types/media";
import { CommandPalette } from "./CommandPalette";

interface DashboardHeaderProps {
  media: MediaItem[];
}

export function DashboardHeader({ media }: DashboardHeaderProps) {
  return (
    <header className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">
          Media Library
        </h1>
        <p className="text-muted-foreground mt-1">
          {media.length} titles in your collection
        </p>
      </div>
      <div className="shrink-0">
        <CommandPalette />
      </div>
    </header>
  );
}
