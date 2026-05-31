<script setup lang="ts">
import { toRef } from "vue";
import { Plus } from "lucide-vue-next";
import type { DiffHunkStyle } from "@git-diff/domain";
import { useDiffStyles } from "@composables/useDiffStyles";
import type { CellRender } from "@composables/useSplitCellRender";
import DiffGutter from "@diff/DiffGutter.vue";
import DiffCodeCell from "@diff/DiffCodeCell.vue";

const props = defineProps<{
  render: { left: CellRender; right: CellRender };
  lineHeight: number;
  diffStyle: DiffHunkStyle;
  canComment: boolean;
}>();

defineEmits<{
  "add-comment": [event: MouseEvent];
}>();

const { bgFor, numBgFor, barFor, sign } = useDiffStyles(toRef(props, "diffStyle"));
</script>

<template>
  <div
    class="gd-row relative"
    :style="{ display: 'grid', gridTemplateColumns: 'var(--gd-split-cols, 1fr 1fr)' }"
  >
    <div
      class="flex"
      :style="{
        background: bgFor(render.left.kind),
        minHeight: `${lineHeight}px`,
        minWidth: 0,
        overflow: 'hidden',
      }"
    >
      <DiffGutter
        :num="render.left.num"
        :bg="numBgFor(render.left.kind)"
        :bar="render.left.kind === 'rem' ? barFor('rem') : undefined"
      />
      <DiffCodeCell :html="render.left.html" :sign="sign(render.left.kind)" />
    </div>
    <div
      class="flex"
      :style="{
        background: bgFor(render.right.kind),
        borderLeft: '1px solid var(--gd-border-soft)',
        minHeight: `${lineHeight}px`,
        minWidth: 0,
        overflow: 'hidden',
      }"
    >
      <DiffGutter
        :num="render.right.num"
        :bg="numBgFor(render.right.kind)"
        :bar="render.right.kind === 'add' ? barFor('add') : undefined"
      />
      <DiffCodeCell :html="render.right.html" :sign="sign(render.right.kind)" />
    </div>
    <button
      v-if="canComment"
      type="button"
      class="gd-add-comment"
      title="Add comment (shift-click to extend selection)"
      @click="(e) => $emit('add-comment', e)"
    >
      <Plus :size="12" :stroke-width="2.5" />
    </button>
  </div>
</template>
