<script setup lang="ts">
import { computed } from "vue";
import { AlertCircle, Check, MessageSquare, Search, X } from "lucide-vue-next";
import FileRow from "./FileRow.vue";
import Kbd from "./Kbd.vue";
import SegGroup from "./SegGroup.vue";
import RepoFileTree from "@entry/components/RepoFileTree.vue";
import { ScrollArea } from "@ui/scroll-area";
import type { ChangedFile, ReviewComment } from "@api";

const props = defineProps<{
  files: ChangedFile[];
  selectedPath: string;
  searchQuery: string;
  scope: "changed" | "all";
  isViewed: (file: ChangedFile) => boolean;
  threadsForFile: (path: string) => number;
  allPaths: string[];
  changedPathsSet: Set<string>;
  comments: ReviewComment[];
}>();

const emit = defineEmits<{
  "update:scope": [scope: "changed" | "all"];
  "update:searchQuery": [value: string];
  select: [path: string];
  "toggle-viewed": [file: ChangedFile];
  "start-review": [];
  "open-review-panel": [intent: "comment" | "request"];
}>();

const totalCount = computed(() => props.files.length);
const viewedCount = computed(() => props.files.filter((f) => props.isViewed(f)).length);

const filteredFiles = computed(() => {
  const q = props.searchQuery.trim().toLowerCase();
  return q ? props.files.filter((f) => f.path.toLowerCase().includes(q)) : props.files;
});

const filteredAllPaths = computed(() => {
  const q = props.searchQuery.trim().toLowerCase();
  return q ? props.allPaths.filter((path) => path.toLowerCase().includes(q)) : props.allPaths;
});

const progressPct = computed(() =>
  totalCount.value === 0 ? 0 : (viewedCount.value / totalCount.value) * 100,
);

function onSearchInput(event: Event) {
  emit("update:searchQuery", (event.target as HTMLInputElement).value);
}

function clearSearch() {
  emit("update:searchQuery", "");
}
</script>

