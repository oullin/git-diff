<script setup lang="ts">
import { computed, ref, toRef } from "vue";
import { Plus } from "lucide-vue-next";
import { languageFor } from "@lib/highlight";
import { useContextExpansion } from "@composables/useContextExpansion";
import { useContextExpansionControls } from "@composables/useContextExpansionControls";
import { useDiffHighlighting } from "@composables/useDiffHighlighting";
import { useDiffStyles } from "@composables/useDiffStyles";
import { useHunkInfo } from "@composables/useHunkInfo";
import { useLineComments, lineSideAndNumber } from "@composables/useLineComments";
import { useSplitCellRender } from "@composables/useSplitCellRender";
import CommentThread from "@diff/CommentThread.vue";
import SplitHandle from "@diff/SplitHandle.vue";
import DiffGutter from "@diff/DiffGutter.vue";
import DiffCodeCell from "@diff/DiffCodeCell.vue";
import DiffHunkHeader from "@diff/DiffHunkHeader.vue";
import DiffSplitRow from "@diff/DiffSplitRow.vue";
import type { RichTextFeatures } from "@ui/rich-text-editor";

import {
  applyExpansions,
  parsePatch,
  splitPatchLines,
  type HunkInfo,
  type PatchLine,
  type SplitRow,
} from "@git-diff/domain/diff";

import {
  useLineSelection,
  type LineAnchor,
  type LineSelectionRange,
  type LineSide,
} from "@composables/useLineSelection";

import type {
  ChangedFile,
  DiffHunkStyle,
  DiffSection,
  DiffViewMode,
  ReviewComment,
} from "@git-diff/domain";

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
    hideResolved?: boolean;
  }>(),
  { splitRatio: 0.5, commitRef: undefined, hideResolved: false },
);

const { getExpansions, isInflight, isDownwardEof, expandUp, expandDown } = useContextExpansion();
const { setAnchor, clear: clearAnchor, rangeTo } = useLineSelection();

const emit = defineEmits<{
  "add-comment": [section: DiffSection, line: PatchLine, range?: LineSelectionRange];
  "delete-comment": [comment: ReviewComment];
  "reply-comment": [parent: ReviewComment, bodyHtml: string];
  "resolve-comment": [comment: ReviewComment, resolved: boolean];
  "update:splitRatio": [value: number];
}>();

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

const { canExpandUp, canExpandDown, onExpandUp, onExpandDown } = useContextExpansionControls({
  getExpansions,
  isDownwardEof,
  nextHunkOldStart,
  expandUp,
  expandDown,
  repoRoot: () => props.repoRoot,
  filePath: () => props.file.path,
  commitRef: () => props.commitRef,
});

const { commentsForLine } = useLineComments({
  comments: () => props.comments,
  filePath: () => props.file.path,
});

const fileComments = computed(() => props.comments.filter((c) => c.filePath === props.file.path));

const presentAnchors = computed(() => {
  const set = new Set<string>();

  for (const section of props.file.sections) {
    for (const line of patchLines(section)) {
      if (line.oldLine != null) {
        set.add(`${section.kind}:left:${line.oldLine}`);
      }

      if (line.newLine != null) {
        set.add(`${section.kind}:right:${line.newLine}`);
      }
    }
  }

  return set;
});

const outdatedIds = computed(() => {
  const set = new Set<number>();

  for (const c of fileComments.value) {
    if (!presentAnchors.value.has(`${c.diffSection}:${c.side}:${c.lineNumber}`)) {
      set.add(c.id);
    }
  }

  return set;
});

const outdatedComments = computed(() => {
  if (props.hideResolved) {
    return [];
  }

  return fileComments.value.filter((c) => outdatedIds.value.has(c.id));
});

// Inline comments, minus resolved ones when the hide tweak is on.
function renderableComments(section: DiffSection, line: PatchLine | undefined): ReviewComment[] {
  return commentsForLine(section, line).filter((c) => !(props.hideResolved && c.resolved));
}

