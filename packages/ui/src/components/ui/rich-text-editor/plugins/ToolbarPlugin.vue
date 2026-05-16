<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from "vue";
import { useLexicalComposer } from "lexical-vue";
import {
  $getSelection,
  $isRangeSelection,
  CAN_REDO_COMMAND,
  CAN_UNDO_COMMAND,
  COMMAND_PRIORITY_LOW,
  FORMAT_TEXT_COMMAND,
  REDO_COMMAND,
  UNDO_COMMAND,
} from "lexical";
import { TOGGLE_LINK_COMMAND } from "@lexical/link";
import { INSERT_TABLE_COMMAND } from "@lexical/table";
import { mergeRegister } from "@lexical/utils";
import {
  Bold,
  Code as CodeIcon,
  ImageIcon,
  Italic,
  Link as LinkIcon,
  Redo2,
  Strikethrough,
  Table as TableIcon,
  Underline as UnderlineIcon,
  Undo2,
} from "lucide-vue-next";
import ToolbarButton from "@rich-text-editor/components/ToolbarButton.vue";
import BlockTypeMenu from "@rich-text-editor/components/BlockTypeMenu.vue";
import { useToolbarState } from "@rich-text-editor/composables/useToolbarState";
import type { ResolvedRichTextFeatures } from "@rich-text-editor/features";

type Props = {
  features?: ResolvedRichTextFeatures;
};
const props = defineProps<Props>();

type Emits = {
  "request-image": [];
};
const emit = defineEmits<Emits>();

const showTables = computed(() => props.features?.tables ?? true);
const showImages = computed(() => props.features?.images ?? true);

const editor = useLexicalComposer();
const state = useToolbarState();

const canUndo = ref(false);
const canRedo = ref(false);

const unregister = mergeRegister(
  editor.registerCommand(
    CAN_UNDO_COMMAND,
    (payload) => {
      canUndo.value = payload;
      return false;
    },
    COMMAND_PRIORITY_LOW,
  ),
  editor.registerCommand(
    CAN_REDO_COMMAND,
    (payload) => {
      canRedo.value = payload;
      return false;
    },
    COMMAND_PRIORITY_LOW,
  ),
);

onBeforeUnmount(() => unregister());

function format(type: "bold" | "italic" | "underline" | "strikethrough" | "code"): void {
  editor.dispatchCommand(FORMAT_TEXT_COMMAND, type);
}

function toggleLink(): void {
  if (state.isLink) {
    editor.dispatchCommand(TOGGLE_LINK_COMMAND, null);
    return;
  }
  editor.getEditorState().read(() => {
    const selection = $getSelection();
    if (!$isRangeSelection(selection) || selection.isCollapsed()) {
      return;
    }
    const url = window.prompt("Enter URL", "https://");
    if (!url) {
      return;
    }
    editor.dispatchCommand(TOGGLE_LINK_COMMAND, url);
  });
}

function insertTable(): void {
  editor.dispatchCommand(INSERT_TABLE_COMMAND, {
    rows: "3",
    columns: "3",
    includeHeaders: true,
  });
}
</script>

<template>
  <div
    class="flex flex-wrap items-center gap-1 border-b border-border bg-muted/40 px-2 py-1.5"
    @mousedown.prevent
  >
    <ToolbarButton
      :pressed="false"
      :disabled="!canUndo"
      title="Undo"
      @click="editor.dispatchCommand(UNDO_COMMAND, undefined)"
    >
      <Undo2 class="h-4 w-4" />
    </ToolbarButton>
    <ToolbarButton
      :pressed="false"
      :disabled="!canRedo"
      title="Redo"
      @click="editor.dispatchCommand(REDO_COMMAND, undefined)"
    >
      <Redo2 class="h-4 w-4" />
    </ToolbarButton>

    <span class="mx-1 h-6 w-px bg-border" aria-hidden="true" />

    <BlockTypeMenu :block-type="state.blockType" />

    <span class="mx-1 h-6 w-px bg-border" aria-hidden="true" />

    <ToolbarButton :pressed="state.isBold" title="Bold (Ctrl+B)" @click="format('bold')">
      <Bold class="h-4 w-4" />
    </ToolbarButton>
    <ToolbarButton :pressed="state.isItalic" title="Italic (Ctrl+I)" @click="format('italic')">
      <Italic class="h-4 w-4" />
    </ToolbarButton>
    <ToolbarButton
      :pressed="state.isUnderline"
      title="Underline (Ctrl+U)"
      @click="format('underline')"
    >
      <UnderlineIcon class="h-4 w-4" />
    </ToolbarButton>
    <ToolbarButton
      :pressed="state.isStrikethrough"
      title="Strikethrough"
      @click="format('strikethrough')"
    >
      <Strikethrough class="h-4 w-4" />
    </ToolbarButton>
    <ToolbarButton :pressed="state.isCode" title="Inline code" @click="format('code')">
      <CodeIcon class="h-4 w-4" />
    </ToolbarButton>

    <span class="mx-1 h-6 w-px bg-border" aria-hidden="true" />

    <ToolbarButton :pressed="state.isLink" title="Link" @click="toggleLink">
      <LinkIcon class="h-4 w-4" />
    </ToolbarButton>
    <ToolbarButton v-if="showTables" title="Insert table" @click="insertTable">
      <TableIcon class="h-4 w-4" />
    </ToolbarButton>
    <ToolbarButton v-if="showImages" title="Insert image" @click="emit('request-image')">
      <ImageIcon class="h-4 w-4" />
    </ToolbarButton>
  </div>
</template>
