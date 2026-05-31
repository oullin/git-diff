<script setup lang="ts">
import { onMounted } from "vue";
import { useLexicalComposer } from "lexical-vue";
import type { LexicalEditor } from "lexical";

const emit = defineEmits<{ ready: [LexicalEditor] }>();

const editor = useLexicalComposer();

onMounted(() => emit("ready", editor));
</script>

<template>
  <span class="hidden" />
</template>
