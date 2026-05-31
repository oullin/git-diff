<script setup lang="ts">
import { computed, ref, watch } from "vue";

import type {
  WalkthroughAction,
  WalkthroughGroup,
  WalkthroughImpact,
  WalkthroughRecord,
} from "@git-diff/domain";

// Backwards compatible: when `groups` is empty but legacy `order`/`notes`
// are present (cached row from before phase 5), we hoist them into a
// single synthetic group so renderings stay non-empty.
const props = defineProps<{
  record: WalkthroughRecord | null;
  error: string;
}>();

const groups = computed<WalkthroughGroup[]>(() => {
  if (!props.record) {
    return [];
  }

  if (props.record.groups?.length) {
    return props.record.groups;
  }

  const order = props.record.order ?? [];
  const notes = props.record.notes ?? {};

  if (order.length === 0) {
    return [];
  }

  return [
    {
      id: "files",
      title: "Files",
      rationale: "",
      files: order.map((path) => ({
        path,
        note: notes[path] ?? "",
        action: "review" as WalkthroughAction,
        impact: "contained" as WalkthroughImpact,
      })),
    },
  ];
});

const collapsedByFingerprint = ref<Record<string, Set<string>>>({});

watch(
  () => props.record?.fingerprint,
  (fp) => {
    if (fp && !collapsedByFingerprint.value[fp]) {
      collapsedByFingerprint.value[fp] = new Set();
    }
  },
  { immediate: true },
);

function isCollapsed(groupId: string): boolean {
  const fp = props.record?.fingerprint;

  if (!fp) {
    return false;
  }

  return collapsedByFingerprint.value[fp]?.has(groupId) ?? false;
}

function toggleGroup(groupId: string): void {
  const fp = props.record?.fingerprint;

  if (!fp) {
    return;
  }

  const set = collapsedByFingerprint.value[fp] ?? new Set();

  if (set.has(groupId)) {
    set.delete(groupId);
  } else {
    set.add(groupId);
  }

  collapsedByFingerprint.value[fp] = new Set(set);
}

const actionStyles: Record<WalkthroughAction, string> = {
  review: "bg-amber-500/15 text-amber-700 dark:text-amber-300",
  scan: "bg-sky-500/15 text-sky-700 dark:text-sky-300",
  skim: "bg-zinc-500/15 text-zinc-700 dark:text-zinc-300",
};

const impactStyles: Record<WalkthroughImpact, string> = {
  wide: "bg-rose-500/15 text-rose-700 dark:text-rose-300",
  contained: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300",
  mechanical: "bg-zinc-500/10 text-zinc-600 dark:text-zinc-400",
};
</script>

<template>
  <div v-if="error || record" class="border-b border-border bg-background/70 px-4 py-2 text-xs">
    <div v-if="error" class="text-red-500">{{ error }}</div>
    <div v-else-if="record" class="flex flex-col gap-2">
      <div class="font-semibold">
        AI walkthrough
        <span class="font-normal text-muted-foreground">
          — {{ record.modelId }}
          <span v-if="record.providerId" class="opacity-60"> ({{ record.providerId }}) </span>
        </span>
      </div>
      <div v-if="record.summary" class="text-muted-foreground">
        {{ record.summary }}
      </div>
      <div v-for="group in groups" :key="group.id" class="rounded border border-border/60">
        <button
          type="button"
          class="flex w-full items-center gap-2 px-2 py-1 text-left hover:bg-accent/40"
          @click="toggleGroup(group.id)"
        >
          <span class="font-mono text-[10px] text-muted-foreground">
            {{ isCollapsed(group.id) ? "▶" : "▼" }}
          </span>
          <span class="font-medium">{{ group.title }}</span>
          <span class="ml-auto text-[11px] text-muted-foreground">
            {{ group.files.length }} file{{ group.files.length === 1 ? "" : "s" }}
          </span>
        </button>
        <div v-if="!isCollapsed(group.id)" class="px-3 py-2">
          <div v-if="group.rationale" class="mb-2 text-muted-foreground">
            {{ group.rationale }}
          </div>
          <ol class="list-decimal pl-5 space-y-1">
            <li
              v-for="file in group.files"
              :key="file.path"
              class="flex flex-wrap items-baseline gap-1"
            >
              <code class="font-mono text-[11px]">{{ file.path }}</code>
              <span
                class="ml-1 rounded px-1 py-px text-[10px] font-medium uppercase"
                :class="actionStyles[file.action] ?? actionStyles.review"
              >
                {{ file.action }}
              </span>
              <span
                class="rounded px-1 py-px text-[10px] font-medium uppercase"
                :class="impactStyles[file.impact] ?? impactStyles.contained"
              >
                {{ file.impact }}
              </span>
              <span v-if="file.note" class="ml-1 text-muted-foreground"> — {{ file.note }} </span>
            </li>
          </ol>
        </div>
      </div>
    </div>
  </div>
</template>
