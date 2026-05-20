<script setup lang="ts">
import { ref, watch } from "vue";
import { GitCommitHorizontal, History, Loader2 } from "lucide-vue-next";
import { Popover, PopoverContent, PopoverTrigger } from "@ui/popover";
import type { CommitSummary } from "@api";

const props = defineProps<{
  commits: CommitSummary[];
  loading: boolean;
  activeSha?: string;
  mode: "working" | "commit";
}>();

const emit = defineEmits<{
  open: [];
  select: [sha: string];
  back: [];
}>();

const open = ref(false);

watch(open, (next) => {
  if (next) emit("open");
});

function pick(sha: string) {
  emit("select", sha);
  open.value = false;
}

function formatDate(iso: string): string {
  if (!iso) return "";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  return date.toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
</script>

<template>
  <div class="flex items-center gap-1">
    <Popover v-model:open="open">
      <PopoverTrigger as-child>
        <button
          type="button"
          class="inline-flex items-center gap-1.5 rounded border border-border bg-background px-2.5 py-1 text-xs font-medium hover:bg-muted"
        >
          <History class="h-3.5 w-3.5" />
          <span>{{ mode === "commit" ? "Browse commits" : "History" }}</span>
        </button>
      </PopoverTrigger>
      <PopoverContent class="w-[28rem] p-0" align="start">
        <header
          class="border-b border-border px-3 py-2 text-xs font-semibold text-muted-foreground"
        >
          Recent commits
        </header>
        <div
          v-if="loading"
          class="flex items-center justify-center gap-2 py-6 text-xs text-muted-foreground"
        >
          <Loader2 class="h-3.5 w-3.5 animate-spin" />
          <span>Loading…</span>
        </div>
        <ul
          v-else-if="commits.length > 0"
          class="max-h-[26rem] overflow-y-auto py-1"
          role="listbox"
        >
          <li v-for="commit in commits" :key="commit.sha">
            <button
              type="button"
              role="option"
              :aria-selected="commit.sha === activeSha"
              class="flex w-full items-start gap-2 px-3 py-2 text-left text-xs hover:bg-muted"
              :class="commit.sha === activeSha ? 'bg-muted/60' : ''"
              @click="pick(commit.sha)"
            >
              <GitCommitHorizontal class="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
              <div class="min-w-0 flex-1">
                <div class="truncate font-medium">{{ commit.subject }}</div>
                <div class="mt-0.5 flex items-center gap-2 text-[10px] text-muted-foreground">
                  <code class="font-mono">{{ commit.shortSha }}</code>
                  <span class="truncate">{{ commit.author }}</span>
                  <span class="shrink-0">{{ formatDate(commit.date) }}</span>
                </div>
              </div>
            </button>
          </li>
        </ul>
        <div v-else class="px-3 py-6 text-center text-xs text-muted-foreground">
          No commits found.
        </div>
      </PopoverContent>
    </Popover>
    <button
      v-if="mode === 'commit'"
      type="button"
      class="inline-flex items-center gap-1.5 rounded border border-border bg-background px-2.5 py-1 text-xs font-medium hover:bg-muted"
      @click="emit('back')"
    >
      Back to working tree
    </button>
  </div>
</template>