const { renderPair } = useSplitCellRender({
  wordHighlight: () => props.wordHighlight,
  highlightHtml,
  withRanges,
  computeWordHi,
});
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
                <DiffHunkHeader
                  v-if="row.line.text.startsWith('@@')"
                  :text="row.line.text"
                  :can-up="canExpandUp(section, findHunk(section, row.line.id))"
                  :can-down="canExpandDown(section, findHunk(section, row.line.id))"
                  :up-inflight="isInflight(section.id, 'up', row.line.id)"
                  :down-inflight="isInflight(section.id, 'down', row.line.id)"
                  @expand-up="onExpandUp(section, findHunk(section, row.line.id))"
                  @expand-down="onExpandDown(section, findHunk(section, row.line.id))"
                />
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
                    <DiffGutter :num="row.line.oldLine ?? ''" />
                    <DiffCodeCell :html="highlightHtml(row.line.text)" />
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
                    <DiffGutter :num="row.line.newLine ?? ''" />
                    <DiffCodeCell :html="highlightHtml(row.line.text)" />
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
                <template v-for="c in renderableComments(section, row.line)" :key="c.id">
                  <CommentThread
                    :comment="c"
                    :reply-features="replyFeatures"
                    @delete="emit('delete-comment', c)"
                    @reply="(body) => emit('reply-comment', c, body)"
                    @resolve="(r) => emit('resolve-comment', c, r)"
                  />
                </template>
              </template>
              <template v-else>
                <DiffSplitRow
                  :render="renderPair(row)"
                  :line-height="lineH"
                  :diff-style="diffStyle"
                  :can-comment="!!(row.left || row.right)"
                  @add-comment="
                    (e) => handleAddCommentClick(e, section, (row.right ?? row.left) as PatchLine)
                  "
                />
                <template
                  v-for="line in [row.left, row.right].filter(Boolean) as PatchLine[]"
                  :key="`${row.id}:${line.id}`"
                >
                  <CommentThread
                    v-for="c in renderableComments(section, line)"
                    :key="c.id"
                    :comment="c"
                    :reply-features="replyFeatures"
                    @delete="emit('delete-comment', c)"
                    @reply="(body) => emit('reply-comment', c, body)"
                    @resolve="(r) => emit('resolve-comment', c, r)"
                  />
                </template>
              </template>
            </template>
          </template>
          <template v-else>
            <template v-for="line in patchLines(section)" :key="line.id">
              <DiffHunkHeader
                v-if="line.type === 'meta' && line.text.startsWith('@@')"
                :text="line.text"
                :can-up="canExpandUp(section, findHunk(section, line.id))"
                :can-down="canExpandDown(section, findHunk(section, line.id))"
                :up-inflight="isInflight(section.id, 'up', line.id)"
                :down-inflight="isInflight(section.id, 'down', line.id)"
                @expand-up="onExpandUp(section, findHunk(section, line.id))"
                @expand-down="onExpandDown(section, findHunk(section, line.id))"
              />
              <template v-else-if="line.type !== 'meta'">
                <div
                  class="gd-row relative flex"
                  :style="{
                    background: bgFor(
                      line.type === 'add' ? 'add' : line.type === 'del' ? 'rem' : 'ctx',
                    ),
                    minHeight: `${lineH}px`,
                  }"
                >
                  <DiffGutter
                    :num="line.oldLine ?? ''"
                    :bg="
                      numBgFor(line.type === 'add' ? 'add' : line.type === 'del' ? 'rem' : 'ctx')
                    "
                  />
                  <DiffGutter
                    :num="line.newLine ?? ''"
                    :bg="
                      numBgFor(line.type === 'add' ? 'add' : line.type === 'del' ? 'rem' : 'ctx')
                    "
                    :bar="
                      line.type === 'add' || line.type === 'del'
                        ? barFor(line.type === 'add' ? 'add' : 'rem')
                        : undefined
                    "
                  />
                  <DiffCodeCell
                    :html="highlightHtml(line.text)"
                    :sign="sign(line.type === 'add' ? 'add' : line.type === 'del' ? 'rem' : 'ctx')"
                  />
                  <button
                    type="button"
                    class="gd-add-comment"
                    title="Add comment (shift-click to extend selection)"
                    @click="(e) => handleAddCommentClick(e, section, line)"
                  >
                    <Plus :size="12" :stroke-width="2.5" />
                  </button>
                </div>
                <template v-for="c in renderableComments(section, line)" :key="c.id">
                  <CommentThread
                    :comment="c"
                    :reply-features="replyFeatures"
                    @delete="emit('delete-comment', c)"
                    @reply="(body) => emit('reply-comment', c, body)"
                    @resolve="(r) => emit('resolve-comment', c, r)"
                  />
                </template>
              </template>
            </template>
          </template>
        </div>
      </div>
    </section>
    <div v-if="outdatedComments.length">
      <CommentThread
        v-for="c in outdatedComments"
        :key="`outdated-${c.id}`"
        :comment="c"
        :reply-features="replyFeatures"
        :outdated="true"
        @delete="emit('delete-comment', c)"
        @reply="(body) => emit('reply-comment', c, body)"
        @resolve="(r) => emit('resolve-comment', c, r)"
      />
    </div>
  </div>
</template>