<template>
  <div
    class="flex flex-col"
    :style="{
      width: '308px',
      flexShrink: 0,
      background:
        'linear-gradient(180deg, rgb(255 255 255 / 0.012), transparent 220px), var(--gd-bg-rail, var(--gd-panel))',
      borderRight: '1px solid var(--gd-border)',
      boxShadow: '1px 0 0 var(--gd-edge-hi-2) inset, inset -8px 0 16px -16px rgb(0 0 0 / 0.6)',
      minHeight: 0,
    }"
  >
    <div :style="{ padding: '10px 12px', borderBottom: '1px solid var(--gd-border)' }">
      <div
        class="flex items-center"
        :style="{
          gap: '6px',
          height: '30px',
          padding: '0 10px',
          borderRadius: '7px',
          background: 'var(--gd-panel-2)',
          border: '1px solid var(--gd-border)',
        }"
      >
        <Search :size="12" :style="{ color: 'var(--gd-text-3)' }" />
        <input
          :value="searchQuery"
          placeholder="Filter files"
          :style="{
            flex: 1,
            background: 'transparent',
            border: 'none',
            outline: 'none',
            color: 'var(--gd-text)',
            fontSize: '13px',
          }"
          @input="onSearchInput"
        />
        <button
          v-if="searchQuery"
          type="button"
          aria-label="Clear filter"
          :style="{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: '18px',
            height: '18px',
            borderRadius: '4px',
            border: 'none',
            background: 'transparent',
            color: 'var(--gd-text-3)',
            cursor: 'pointer',
            padding: 0,
          }"
          @click="clearSearch"
        >
          <X :size="12" />
        </button>
      </div>
      <div class="flex items-center justify-between" :style="{ marginTop: '10px' }">
        <div :style="{ fontSize: '11.5px', color: 'var(--gd-text-3)', fontWeight: 500 }">
          <span :style="{ color: 'var(--gd-text-2)' }">{{ filteredFiles.length }}</span>
          changed ·
          <span :style="{ color: 'var(--gd-text-2)' }">{{ viewedCount }}</span
          >/{{ totalCount }} viewed
        </div>
        <SegGroup
          :model-value="scope"
          :options="[
            { value: 'changed', label: 'Changed' },
            { value: 'all', label: 'All' },
          ]"
          @update:model-value="(value) => emit('update:scope', value as 'changed' | 'all')"
        />
      </div>
      <div
        :style="{
          height: '3px',
          background: 'var(--gd-panel-3)',
          borderRadius: '2px',
          marginTop: '10px',
          overflow: 'hidden',
        }"
      >
        <div
          :style="{
            width: `${progressPct}%`,
            height: '100%',
            background: 'linear-gradient(90deg, var(--gd-accent-strong), var(--gd-accent))',
            transition: 'width 0.3s',
          }"
        />
      </div>
    </div>

    <div
      :style="{
        padding: '8px 10px 4px',
        fontSize: '11px',
        fontWeight: 600,
        letterSpacing: '0.4px',
        textTransform: 'uppercase',
        color: 'var(--gd-text-muted)',
      }"
    >
      Workspace
    </div>
    <ScrollArea v-if="scope === 'changed'" class="min-h-0 flex-1">
      <div :style="{ padding: '0 6px 12px' }">
        <FileRow
          v-for="file in filteredFiles"
          :key="file.path"
          :file="file"
          :selected="selectedPath === file.path"
          :viewed="isViewed(file)"
          :threads="threadsForFile(file.path)"
          @select="emit('select', file.path)"
          @toggle-viewed="emit('toggle-viewed', file)"
        />
        <div
          v-if="filteredFiles.length === 0"
          :style="{ padding: '12px 10px', fontSize: '12px', color: 'var(--gd-text-muted)' }"
        >
          No changed files match the filter.
        </div>
      </div>
    </ScrollArea>
    <div v-else class="flex min-h-0 flex-1 flex-col" :style="{ padding: '0 6px 12px' }">
      <RepoFileTree
        v-if="filteredAllPaths.length > 0"
        :paths="filteredAllPaths"
        :selected-path="selectedPath"
        :changed-paths="changedPathsSet"
        :initial-expansion="searchQuery.trim() ? 'open' : 'closed'"
        :expand-all-on-reset="!!searchQuery.trim()"
        class="min-h-0 flex-1 p-2"
        @select="(path) => emit('select', path)"
      />
      <div
        v-else
        :style="{ padding: '12px 10px', fontSize: '12px', color: 'var(--gd-text-muted)' }"
      >
        {{ allPaths.length === 0 ? "No tracked files yet." : "No files match the filter." }}
      </div>
    </div>

    <div
      :style="{
        padding: '6px 6px 14px',
        borderTop: '1px solid var(--gd-border)',
        background: 'var(--gd-panel)',
        display: 'flex',
        flexDirection: 'column',
        gap: '14px',
      }"
    >
      <button
        type="button"
        :style="{
          height: '36px',
          borderRadius: '8px',
          border: '1px solid transparent',
          background: 'var(--gd-accent-strong)',
          color: '#fff',
          fontSize: '13.5px',
          fontWeight: 600,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: '8px',
          whiteSpace: 'nowrap',
          boxShadow:
            '0 0 0 1px var(--gd-accent-strong), 0 1px 0 rgba(255,255,255,0.08) inset, 0 6px 14px -4px var(--gd-accent-soft)',
          cursor: 'pointer',
        }"
        @click="emit('start-review')"
      >
        <Check :size="13" />
        Submit review
        <Kbd tone="on-accent">⌘↵</Kbd>
      </button>
      <div :style="{ display: 'flex', gap: '6px' }">
        <button
          type="button"
          :style="{
            flex: 1,
            height: '28px',
            borderRadius: '7px',
            border: '1px solid var(--gd-border)',
            background: 'var(--gd-panel-2)',
            color: 'var(--gd-text-2)',
            fontSize: '12px',
            fontWeight: 500,
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '5px',
            cursor: 'pointer',
          }"
          @click="emit('open-review-panel', 'comment')"
        >
          <MessageSquare :size="12" />
          Comment
        </button>
        <button
          type="button"
          :style="{
            flex: 1,
            height: '28px',
            borderRadius: '7px',
            border: '1px solid var(--gd-border)',
            background: 'var(--gd-panel-2)',
            color: 'var(--gd-text-2)',
            fontSize: '12px',
            fontWeight: 500,
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '5px',
            cursor: 'pointer',
          }"
          @click="emit('open-review-panel', 'request')"
        >
          <AlertCircle :size="12" />
          Request
        </button>
      </div>
    </div>
  </div>
</template>
