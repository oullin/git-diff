import { ref, type Ref } from "vue";
import type { ChangedFile, DiffSection } from "@git-diff/contracts";

// CommentTarget pins a comment draft to a specific line. side mirrors the
// side string the backend expects ("left" / "right") and is derived from
// whether the patch line came from the old or new file.
export interface CommentTarget {
    file: ChangedFile;
    section: DiffSection;
    line: number;
    side: string;
}

// useCommentDraft holds the AddCommentDialog state: the target line, the
// in-progress draft HTML, and the dialog-open flag. Opening / closing the
// draft is straightforward, but saving it requires backend access plus the
// active-review pointer, so that stays in App.vue.
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
