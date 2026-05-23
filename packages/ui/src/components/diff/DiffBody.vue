<script setup lang="ts">
import { computed, ref, toRef } from "vue";
import { ChevronDown, ChevronUp, Loader2, Plus } from "lucide-vue-next";
import { languageFor } from "@lib/highlight";
import {
    applyExpansions,
    parsePatch,
    splitPatchLines,
    type HunkInfo,
    type PatchLine,
    type SplitRow,
} from "@lib/patch";
import { useContextExpansion } from "@composables/useContextExpansion";
import { useDiffHighlighting } from "@composables/useDiffHighlighting";
import { useDiffStyles } from "@composables/useDiffStyles";
import { useHunkInfo } from "@composables/useHunkInfo";
import {
    useLineSelection,
    type LineAnchor,
    type LineSelectionRange,
    type LineSide,
} from "@composables/useLineSelection";
import CommentThread from "@diff/CommentThread.vue";
import SplitHandle from "@diff/SplitHandle.vue";
import type {
    ChangedFile,
    DiffHunkStyle,
    DiffSection,
    DiffViewMode,
    ReviewComment,
} from "@git-diff/contracts";
import type { RichTextFeatures } from "@ui/rich-text-editor";

const props = withDefaults(
    defineProps<{
        file: ChangedFile;
        viewMode: DiffViewMode;
        diffStyle: DiffHunkStyle;
        density: "comfortable" | "compact";
        wordHighlight: boolean;
        hideWhitespace: boolean;
        comments: ReviewComment[];
        replyFeatures: RichTextFeatures;
        repoRoot: string;
        commitRef?: string;
        splitRatio?: number;
    }>(),
    { splitRatio: 0.5, commitRef: undefined },
);

const { getExpansions, isInflight, isDownwardEof, expandUp, expandDown } = useContextExpansion();
const { setAnchor, getAnchor, clear: clearAnchor, rangeTo } = useLineSelection();

const emit = defineEmits<{
    "add-comment": [section: DiffSection, line: PatchLine, range?: LineSelectionRange];
    "delete-comment": [comment: ReviewComment];
    "reply-comment": [parent: ReviewComment, bodyHtml: string];
    "update:splitRatio": [value: number];
}>();

function lineSideAndNumber(line: PatchLine): { side: LineSide; lineNumber: number } | null {
    if (line.type === "add" && line.newLine != null) {
        return { side: "right", lineNumber: line.newLine };
    }

    if (line.type === "del" && line.oldLine != null) {
        return { side: "left", lineNumber: line.oldLine };
    }

    // Context lines render on both sides in split view; default to right so a
    // single-line comment lands on the new-file side.
    if (line.newLine != null) {
        return { side: "right", lineNumber: line.newLine };
    }

    if (line.oldLine != null) {
        return { side: "left", lineNumber: line.oldLine };
    }

    return null;
}

function handleAddCommentClick(
    event: MouseEvent,
    section: DiffSection,
    line: PatchLine,
    explicitSide?: LineSide,
): void {
    const anchored = lineSideAndNumber(line);

    if (!anchored) {
        return;
    }

    const target: LineAnchor = {
        sectionId: section.id,
        side: explicitSide ?? anchored.side,
        lineNumber:
            explicitSide === "left"
                ? (line.oldLine ?? anchored.lineNumber)
                : explicitSide === "right"
                  ? (line.newLine ?? anchored.lineNumber)
                  : anchored.lineNumber,
    };

    if (event.shiftKey) {
        const range = rangeTo(target);

        if (range) {
            emit("add-comment", section, line, range);
            clearAnchor();

            return;
        }
    }

    setAnchor(target);
    emit("add-comment", section, line);
}

function isAnchored(section: DiffSection, line: PatchLine, side: LineSide): boolean {
    const anchor = getAnchor();

    if (!anchor || anchor.sectionId !== section.id || anchor.side !== side) {
        return false;
    }

    const lineNum = side === "left" ? line.oldLine : line.newLine;

    return lineNum != null && lineNum === anchor.lineNumber;
}

