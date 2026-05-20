import type { ChangedFile } from "@git-diff/contracts";
import type { ComputedRef, Ref } from "vue";

export interface UseDiffNavigationOptions {
  files: ComputedRef<ChangedFile[]>;
  changedIndex: ComputedRef<number>;
  onSelect: (path: string) => void;
}

// Provides the keyboard navigation primitives used by App.vue: moving between
// changed files in the diff list and jumping between hunk anchors in the body.
export function useDiffNavigation({ files, changedIndex, onSelect }: UseDiffNavigationOptions) {
  function selectAdjacent(delta: number): void {
    if (files.value.length === 0) return;

    const next = Math.min(Math.max(changedIndex.value + delta, 0), files.value.length - 1);
    const file = files.value[next];

    if (file) onSelect(file.path);
  }

  function jumpToHunk(delta: number): void {
    const anchors = Array.from(document.querySelectorAll<HTMLElement>("[data-hunk-anchor]"));

    if (anchors.length === 0) return;

    const midpoint = window.innerHeight / 2;
    let current = 0;

    for (let i = 0; i < anchors.length; i++) {
      const rect = anchors[i]!.getBoundingClientRect();
      if (rect.top <= midpoint) current = i;
      else break;
    }

    const target = Math.max(0, Math.min(anchors.length - 1, current + delta));
    const element = anchors[target]!;
    element.scrollIntoView({ block: "center", behavior: "smooth" });
    element.classList.add("gd-hunk-flash");
    window.setTimeout(() => element.classList.remove("gd-hunk-flash"), 350);
  }

  return { selectAdjacent, jumpToHunk };
}

export interface UseKeyboardShortcutsOptions {
  enabled: Ref<boolean>;
  onSelectAdjacent: (delta: number) => void;
  onJumpToHunk: (delta: number) => void;
  onToggleViewed: () => void;
  onStartReview: () => void;
  onOpenSearch: () => void;
}

// Binds the global keydown listener that powers j/k/v/n/p/Cmd+Enter/Cmd-F
// shortcuts. The enabled ref gates the listener (e.g. only fires when
// authMode === "ready"). Returns an unbind function for onUnmounted.
export function useKeyboardShortcuts(opts: UseKeyboardShortcutsOptions): () => void {
  function onKeydown(event: KeyboardEvent): void {
    if (!opts.enabled.value) return;

    const target = event.target as HTMLElement | null;

    if (
      target &&
      (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable)
    ) {
      return;
    }

    if (event.key === "j" || event.key === "ArrowDown") {
      if (event.metaKey || event.ctrlKey || event.altKey) return;
      event.preventDefault();
      opts.onSelectAdjacent(1);
    } else if (event.key === "k" || event.key === "ArrowUp") {
      if (event.metaKey || event.ctrlKey || event.altKey) return;
      event.preventDefault();
      opts.onSelectAdjacent(-1);
    } else if (event.key === "v") {
      opts.onToggleViewed();
    } else if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
      event.preventDefault();
      opts.onStartReview();
    } else if (event.key === "n" || event.key === "p") {
      if (event.metaKey || event.ctrlKey || event.altKey) return;
      event.preventDefault();
      opts.onJumpToHunk(event.key === "n" ? 1 : -1);
    } else if ((event.key === "f" && (event.metaKey || event.ctrlKey)) || event.key === "/") {
      if (event.key === "/" && (event.metaKey || event.ctrlKey || event.altKey)) return;
      event.preventDefault();
      opts.onOpenSearch();
    }
  }

  document.addEventListener("keydown", onKeydown);

  return () => document.removeEventListener("keydown", onKeydown);
}
