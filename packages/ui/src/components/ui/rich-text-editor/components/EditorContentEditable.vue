<script setup lang="ts">
import { useLexicalComposer } from "lexical-vue";
import type { AriaAttributes, HTMLAttributes } from "vue";
import { onMounted, onUnmounted, ref } from "vue";

defineOptions({
  inheritAttrs: false,
});

type Props = {
  ariaLabel?: AriaAttributes["aria-label"];
  class?: HTMLAttributes["class"];
  role?: HTMLAttributes["role"];
  spellcheck?: HTMLAttributes["spellcheck"];
  tabindex?: HTMLAttributes["tabindex"];
};

const props = withDefaults(defineProps<Props>(), {
  role: "textbox",
  spellcheck: true,
});

const editor = useLexicalComposer();

const root = ref<HTMLDivElement | null>(null);

const isEditable = ref(editor.isEditable());

let unregisterEditableListener: (() => void) | undefined;

onMounted(() => {
  if (root.value?.ownerDocument.defaultView) {
    editor.setRootElement(root.value);
  }

  isEditable.value = editor.isEditable();
  unregisterEditableListener = editor.registerEditableListener((currentIsEditable) => {
    isEditable.value = currentIsEditable;
  });
});

onUnmounted(() => {
  unregisterEditableListener?.();
  editor.setRootElement(null);
});
</script>

<template>
  <div
    ref="root"
    :aria-label="props.ariaLabel"
    :aria-readonly="isEditable ? undefined : true"
    :class="props.class"
    :contenteditable="isEditable"
    :role="props.role"
    :spellcheck="props.spellcheck"
    :tabindex="props.tabindex"
  />
</template>
