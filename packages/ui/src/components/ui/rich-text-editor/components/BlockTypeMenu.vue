<script setup lang="ts">
import { computed } from "vue";
import { useLexicalComposer } from "lexical-vue";
import { $createParagraphNode, $getSelection, $isRangeSelection } from "lexical";
import { $setBlocksType } from "@lexical/selection";
import { $createHeadingNode, $createQuoteNode, type HeadingTagType } from "@lexical/rich-text";
import { $createCodeNode } from "@lexical/code";
import {
  INSERT_ORDERED_LIST_COMMAND,
  INSERT_UNORDERED_LIST_COMMAND,
  REMOVE_LIST_COMMAND,
} from "@lexical/list";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@components/ui/dropdown-menu";
import {
  ChevronDown,
  Code,
  Heading1,
  Heading2,
  Heading3,
  List,
  ListOrdered,
  Pilcrow,
  Quote,
} from "lucide-vue-next";
import type { BlockType } from "@rich-text-editor/composables/useToolbarState";

type Props = {
  blockType: BlockType;
};

const props = defineProps<Props>();
const editor = useLexicalComposer();

const labelMap: Record<BlockType, { label: string; Icon: typeof Pilcrow }> = {
  paragraph: { label: "Paragraph", Icon: Pilcrow },
  h1: { label: "Heading 1", Icon: Heading1 },
  h2: { label: "Heading 2", Icon: Heading2 },
  h3: { label: "Heading 3", Icon: Heading3 },
  h4: { label: "Heading 4", Icon: Heading3 },
  h5: { label: "Heading 5", Icon: Heading3 },
  h6: { label: "Heading 6", Icon: Heading3 },
  quote: { label: "Quote", Icon: Quote },
  code: { label: "Code block", Icon: Code },
  bullet: { label: "Bulleted list", Icon: List },
  number: { label: "Numbered list", Icon: ListOrdered },
  check: { label: "Checklist", Icon: List },
};

const current = computed(() => labelMap[props.blockType] ?? labelMap.paragraph);

function setParagraph(): void {
  if (props.blockType === "bullet" || props.blockType === "number" || props.blockType === "check") {
    editor.dispatchCommand(REMOVE_LIST_COMMAND, undefined);
    return;
  }
  editor.update(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      $setBlocksType(selection, () => $createParagraphNode());
    }
  });
}

function setHeading(level: HeadingTagType): void {
  editor.update(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      $setBlocksType(selection, () => $createHeadingNode(level));
    }
  });
}

function setQuote(): void {
  editor.update(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      $setBlocksType(selection, () => $createQuoteNode());
    }
  });
}

function setCode(): void {
  editor.update(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      if (selection.isCollapsed()) {
        $setBlocksType(selection, () => $createCodeNode());
      } else {
        const text = selection.getTextContent();
        const code = $createCodeNode();
        selection.insertNodes([code]);
        if (text) {
          code.select();
          selection.insertRawText(text);
        }
      }
    }
  });
}

function setBullet(): void {
  if (props.blockType === "bullet") {
    editor.dispatchCommand(REMOVE_LIST_COMMAND, undefined);
  } else {
    editor.dispatchCommand(INSERT_UNORDERED_LIST_COMMAND, undefined);
  }
}

function setNumber(): void {
  if (props.blockType === "number") {
    editor.dispatchCommand(REMOVE_LIST_COMMAND, undefined);
  } else {
    editor.dispatchCommand(INSERT_ORDERED_LIST_COMMAND, undefined);
  }
}
</script>

<template>
  <DropdownMenu :modal="false">
    <DropdownMenuTrigger
      class="inline-flex h-8 min-w-[10rem] items-center justify-between gap-2 rounded-md border border-input bg-background px-2 text-sm hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
    >
      <span class="inline-flex items-center gap-2">
        <component :is="current.Icon" class="h-4 w-4" />
        <span>{{ current.label }}</span>
      </span>
      <ChevronDown class="h-3.5 w-3.5 opacity-60" />
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start" class="w-48">
      <DropdownMenuItem @select="setParagraph">
        <Pilcrow class="mr-2 h-4 w-4" />Paragraph
      </DropdownMenuItem>
      <DropdownMenuItem @select="() => setHeading('h1')">
        <Heading1 class="mr-2 h-4 w-4" />Heading 1
      </DropdownMenuItem>
      <DropdownMenuItem @select="() => setHeading('h2')">
        <Heading2 class="mr-2 h-4 w-4" />Heading 2
      </DropdownMenuItem>
      <DropdownMenuItem @select="() => setHeading('h3')">
        <Heading3 class="mr-2 h-4 w-4" />Heading 3
      </DropdownMenuItem>
      <DropdownMenuItem @select="setBullet">
        <List class="mr-2 h-4 w-4" />Bulleted list
      </DropdownMenuItem>
      <DropdownMenuItem @select="setNumber">
        <ListOrdered class="mr-2 h-4 w-4" />Numbered list
      </DropdownMenuItem>
      <DropdownMenuItem @select="setQuote"> <Quote class="mr-2 h-4 w-4" />Quote </DropdownMenuItem>
      <DropdownMenuItem @select="setCode">
        <Code class="mr-2 h-4 w-4" />Code block
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