const wrapperStyle = computed(() => {
    const left = Math.min(80, Math.max(20, props.splitRatio * 100));
    const right = 100 - left;

    return {
        minWidth: props.viewMode === "split" ? "1800px" : "1200px",
        position: "relative" as const,
        "--gd-split-cols": `${left.toFixed(2)}% ${right.toFixed(2)}%`,
    };
});

const sectionRefs = ref<Record<string, HTMLElement | null>>({});

function setSectionRef(id: string) {
    return (el: unknown) => {
        sectionRefs.value[id] = el instanceof HTMLElement ? el : null;
    };
}

const lineH = computed(() => (props.density === "compact" ? 22 : 24));
const lang = computed(() => languageFor(props.file.path));

const { colors, bgFor, numBgFor, barFor, sign } = useDiffStyles(toRef(props, "diffStyle"));
const { highlightHtml, withRanges, computeWordHi } = useDiffHighlighting(lang);
const { hunksFor, findHunk, nextHunkOldStart } = useHunkInfo(toRef(props, "hideWhitespace"));

function patchLines(section: DiffSection): PatchLine[] {
    const base = parsePatch(section, props.hideWhitespace);

    return applyExpansions(base, getExpansions(section.id));
}

function splitRows(section: DiffSection): SplitRow[] {
    return splitPatchLines(patchLines(section));
}

function expansionRequest(section: DiffSection) {
    return {
        sectionId: section.id,
        repoRoot: props.repoRoot,
        filePath: props.file.path,
        ref: props.commitRef,
    };
}

function onExpandUp(section: DiffSection, hunk: HunkInfo | null): void {
    if (!hunk || !props.repoRoot) {
        return;
    }

    void expandUp(expansionRequest(section), hunk);
}

function onExpandDown(section: DiffSection, hunk: HunkInfo | null): void {
    if (!hunk || !props.repoRoot) {
        return;
    }

    void expandDown(expansionRequest(section), hunk, nextHunkOldStart(section, hunk));
}

function canExpandUp(section: DiffSection, hunk: HunkInfo | null): boolean {
    if (!hunk) {
        return false;
    }

    const lowest = lowestVisibleAbove(section, hunk);

    return lowest > hunk.prevOldEnd + 1;
}

function canExpandDown(section: DiffSection, hunk: HunkInfo | null): boolean {
    if (!hunk) {
        return false;
    }

    const next = nextHunkOldStart(section, hunk);
    const highest = highestVisibleBelow(section, hunk, next);

    if (next != null) {
        return highest < next - 1;
    }

    return !isDownwardEof(section.id);
}

function lowestVisibleAbove(section: DiffSection, hunk: HunkInfo): number {
    let lowest = hunk.oldStart;

    for (const exp of getExpansions(section.id)) {
        if (exp.oldLine > hunk.prevOldEnd && exp.oldLine < hunk.oldStart && exp.oldLine < lowest) {
            lowest = exp.oldLine;
        }
    }

    return lowest;
}

function highestVisibleBelow(
    section: DiffSection,
    hunk: HunkInfo,
    nextStart: number | null,
): number {
    let highest = hunk.oldEnd;
    const upper = nextStart ?? Number.POSITIVE_INFINITY;

    for (const exp of getExpansions(section.id)) {
        if (exp.oldLine > hunk.oldEnd && exp.oldLine < upper && exp.oldLine > highest) {
            highest = exp.oldLine;
        }
    }

    return highest;
}

function commentsForLine(section: DiffSection, line: PatchLine | undefined): ReviewComment[] {
    if (!line) {
        return [];
    }

    return props.comments.filter((c) => {
        if (c.filePath !== props.file.path || c.diffSection !== section.kind) {
            return false;
        }

        const lineNumberForSide = c.side === "left" ? line.oldLine : line.newLine;

        return lineNumberForSide === c.lineNumber;
    });
}

interface CellRender {
    kind: "ctx" | "add" | "rem" | "empty";
    num: number | "";
    side: "left" | "right";
    html: string;
}

