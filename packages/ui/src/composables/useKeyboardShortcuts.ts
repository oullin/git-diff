import type { Ref } from "vue";

export interface UseKeyboardShortcutsOptions {
    enabled: Ref<boolean>;
    onSelectAdjacent: (delta: number) => void;
    onJumpToHunk: (delta: number) => void;
    onToggleViewed: () => void;
    onStartReview: () => void;
    onOpenSearch: () => void;
}

/**
 * Binds the global keydown listener that powers j/k/v/n/p/Cmd+Enter/Cmd+F
 * shortcuts. The `enabled` ref gates the listener (e.g. only fires when
 * authMode === "ready"). Returns an unbind function suitable for
 * onUnmounted.
 */
export function useKeyboardShortcuts(opts: UseKeyboardShortcutsOptions): () => void {
    function onKeydown(event: KeyboardEvent): void {
        if (!opts.enabled.value) {
            return;
        }

        const target = event.target as HTMLElement | null;

        if (
            target &&
            (target.tagName === "INPUT" ||
                target.tagName === "TEXTAREA" ||
                target.isContentEditable)
        ) {
            return;
        }

        if (event.key === "j" || event.key === "ArrowDown") {
            if (event.metaKey || event.ctrlKey || event.altKey) {
                return;
            }

            event.preventDefault();
            opts.onSelectAdjacent(1);
        } else if (event.key === "k" || event.key === "ArrowUp") {
            if (event.metaKey || event.ctrlKey || event.altKey) {
                return;
            }

            event.preventDefault();
            opts.onSelectAdjacent(-1);
        } else if (event.key === "v") {
            opts.onToggleViewed();
        } else if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
            event.preventDefault();
            opts.onStartReview();
        } else if (event.key === "n" || event.key === "p") {
            if (event.metaKey || event.ctrlKey || event.altKey) {
                return;
            }

            event.preventDefault();
            opts.onJumpToHunk(event.key === "n" ? 1 : -1);
        } else if ((event.key === "f" && (event.metaKey || event.ctrlKey)) || event.key === "/") {
            if (event.key === "/" && (event.metaKey || event.ctrlKey || event.altKey)) {
                return;
            }

            event.preventDefault();
            opts.onOpenSearch();
        }
    }

    document.addEventListener("keydown", onKeydown);

    return () => document.removeEventListener("keydown", onKeydown);
}
