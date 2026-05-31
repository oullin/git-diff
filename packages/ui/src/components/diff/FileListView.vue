<script setup lang="ts">
import type { ChangedFile } from "@git-diff/domain";
import FileRow from "@components/diff/FileRow.vue";
import { ScrollArea } from "@ui/scroll-area";

defineProps<{
  files: ChangedFile[];
  selectedPath: string;
  isViewed: (file: ChangedFile) => boolean;
  threadsForFile: (path: string) => number;
}>();

const emit = defineEmits<{
  select: [path: string];
  "toggle-viewed": [file: ChangedFile];
}>();
</script>

<template>
  <ScrollArea class="min-h-0 flex-1">
    <div :style="{ padding: '0 6px 12px' }">
      <FileRow
        v-for="file in files"
        :key="file.path"
        :file="file"
        :selected="selectedPath === file.path"
        :viewed="isViewed(file)"
        :threads="threadsForFile(file.path)"
        @select="emit('select', file.path)"
        @toggle-viewed="emit('toggle-viewed', file)"
      />
      <div
        v-if="files.length === 0"
        :style="{
          padding: '12px 10px',
          fontSize: '12px',
          color: 'var(--gd-text-muted)',
        }"
      >
        No changed files match the filter.
      </div>
    </div>
  </ScrollArea>
</template>
