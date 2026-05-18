<script setup lang="ts">
import { Keyboard } from "lucide-vue-next";
import Kbd from "./Kbd.vue";
import { languageFor } from "@lib/highlight";
import { computed } from "vue";

const props = defineProps<{
  filePath: string;
  viewedCount: number;
  total: number;
}>();

const lang = computed(() => languageFor(props.filePath) ?? "text");
</script>

<template>
  <div class="gd-statusbar">
    <span class="inline-flex items-center" :style="{ gap: '5px' }">
      <span
        :style="{ width: '7px', height: '7px', borderRadius: '50%', background: 'var(--gd-added)' }"
      />
      Synced
    </span>
    <span :style="{ color: 'var(--gd-text-muted)' }">·</span>
    <span :style="{ fontFamily: 'var(--font-mono)' }">{{ filePath || "—" }}</span>
    <span :style="{ color: 'var(--gd-text-muted)' }">·</span>
    <span>UTF-8 · LF · {{ lang }}</span>

    <div class="flex-1" />

    <span class="inline-flex items-center" :style="{ gap: '4px' }">
      <Kbd>J / K</Kbd>
      <span>Navigate</span>
    </span>
    <span class="inline-flex items-center" :style="{ gap: '4px' }">
      <Kbd>V</Kbd>
      <span>Mark viewed</span>
    </span>
    <span class="inline-flex items-center" :style="{ gap: '4px' }">
      <Kbd>C</Kbd>
      <span>Comment</span>
    </span>
    <span class="inline-flex items-center" :style="{ gap: '4px' }">
      <Kbd>⌘↵</Kbd>
      <span>Submit</span>
    </span>

    <div class="flex-1" />

    <span>{{ viewedCount }}/{{ total }} viewed</span>
    <span :style="{ color: 'var(--gd-text-muted)' }">·</span>
    <span class="inline-flex items-center" :style="{ gap: '5px' }">
      <Keyboard :size="11" /> Shortcuts
    </span>
  </div>
</template>
