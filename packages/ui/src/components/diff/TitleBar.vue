<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { Check, ChevronDown, Folder, Plus, Trash2 } from "lucide-vue-next";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@ui/dropdown-menu";
import DiffLogo from "./DiffLogo.vue";
import { Skeleton } from "@ui/skeleton";
import type { Repository, RepositoryState, SystemStats } from "@api";

const props = defineProps<{
  state: RepositoryState | null;
  repositories: Repository[];
  repositoriesLoading: boolean;
  activeRepoPath: string;
}>();

const emit = defineEmits<{
  "select-repo": [path: string];
  "add-repo": [];
  "remove-repo": [path: string];
  "refresh-repos": [];
}>();

function onRepoMenuOpen(open: boolean) {
  if (open) emit("refresh-repos");
}

const workspaceLabel = computed(() => {
  const root = props.state?.root ?? props.activeRepoPath;
  if (!root) return "Select repository";
  const parts = root.split("/").filter(Boolean);
  return parts[parts.length - 1] ?? root;
});

const STATS_INTERVAL_MS = 5000;

const stats = ref<SystemStats | null>(null);
let statsTimer: ReturnType<typeof setInterval> | null = null;

async function refreshStats() {
  try {
    stats.value = await window.diffApp.getSystemStats();
  } catch {
    /* stats are non-critical — leave the previous value in place */
  }
}

function startPolling() {
  if (statsTimer) return;
  refreshStats();
  statsTimer = setInterval(refreshStats, STATS_INTERVAL_MS);
}

function stopPolling() {
  if (!statsTimer) return;
  clearInterval(statsTimer);
  statsTimer = null;
}

function handleVisibilityChange() {
  if (document.hidden) {
    stopPolling();
  } else {
    startPolling();
  }
}

onMounted(() => {
  if (!document.hidden) startPolling();
  document.addEventListener("visibilitychange", handleVisibilityChange);
});

onBeforeUnmount(() => {
  document.removeEventListener("visibilitychange", handleVisibilityChange);
  stopPolling();
});
</script>