function renderPair(row: Extract<SplitRow, { kind: "pair" }>): {
    left: CellRender;
    right: CellRender;
} {
    const left = row.left;
    const right = row.right;
    let leftHtml = "";
    let rightHtml = "";
    let leftKind: CellRender["kind"] = left ? "rem" : "empty";
    let rightKind: CellRender["kind"] = right ? "add" : "empty";

    if (left && right && props.wordHighlight) {
        const { hiL, hiR } = computeWordHi(left.text, right.text);

        leftHtml = withRanges(left.text, hiL, "wh-rem");
        rightHtml = withRanges(right.text, hiR, "wh-add");
    } else {
        leftHtml = left ? highlightHtml(left.text) : "";
        rightHtml = right ? highlightHtml(right.text) : "";
    }

    return {
        left: { kind: leftKind, num: left?.oldLine ?? "", side: "left", html: leftHtml },
        right: { kind: rightKind, num: right?.newLine ?? "", side: "right", html: rightHtml },
    };
}

function hunkHeaderText(text: string): { range: string; trailer: string } {
    const parts = text.split("@@");

    if (parts.length < 3) {
        return { range: text, trailer: "" };
    }

    return {
        range: `@@${parts[1]}@@`,
        trailer: parts.slice(2).join("@@").trim(),
    };
}
</script>

