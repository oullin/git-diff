<script setup lang="ts">
import { computed, ref } from "vue";
import type { RichTextFeatures } from "@ui/rich-text-editor";
import type { LineSelectionRange } from "@composables/useLineSelection";
import type { PatchLine } from "@git-diff/domain/diff";
import DiffBody from "@entry/components/diff/DiffBody.vue";
import DiffPlaceholder from "@entry/components/diff/DiffPlaceholder.vue";
import { LAZY_DIFF_LINE_THRESHOLD, useLazyDiffFile } from "@composables/useLazyDiffFile";
import { forceRenderDiffFile } from "@composables/useLazyRender";

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
    hideResolved?: boolean;
    comments: ReviewComment[];
    replyFeatures: RichTextFeatures;
    repoRoot: string;
    commitRef?: string;
    splitRatio?: number;
  }>(),
  { splitRatio: 0.5, commitRef: undefined, hideResolved: false },
);

const emit = defineEmits<{
  "add-comment": [section: DiffSection, line: PatchLine, range?: LineSelectionRange];
  "delete-comment": [comment: ReviewComment];
  "reply-comment": [parent: ReviewComment, bodyHtml: string];
  "resolve-comment": [comment: ReviewComment, resolved: boolean];
  "update:splitRatio": [value: number];
}>();

const root = ref<HTMLElement | null>(null);

const shouldDefer = computed(
  () => props.file.additions + props.file.deletions > LAZY_DIFF_LINE_THRESHOLD,
);

const { rendered } = useLazyDiffFile(root, {
  path: computed(() => props.file.path),
  shouldDefer,
});

function loadNow() {
  forceRenderDiffFile(props.file.path);
}
</script>

<template>
  <div ref="root">
    <DiffBody
      v-if="rendered"
      :file="file"
      :view-mode="viewMode"
      :diff-style="diffStyle"
      :density="density"
      :word-highlight="wordHighlight"
      :hide-whitespace="hideWhitespace"
      :hide-resolved="hideResolved"
      :comments="comments"
      :reply-features="replyFeatures"
      :repo-root="repoRoot"
      :commit-ref="commitRef"
      :split-ratio="splitRatio"
      @add-comment="(section, line, range) => emit('add-comment', section, line, range)"
      @delete-comment="(comment) => emit('delete-comment', comment)"
      @reply-comment="(parent, bodyHtml) => emit('reply-comment', parent, bodyHtml)"
      @resolve-comment="(comment, resolved) => emit('resolve-comment', comment, resolved)"
      @update:split-ratio="(value) => emit('update:splitRatio', value)"
    />
    <DiffPlaceholder v-else :file="file" @load="loadNow" />
  </div>
</template>
