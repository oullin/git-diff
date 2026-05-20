import { ref, type Ref } from "vue";
import type { ReviewDetail } from "@git-diff/contracts";
import { formatReviewAsMarkdown } from "@lib/reviewMarkdown";

// useReviewMarkdownCopy encapsulates the "Copy review as Markdown" toolbar
// button: a tri-state status ref ("idle" | "copied" | "error") plus the copy
// action itself. The button re-arms 1.5s after a successful copy so the
// host doesn't need to schedule that itself.
export interface UseReviewMarkdownCopy {
  state: Ref<"idle" | "copied" | "error">;
  copy: (detail: ReviewDetail) => Promise<void>;
}

export interface UseReviewMarkdownCopyOptions {
  resetMs?: number;
  onError?: (cause: unknown) => void;
}

export function useReviewMarkdownCopy(
  opts: UseReviewMarkdownCopyOptions = {},
): UseReviewMarkdownCopy {
  const resetMs = opts.resetMs ?? 1500;
  const state = ref<"idle" | "copied" | "error">("idle");

  async function copy(detail: ReviewDetail): Promise<void> {
    try {
      const markdown = formatReviewAsMarkdown(detail);
      await navigator.clipboard.writeText(markdown);
      state.value = "copied";

      setTimeout(() => {
        if (state.value === "copied") state.value = "idle";
      }, resetMs);
    } catch (cause) {
      state.value = "error";
      opts.onError?.(cause);
    }
  }

  return { state, copy };
}
