import { ref, type Ref } from "vue";
import type { ChangedFile, DiffSection } from "@git-diff/domain";

// side is "left" / "right" — the value the backend expects.
export interface CommentTarget {
    file: ChangedFile;
    section: DiffSection;
    line: number;
    side: string;
    /** First line of the comment range; undefined for single-line drafts. */
    startLine?: number;
    /** Side of the range start; undefined when the range stays on `side`. */
    startSide?: string;
}

export interface UseCommentDraft {
    target: Ref<CommentTarget | null>;
    draft: Ref<string>;
    open: Ref<boolean>;
    begin: (target: CommentTarget) => void;
    setDraft: (value: string) => void;
    cancel: () => void;
    close: () => void;
}

export function useCommentDraft(): UseCommentDraft {
    const target = ref<CommentTarget | null>(null);
    const draft = ref("");
    const open = ref(false);

    function begin(t: CommentTarget): void {
        target.value = t;
        draft.value = "";
        open.value = true;
    }

    function setDraft(value: string): void {
        draft.value = value;
    }

    function cancel(): void {
        target.value = null;
        draft.value = "";
        open.value = false;
    }

    function close(): void {
        target.value = null;
        draft.value = "";
        open.value = false;
    }

    return { target, draft, open, begin, setDraft, cancel, close };
}