<template>
    <div
        :style="{
            fontFamily: 'var(--font-mono)',
            fontSize: '13.5px',
            lineHeight: `${lineH}px`,
            background: 'var(--gd-bg-code, var(--gd-bg))',
        }"
    >
        <section v-for="section in file.sections" :key="section.id">
            <div data-diff-scroller="true" :style="{ overflowX: 'auto', overflowY: 'visible' }">
                <div :ref="setSectionRef(section.id)" :style="wrapperStyle">
                    <SplitHandle
                        v-if="viewMode === 'split'"
                        :ratio="splitRatio"
                        :container-el="sectionRefs[section.id] ?? null"
                        @update:ratio="(value) => emit('update:splitRatio', value)"
                    />
                    <template v-if="viewMode === 'split'">
                        <template v-for="row in splitRows(section)" :key="row.id">
                            <template v-if="row.kind === 'meta'">
                                <div
                                    v-if="row.line.text.startsWith('@@')"
                                    class="flex items-center"
                                    data-hunk-anchor
                                    :style="{
                                        gap: '10px',
                                        padding: '8px 16px',
                                        background: 'var(--gd-bg-gutter, var(--gd-panel))',
                                        color: 'var(--gd-text-3)',
                                        fontSize: '13px',
                                        borderTop: '1px solid var(--gd-border-soft)',
                                        borderBottom: '1px solid var(--gd-border-soft)',
                                        boxShadow: '0 1px 0 var(--gd-edge-hi-2) inset',
                                    }"
                                >
                                    <ChevronDown :size="12" />
                                    <span
                                        :style="{
                                            fontFamily: 'var(--font-mono)',
                                            color: 'var(--gd-accent)',
                                        }"
                                        >{{ hunkHeaderText(row.line.text).range }}</span
                                    >
                                    <span :style="{ color: 'var(--gd-text-3)' }">{{
                                        hunkHeaderText(row.line.text).trailer
                                    }}</span>
                                    <div class="flex-1" />
                                    <button
                                        type="button"
                                        title="Expand context up"
                                        :disabled="
                                            !canExpandUp(section, findHunk(section, row.line.id)) ||
                                            isInflight(section.id, 'up', row.line.id)
                                        "
                                        :style="{
                                            width: '26px',
                                            height: '26px',
                                            borderRadius: '6px',
                                            border: '1px solid transparent',
                                            background: 'transparent',
                                            color: 'var(--gd-text-3)',
                                            display: 'inline-flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            padding: 0,
                                            cursor:
                                                canExpandUp(
                                                    section,
                                                    findHunk(section, row.line.id),
                                                ) && !isInflight(section.id, 'up', row.line.id)
                                                    ? 'pointer'
                                                    : 'not-allowed',
                                            opacity: canExpandUp(
                                                section,
                                                findHunk(section, row.line.id),
                                            )
                                                ? 1
                                                : 0.35,
                                        }"
                                        @click="onExpandUp(section, findHunk(section, row.line.id))"
                                    >
                                        <Loader2
                                            v-if="isInflight(section.id, 'up', row.line.id)"
                                            :size="12"
                                            class="animate-spin"
                                        />
                                        <ChevronUp v-else :size="12" />
                                    </button>
                                    <button
                                        type="button"
                                        title="Expand context down"
                                        :disabled="
                                            !canExpandDown(
                                                section,
                                                findHunk(section, row.line.id),
                                            ) || isInflight(section.id, 'down', row.line.id)
                                        "
                                        :style="{
                                            width: '26px',
                                            height: '26px',
                                            borderRadius: '6px',
                                            border: '1px solid transparent',
                                            background: 'transparent',
                                            color: 'var(--gd-text-3)',
                                            display: 'inline-flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            padding: 0,
                                            cursor:
                                                canExpandDown(
                                                    section,
                                                    findHunk(section, row.line.id),
                                                ) && !isInflight(section.id, 'down', row.line.id)
                                                    ? 'pointer'
                                                    : 'not-allowed',
                                            opacity: canExpandDown(
                                                section,
                                                findHunk(section, row.line.id),
                                            )
                                                ? 1
                                                : 0.35,
                                        }"
                                        @click="
                                            onExpandDown(section, findHunk(section, row.line.id))
                                        "
                                    >
                                        <Loader2
                                            v-if="isInflight(section.id, 'down', row.line.id)"
                                            :size="12"
                                            class="animate-spin"
                                        />
                                        <ChevronDown v-else :size="12" />
                                    </button>
                                </div>
                            </template>
                            <template v-else-if="row.kind === 'context'">
                                <div
                                    class="gd-row relative"
                                    :style="{
                                        display: 'grid',
                                        gridTemplateColumns: 'var(--gd-split-cols, 1fr 1fr)',
                                    }"
                                >
                                    <div
                                        class="flex"
                                        :style="{
                                            background: 'transparent',
                                            minHeight: `${lineH}px`,
                                            minWidth: 0,
                                            overflow: 'hidden',
                                        }"
                                    >
                                        <div
                                            :style="{
                                                width: '56px',
                                                flexShrink: 0,
                                                textAlign: 'right',
                                                padding: '0 10px 0 4px',
                                                color: 'var(--gd-text-muted)',
                                                fontSize: '12px',
                                                fontVariantNumeric: 'tabular-nums',
                                                userSelect: 'none',
                                            }"
                                        >
                                            {{ row.line.oldLine ?? "" }}
                                        </div>
                                        <div
                                            :style="{
                                                flex: 1,
                                                minWidth: 0,
                                                padding: '0 12px',
                                                whiteSpace: 'pre',
                                                color: 'var(--gd-text)',
                                            }"
                                        >
                                            <span
                                                :style="{
                                                    display: 'inline-block',
                                                    width: '12px',
                                                    color: 'var(--gd-text-muted)',
                                                }"
                                                >&nbsp;</span
                                            ><code
                                                data-diff-line-text
                                                v-html="highlightHtml(row.line.text)"
                                            />
                                        </div>
                                    </div>
                                    <div
                                        class="flex"
                                        :style="{
                                            background: 'transparent',
                                            borderLeft: '1px solid var(--gd-border-soft)',
                                            minHeight: `${lineH}px`,
                                            minWidth: 0,
                                            overflow: 'hidden',
                                        }"
                                    >
                                        <div
                                            :style="{
                                                width: '56px',
                                                flexShrink: 0,
                                                textAlign: 'right',
                                                padding: '0 10px 0 4px',
                                                color: 'var(--gd-text-muted)',
                                                fontSize: '12px',
                                                fontVariantNumeric: 'tabular-nums',
                                                userSelect: 'none',
                                            }"
                                        >
                                            {{ row.line.newLine ?? "" }}
                                        </div>
                                        <div
                                            :style="{
                                                flex: 1,
                                                minWidth: 0,
                                                padding: '0 12px',
                                                whiteSpace: 'pre',
                                                color: 'var(--gd-text)',
                                            }"
                                        >
                                            <span
                                                :style="{
                                                    display: 'inline-block',
                                                    width: '12px',
                                                    color: 'var(--gd-text-muted)',
                                                }"
                                                >&nbsp;</span
                                            ><code
                                                data-diff-line-text
                                                v-html="highlightHtml(row.line.text)"
                                            />
                                        </div>
                                    </div>
                                    <button
                                        type="button"
                                        class="gd-add-comment"
                                        title="Add comment (shift-click to extend selection)"
                                        @click="(e) => handleAddCommentClick(e, section, row.line)"
                                    >
                                        <Plus :size="12" :stroke-width="2.5" />
                                    </button>
                                </div>
                                <template
                                    v-for="c in commentsForLine(section, row.line)"
                                    :key="c.id"
                                >
                                    <CommentThread
                                        :comment="c"
                                        :reply-features="replyFeatures"
                                        @delete="emit('delete-comment', c)"
                                        @reply="(body) => emit('reply-comment', c, body)"
                                    />
                                </template>
                            </template>
                            <template v-else>
                                <div
                                    class="gd-row relative"
                                    :style="{
                                        display: 'grid',
                                        gridTemplateColumns: 'var(--gd-split-cols, 1fr 1fr)',
                                    }"
                                >
                                    <div
                                        class="flex"
                                        :style="{
                                            background: bgFor(renderPair(row).left.kind),
                                            minHeight: `${lineH}px`,
                                            minWidth: 0,
                                            overflow: 'hidden',
                                        }"
                                    >
                                        <div
                                            :style="{
                                                width: '56px',
                                                flexShrink: 0,
                                                textAlign: 'right',
                                                padding: '0 10px 0 4px',
                                                background: numBgFor(renderPair(row).left.kind),
                                                color: 'var(--gd-text-muted)',
                                                fontSize: '12px',
                                                fontVariantNumeric: 'tabular-nums',
                                                userSelect: 'none',
                                                position: 'relative',
                                            }"
                                        >
                                            <div
                                                v-if="renderPair(row).left.kind === 'rem'"
                                                :style="{
                                                    position: 'absolute',
                                                    left: 0,
                                                    top: 0,
                                                    bottom: 0,
                                                    width: '2px',
                                                    background: barFor('rem'),
                                                }"
                                            />
                                            {{ renderPair(row).left.num }}
                                        </div>
                                        <div
                                            :style="{
                                                flex: 1,
                                                minWidth: 0,
                                                padding: '0 12px',
                                                whiteSpace: 'pre',
                                                color: 'var(--gd-text)',
                                            }"
                                        >
                                            <span
                                                :style="{
                                                    display: 'inline-block',
                                                    width: '12px',
                                                    color: 'var(--gd-text-muted)',
                                                }"
                                                >{{ sign(renderPair(row).left.kind) }}</span
                                            ><code
                                                data-diff-line-text
                                                v-html="renderPair(row).left.html"
                                            />
                                        </div>
                                    </div>
                                    <div
                                        class="flex"
                                        :style="{
                                            background: bgFor(renderPair(row).right.kind),
                                            borderLeft: '1px solid var(--gd-border-soft)',
                                            minHeight: `${lineH}px`,
                                            minWidth: 0,
                                            overflow: 'hidden',
                                        }"
                                    >
                                        <div
                                            :style="{
                                                width: '56px',
                                                flexShrink: 0,
                                                textAlign: 'right',
                                                padding: '0 10px 0 4px',
                                                background: numBgFor(renderPair(row).right.kind),
                                                color: 'var(--gd-text-muted)',
                                                fontSize: '12px',
                                                fontVariantNumeric: 'tabular-nums',
                                                userSelect: 'none',
                                                position: 'relative',
                                            }"
                                        >
                                            <div
                                                v-if="renderPair(row).right.kind === 'add'"
                                                :style="{
                                                    position: 'absolute',
                                                    left: 0,
                                                    top: 0,
                                                    bottom: 0,
                                                    width: '2px',
                                                    background: barFor('add'),
                                                }"
                                            />
                                            {{ renderPair(row).right.num }}
                                        </div>
                                        <div
                                            :style="{
                                                flex: 1,
                                                minWidth: 0,
                                                padding: '0 12px',
                                                whiteSpace: 'pre',
                                                color: 'var(--gd-text)',
                                            }"
                                        >
                                            <span
                                                :style="{
                                                    display: 'inline-block',
                                                    width: '12px',
                                                    color: 'var(--gd-text-muted)',
                                                }"
                                                >{{ sign(renderPair(row).right.kind) }}</span
                                            ><code
                                                data-diff-line-text
                                                v-html="renderPair(row).right.html"
                                            />
                                        </div>
                                    </div>
                                    <button
                                        v-if="row.left || row.right"
                                        type="button"
                                        class="gd-add-comment"
                                        title="Add comment (shift-click to extend selection)"
                                        @click="
                                            (e) =>
                                                handleAddCommentClick(
                                                    e,
                                                    section,
                                                    (row.right ?? row.left) as PatchLine,
                                                )
                                        "
                                    >
                                        <Plus :size="12" :stroke-width="2.5" />
                                    </button>
                                </div>
                                <template
                                    v-for="line in [row.left, row.right].filter(
                                        Boolean,
                                    ) as PatchLine[]"
                                    :key="`${row.id}:${line.id}`"
                                >
                                    <CommentThread
                                        v-for="c in commentsForLine(section, line)"
                                        :key="c.id"
                                        :comment="c"
                                        :reply-features="replyFeatures"
                                        @delete="emit('delete-comment', c)"
                                        @reply="(body) => emit('reply-comment', c, body)"
                                    />
                                </template>
                            </template>
                        </template>
                    </template>
                    <template v-else>
                        <template v-for="line in patchLines(section)" :key="line.id">
                            <div
                                v-if="line.type === 'meta' && line.text.startsWith('@@')"
                                class="flex items-center"
                                data-hunk-anchor
                                :style="{
                                    gap: '10px',
                                    padding: '8px 16px',
                                    background: 'var(--gd-bg-gutter, var(--gd-panel))',
                                    color: 'var(--gd-text-3)',
                                    fontSize: '13px',
                                    borderTop: '1px solid var(--gd-border-soft)',
                                    borderBottom: '1px solid var(--gd-border-soft)',
                                    boxShadow: '0 1px 0 var(--gd-edge-hi-2) inset',
                                }"
                            >
                                <ChevronDown :size="12" />
                                <span
                                    :style="{
                                        fontFamily: 'var(--font-mono)',
                                        color: 'var(--gd-accent)',
                                    }"
                                    >{{ hunkHeaderText(line.text).range }}</span
                                >
                                <span :style="{ color: 'var(--gd-text-3)' }">{{
                                    hunkHeaderText(line.text).trailer
                                }}</span>
                                <div class="flex-1" />
                                <button
                                    type="button"
                                    title="Expand context up"
                                    :disabled="
                                        !canExpandUp(section, findHunk(section, line.id)) ||
                                        isInflight(section.id, 'up', line.id)
                                    "
                                    :style="{
                                        width: '26px',
                                        height: '26px',
                                        borderRadius: '6px',
                                        border: '1px solid transparent',
                                        background: 'transparent',
                                        color: 'var(--gd-text-3)',
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        padding: 0,
                                        cursor:
                                            canExpandUp(section, findHunk(section, line.id)) &&
                                            !isInflight(section.id, 'up', line.id)
                                                ? 'pointer'
                                                : 'not-allowed',
                                        opacity: canExpandUp(section, findHunk(section, line.id))
                                            ? 1
                                            : 0.35,
                                    }"
                                    @click="onExpandUp(section, findHunk(section, line.id))"
                                >
                                    <Loader2
                                        v-if="isInflight(section.id, 'up', line.id)"
                                        :size="12"
                                        class="animate-spin"
                                    />
                                    <ChevronUp v-else :size="12" />
                                </button>
                                <button
                                    type="button"
                                    title="Expand context down"
                                    :disabled="
                                        !canExpandDown(section, findHunk(section, line.id)) ||
                                        isInflight(section.id, 'down', line.id)
                                    "
                                    :style="{
                                        width: '26px',
                                        height: '26px',
                                        borderRadius: '6px',
                                        border: '1px solid transparent',
                                        background: 'transparent',
                                        color: 'var(--gd-text-3)',
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        padding: 0,
                                        cursor:
                                            canExpandDown(section, findHunk(section, line.id)) &&
                                            !isInflight(section.id, 'down', line.id)
                                                ? 'pointer'
                                                : 'not-allowed',
                                        opacity: canExpandDown(section, findHunk(section, line.id))
                                            ? 1
                                            : 0.35,
                                    }"
                                    @click="onExpandDown(section, findHunk(section, line.id))"
                                >
                                    <Loader2
                                        v-if="isInflight(section.id, 'down', line.id)"
                                        :size="12"
                                        class="animate-spin"
                                    />
                                    <ChevronDown v-else :size="12" />
                                </button>
                            </div>
                            <template v-else-if="line.type !== 'meta'">
                                <div
                                    class="gd-row relative flex"
                                    :style="{
                                        background: bgFor(
                                            line.type === 'add'
                                                ? 'add'
                                                : line.type === 'del'
                                                  ? 'rem'
                                                  : 'ctx',
                                        ),
                                        minHeight: `${lineH}px`,
                                    }"
                                >
                                    <div
                                        :style="{
                                            width: '56px',
                                            flexShrink: 0,
                                            textAlign: 'right',
                                            padding: '0 10px 0 4px',
                                            background: numBgFor(
                                                line.type === 'add'
                                                    ? 'add'
                                                    : line.type === 'del'
                                                      ? 'rem'
                                                      : 'ctx',
                                            ),
                                            color: 'var(--gd-text-muted)',
                                            fontSize: '12px',
                                            fontVariantNumeric: 'tabular-nums',
                                            userSelect: 'none',
                                            position: 'relative',
                                        }"
                                    >
                                        {{ line.oldLine ?? "" }}
                                    </div>
                                    <div
                                        :style="{
                                            width: '56px',
                                            flexShrink: 0,
                                            textAlign: 'right',
                                            padding: '0 10px 0 4px',
                                            background: numBgFor(
                                                line.type === 'add'
                                                    ? 'add'
                                                    : line.type === 'del'
                                                      ? 'rem'
                                                      : 'ctx',
                                            ),
                                            color: 'var(--gd-text-muted)',
                                            fontSize: '12px',
                                            fontVariantNumeric: 'tabular-nums',
                                            userSelect: 'none',
                                            position: 'relative',
                                        }"
                                    >
                                        <div
                                            v-if="line.type === 'add' || line.type === 'del'"
                                            :style="{
                                                position: 'absolute',
                                                left: 0,
                                                top: 0,
                                                bottom: 0,
                                                width: '2px',
                                                background: barFor(
                                                    line.type === 'add' ? 'add' : 'rem',
                                                ),
                                            }"
                                        />
                                        {{ line.newLine ?? "" }}
                                    </div>
                                    <div
                                        :style="{
                                            flex: 1,
                                            minWidth: 0,
                                            padding: '0 12px',
                                            whiteSpace: 'pre',
                                            color: 'var(--gd-text)',
                                        }"
                                    >
                                        <span
                                            :style="{
                                                display: 'inline-block',
                                                width: '12px',
                                                color: 'var(--gd-text-muted)',
                                            }"
                                            >{{
                                                sign(
                                                    line.type === "add"
                                                        ? "add"
                                                        : line.type === "del"
                                                          ? "rem"
                                                          : "ctx",
                                                )
                                            }}</span
                                        ><code
                                            data-diff-line-text
                                            v-html="highlightHtml(line.text)"
                                        />
                                    </div>
                                    <button
                                        type="button"
                                        class="gd-add-comment"
                                        title="Add comment (shift-click to extend selection)"
                                        @click="(e) => handleAddCommentClick(e, section, line)"
                                    >
                                        <Plus :size="12" :stroke-width="2.5" />
                                    </button>
                                </div>
                                <template v-for="c in commentsForLine(section, line)" :key="c.id">
                                    <CommentThread
                                        :comment="c"
                                        :reply-features="replyFeatures"
                                        @delete="emit('delete-comment', c)"
                                        @reply="(body) => emit('reply-comment', c, body)"
                                    />
                                </template>
                            </template>
                        </template>
                    </template>
                </div>
            </div>
        </section>
    </div>
</template>