<template>
  <div
    class="gd-titlebar relative flex items-center"
    :style="{
      height: '44px',
      gap: '12px',
      padding: '0 14px',
      borderBottom: '1px solid var(--gd-border)',
      background: 'linear-gradient(180deg, var(--gd-panel) 0%, var(--gd-bg) 100%)',
      flexShrink: 0,
    }"
  >
    <div class="flex items-center" style="gap: 9px" data-no-drag>
      <DiffLogo />
      <span
        :style="{
          fontSize: '14px',
          fontWeight: 700,
          color: 'var(--gd-text)',
          letterSpacing: '-0.1px',
          whiteSpace: 'nowrap',
        }"
        >Git Diff Review</span
      >
      <span
        :style="{
          fontSize: '10px',
          fontWeight: 600,
          letterSpacing: '0.6px',
          textTransform: 'uppercase',
          color: 'var(--gd-text-3)',
          padding: '2px 6px',
          background: 'var(--gd-panel-2)',
          border: '1px solid var(--gd-border)',
          borderRadius: '4px',
          marginLeft: '2px',
        }"
        >Beta</span
      >
    </div>

    <DropdownMenu @update:open="onRepoMenuOpen">
      <DropdownMenuTrigger as-child>
        <button
          type="button"
          class="inline-flex items-center"
          :style="{
            gap: '6px',
            paddingLeft: '6px',
            background: 'transparent',
            border: 0,
            color: 'var(--gd-text-2)',
            cursor: 'pointer',
            whiteSpace: 'nowrap',
          }"
          :title="state?.root ?? ''"
          data-no-drag
        >
          <span :style="{ color: 'var(--gd-text-muted)', fontSize: '13px' }">/</span>
          <Folder :size="12" :style="{ color: 'var(--gd-text-3)' }" />
          <span
            :style="{
              fontSize: '13px',
              color: 'var(--gd-text-2)',
              fontFamily: 'var(--font-mono)',
              whiteSpace: 'nowrap',
            }"
            >{{ workspaceLabel }}</span
          >
          <ChevronDown :size="12" :style="{ color: 'var(--gd-text-3)' }" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" class="w-[320px]">
        <DropdownMenuLabel>Repositories</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <template v-if="repositoriesLoading">
          <div aria-busy="true" aria-label="Loading repositories">
            <DropdownMenuItem
              v-for="(width, i) in ['70%', '55%', '80%']"
              :key="`repo-skeleton-${i}`"
              class="group flex items-center gap-2"
              disabled
              aria-hidden="true"
              @select.prevent
            >
              <span class="h-3.5 w-3.5 shrink-0" />
              <Skeleton class="h-4 min-w-0 flex-1" :style="{ width }" />
              <span class="h-3.5 w-3.5 shrink-0 opacity-0" />
            </DropdownMenuItem>
          </div>
        </template>
        <template v-else>
          <div v-if="repositories.length === 0" class="px-2 py-3 text-xs text-muted-foreground">
            No repositories yet.
          </div>
          <DropdownMenuItem
            v-for="repo in repositories"
            :key="repo.path"
            class="group flex items-center gap-2"
            @select="emit('select-repo', repo.path)"
          >
            <Check
              :class="[
                'h-3.5 w-3.5 shrink-0',
                activeRepoPath === repo.path ? 'opacity-100' : 'opacity-0',
              ]"
            />
            <span class="min-w-0 flex-1 truncate" :title="repo.path">{{ repo.name }}</span>
            <button
              class="icon-btn opacity-0 group-hover:opacity-100"
              type="button"
              title="Remove from list"
              @click.stop="emit('remove-repo', repo.path)"
            >
              <Trash2 class="h-3.5 w-3.5" />
            </button>
          </DropdownMenuItem>
        </template>
        <DropdownMenuSeparator />
        <DropdownMenuItem @select="emit('add-repo')">
          <Plus class="h-4 w-4" />
          Add repository…
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>

    <div class="flex-1" />

    <div
      v-if="stats"
      class="flex items-center"
      :style="{
        gap: '10px',
        fontFamily: 'var(--font-mono)',
        fontSize: '12px',
        color: 'var(--gd-text-2)',
        whiteSpace: 'nowrap',
      }"
      data-no-drag
    >
      <span>
        <span :style="{ color: 'var(--gd-text-muted)' }">CPU</span>
        {{ stats.cpuPercent.toFixed(0) }}%
      </span>
      <span :style="{ color: 'var(--gd-text-muted)' }">·</span>
      <span>
        <span :style="{ color: 'var(--gd-text-muted)' }">MEM</span>
        {{ stats.memoryUsedGB.toFixed(1) }} / {{ stats.memoryTotalGB.toFixed(1) }} GB
      </span>
      <span :style="{ color: 'var(--gd-text-muted)' }">·</span>
      <span>
        <span :style="{ color: 'var(--gd-text-muted)' }">LOAD</span>
        {{ stats.loadAvg1.toFixed(2) }}
      </span>
    </div>
    <div
      v-else
      class="flex items-center"
      :style="{
        gap: '10px',
        fontFamily: 'var(--font-mono)',
        fontSize: '12px',
        color: 'var(--gd-text-muted)',
        whiteSpace: 'nowrap',
      }"
      data-no-drag
      aria-busy="true"
      aria-label="Loading system stats"
    >
      <span class="flex items-center" :style="{ gap: '6px' }">
        <span>CPU</span>
        <Skeleton :style="{ width: '28px', height: '10px' }" />
      </span>
      <span>·</span>
      <span class="flex items-center" :style="{ gap: '6px' }">
        <span>MEM</span>
        <Skeleton :style="{ width: '78px', height: '10px' }" />
      </span>
      <span>·</span>
      <span class="flex items-center" :style="{ gap: '6px' }">
        <span>LOAD</span>
        <Skeleton :style="{ width: '32px', height: '10px' }" />
      </span>
    </div>
  </div>
</template>
