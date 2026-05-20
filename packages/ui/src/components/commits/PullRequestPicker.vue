<script setup lang="ts">
import { ref, watch } from "vue";
import { GitPullRequest, Loader2 } from "lucide-vue-next";
import { Popover, PopoverContent, PopoverTrigger } from "@ui/popover";
import type { PullRequestSummary } from "@api";

defineProps<{
  pullRequests: PullRequestSummary[];
  loading: boolean;
  activeNumber?: number;
}>();

const emit = defineEmits<{
  open: [];
  select: [number: number];
}>();

const open = ref(false);

watch(open, (next) => {
  if (next) emit("open");
});

function pick(number: number) {
  emit("select", number);
  open.value = false;
}
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <button
        type="button"
        class="inline-flex items-center gap-1.5 rounded border border-border bg-background px-2.5 py-1 text-xs font-medium hover:bg-muted"
      >
        <GitPullRequest class="h-3.5 w-3.5" />
        <span>Pull requests</span>
      </button>
    </PopoverTrigger>
    <PopoverContent class="w-[32rem] p-0" align="start">
      <header class="border-b border-border px-3 py-2 text-xs font-semibold text-muted-foreground">
        Open pull requests
      </header>
      <div
        v-if="loading"
        class="flex items-center justify-center gap-2 py-6 text-xs text-muted-foreground"
      >
        <Loader2 class="h-3.5 w-3.5 animate-spin" />
        <span>Loading…</span>
      </div>
      <ul
        v-else-if="pullRequests.length > 0"
        class="max-h-[26rem] overflow-y-auto py-1"
        role="listbox"
      >
        <li v-for="pr in pullRequests" :key="pr.number">
          <button
            type="button"
            role="option"
            :aria-selected="pr.number === activeNumber"
            class="flex w-full items-start gap-2 px-3 py-2 text-left text-xs hover:bg-muted"
            :class="pr.number === activeNumber ? 'bg-muted/60' : ''"
            @click="pick(pr.number)"
          >
            <GitPullRequest class="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <div class="truncate font-medium">#{{ pr.number }} · {{ pr.title }}</div>
              <div class="mt-0.5 flex items-center gap-2 text-[10px] text-muted-foreground">
                <span class="truncate">{{ pr.author }}</span>
                <code class="font-mono">{{ pr.baseRef }} ← {{ pr.headRef }}</code>
              </div>
            </div>
          </button>
        </li>
      </ul>
      <div v-else class="px-3 py-6 text-center text-xs text-muted-foreground">
        No open pull requests.
      </div>
    </PopoverContent>
  </Popover>
</template>
