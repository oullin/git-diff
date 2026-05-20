<script setup lang="ts">
import type { WalkthroughRecord } from "@git-diff/contracts";

// WalkthroughPanel renders the AI-generated walkthrough summary, the
// ordered file list, and any per-file notes. The error case is mutually
// exclusive with the success case; both share the same chrome so we keep
// them in one component rather than two.
defineProps<{
  record: WalkthroughRecord | null;
  error: string;
}>();
</script>

<template>
  <div
    v-if="error || record"
    class="border-b border-border bg-background/70 px-4 py-2 text-xs"
  >
    <div v-if="error" class="text-red-500">{{ error }}</div>
    <div v-else-if="record" class="flex flex-col gap-1">
      <div class="font-semibold">
        AI walkthrough
        <span class="font-normal text-muted-foreground">— {{ record.modelId }}</span>
      </div>
      <div v-if="record.summary" class="text-muted-foreground">
        {{ record.summary }}
      </div>
      <ol class="mt-1 list-decimal pl-5 space-y-0.5">
        <li v-for="path in record.order" :key="path">
          <code class="font-mono text-[11px]">{{ path }}</code>
          <span v-if="record.notes[path]" class="ml-2 text-muted-foreground">
            — {{ record.notes[path] }}
          </span>
        </li>
      </ol>
    </div>
  </div>
</template>
