<script setup lang="ts">
import { computed, ref, watch, type HTMLAttributes } from "vue";
import {
  LexicalComposer,
  RichTextPlugin,
  HistoryPlugin,
  ListPlugin,
  LinkPlugin,
  AutoLinkPlugin,
  ClickableLinkPlugin,
  CheckListPlugin,
  MarkdownShortcutPlugin,
  DEFAULT_TRANSFORMERS,
  AutoFocusPlugin,
  TabIndentationPlugin,
  TablePlugin,
  HorizontalRulePlugin,
  SelectionAlwaysOnDisplay,
  type InitialConfigType,
} from "lexical-vue";
import type { LexicalEditor } from "lexical";
import { cn } from "@/lib/utils";
import { editorTheme } from "@rich-text-editor/theme";
import { editorNodes } from "@rich-text-editor/nodes";
import HtmlSyncPlugin from "@rich-text-editor/plugins/HtmlSyncPlugin.vue";
import ToolbarPlugin from "@rich-text-editor/plugins/ToolbarPlugin.vue";
import ImagePlugin from "@rich-text-editor/plugins/ImagePlugin.vue";
import { INSERT_IMAGE_COMMAND } from "@rich-text-editor/plugins/imageCommand";
import SlashMenuPlugin from "@rich-text-editor/plugins/SlashMenuPlugin.vue";
import MentionsPlugin from "@rich-text-editor/plugins/MentionsPlugin.vue";
import type { MentionItem } from "@rich-text-editor/plugins/mentions";
import EditorRefPlugin from "@rich-text-editor/plugins/EditorRefPlugin.vue";
import DecoratorHostPlugin from "@rich-text-editor/plugins/DecoratorHostPlugin.vue";
import LinkEditor from "@rich-text-editor/components/LinkEditor.vue";
import ImageUploadDialog from "@rich-text-editor/components/ImageUploadDialog.vue";
import EditorContentEditable from "@rich-text-editor/components/EditorContentEditable.vue";
import { resolveFeatures, type RichTextFeatures } from "@rich-text-editor/features";

type Props = {
  modelValue: string;
  placeholder?: string;
  disabled?: boolean;
  ariaLabel?: string;
  autofocus?: boolean;
  minHeight?: string;
  uploadImage?: (file: File) => Promise<string>;
  mentionLookup?: (query: string) => Promise<MentionItem[]>;
  namespace?: string;
  features?: RichTextFeatures;
  class?: HTMLAttributes["class"];
};

const props = withDefaults(defineProps<Props>(), {
  placeholder: "Start writing…",
  disabled: false,
  autofocus: false,
  minHeight: "12rem",
  namespace: "RichTextEditor",
});

const resolvedFeatures = computed(() => resolveFeatures(props.features));

const emit = defineEmits<{ "update:modelValue": [string] }>();

const initialConfig: InitialConfigType = {
  namespace: props.namespace,
  nodes: editorNodes,
  theme: editorTheme,
  editable: !props.disabled,
  onError(error) {
    // eslint-disable-next-line no-console
    console.error("[RichTextEditor]", error);
  },
};

const showImageDialog = ref(false);
const editorInstance = ref<LexicalEditor | null>(null);

watch(
  [editorInstance, () => props.disabled],
  ([editor, disabled]) => {
    editor?.setEditable(!disabled);
  },
  { immediate: true },
);

function openImageDialog(): void {
  showImageDialog.value = true;
}

function onInsertImage(payload: { src: string; altText: string }): void {
  editorInstance.value?.dispatchCommand(INSERT_IMAGE_COMMAND, {
    src: payload.src,
    altText: payload.altText,
  });
}
</script>

<template>
  <div
    :class="cn('rte-root flex flex-col rounded-md border border-input bg-background', props.class)"
  >
    <LexicalComposer :initial-config="initialConfig">
      <ToolbarPlugin :features="resolvedFeatures" @request-image="openImageDialog" />

      <div class="rte-editor-shell relative" :style="{ '--rte-min-h': props.minHeight }">
        <RichTextPlugin>
          <template #contentEditable>
            <EditorContentEditable
              class="rte-editable prose prose-sm max-w-none px-4 py-3 focus:outline-none dark:prose-invert"
              :aria-label="props.ariaLabel"
            />
          </template>
          <template #placeholder>
            <div
              class="rte-placeholder pointer-events-none absolute left-4 top-3 text-sm text-muted-foreground"
            >
              {{ props.placeholder }}
            </div>
          </template>
        </RichTextPlugin>
        <LinkEditor />
      </div>

      <HtmlSyncPlugin
        :model-value="props.modelValue"
        @update:model-value="(value) => emit('update:modelValue', value)"
      />
      <HistoryPlugin />
      <SelectionAlwaysOnDisplay />
      <ListPlugin />
      <CheckListPlugin v-if="resolvedFeatures.checklist" />
      <LinkPlugin />
      <AutoLinkPlugin :matchers="[]" />
      <ClickableLinkPlugin :disabled="false" :new-tab="true" />
      <MarkdownShortcutPlugin
        v-if="resolvedFeatures.markdownShortcuts"
        :transformers="DEFAULT_TRANSFORMERS"
      />
      <TabIndentationPlugin />
      <TablePlugin
        v-if="resolvedFeatures.tables"
        :has-cell-merge="true"
        :has-cell-background-color="false"
        :has-tab-handler="true"
        :has-horizontal-scroll="true"
      />
      <HorizontalRulePlugin />
      <ImagePlugin v-if="resolvedFeatures.images" />
      <SlashMenuPlugin
        v-if="resolvedFeatures.slashMenu"
        :features="resolvedFeatures"
        @request-image="openImageDialog"
      />
      <MentionsPlugin v-if="resolvedFeatures.mentions" :lookup="props.mentionLookup" />
      <AutoFocusPlugin v-if="props.autofocus" :default-selection="'rootEnd'" />

      <EditorRefPlugin @ready="(ed) => (editorInstance = ed)" />

      <DecoratorHostPlugin />
    </LexicalComposer>

    <ImageUploadDialog
      v-if="resolvedFeatures.images"
      :open="showImageDialog"
      :upload="props.uploadImage"
      @update:open="(value) => (showImageDialog = value)"
      @insert="onInsertImage"
    />
  </div>
</template>
