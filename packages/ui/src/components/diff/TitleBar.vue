<script setup lang="ts">
import { computed } from "vue";
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
import type { AuthUser, Repository, RepositoryState } from "@api";

const props = defineProps<{
  state: RepositoryState | null;
  repositories: Repository[];
  activeRepoPath: string;
  currentUser: AuthUser | null;
}>();

const emit = defineEmits<{
  "select-repo": [path: string];
  "add-repo": [];
  "remove-repo": [path: string];
}>();

const workspaceLabel = computed(() => {
  const root = props.state?.root ?? props.activeRepoPath;
  if (!root) return "Select repository";
  const parts = root.split("/").filter(Boolean);
  return parts[parts.length - 1] ?? root;
});

const userInitials = computed(() => {
  const name = props.currentUser?.displayName ?? props.currentUser?.osUsername ?? "GO";
  return (
    name
      .split(/[\s_-]+/)
      .map((part) => part[0]?.toUpperCase() ?? "")
      .slice(0, 2)
      .join("") || "GO"
  );
});

const branchSummary = computed(() => {
  const s = props.state;
  if (!s) return "—";
  return `${s.branch || "detached"} · ${s.headSha?.slice(0, 8) || "no HEAD"}`;
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
    <div class="flex items-center" style="gap: 8px">
      <span class="gd-traffic-light red" />
      <span class="gd-traffic-light yellow" />
      <span class="gd-traffic-light green" />
    </div>

    <div
      :style="{ width: '1px', height: '16px', background: 'var(--gd-border)', marginLeft: '4px' }"
    />

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

    <DropdownMenu>
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
        <DropdownMenuSeparator />
        <DropdownMenuItem @select="emit('add-repo')">
          <Plus class="h-4 w-4" />
          Add repository…
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>

    <div class="flex-1" />

    <div v-if="state" class="flex items-center" :style="{ gap: '8px' }" data-no-drag>
      <span
        :style="{
          width: '22px',
          height: '22px',
          borderRadius: '50%',
          background: 'linear-gradient(135deg, var(--gd-accent-strong), var(--gd-accent))',
          color: '#fff',
          fontSize: '10px',
          fontWeight: 600,
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: '0 0 0 1px var(--gd-border)',
        }"
        >{{ userInitials }}</span
      >
      <span :style="{ fontSize: '12.5px', color: 'var(--gd-text-2)', whiteSpace: 'nowrap' }">
        <span :style="{ color: 'var(--gd-text)', fontWeight: 500 }">{{
          currentUser?.displayName || currentUser?.osUsername || "—"
        }}</span>
        <span :style="{ color: 'var(--gd-text-muted)', margin: '0 6px' }">·</span>
        <span :style="{ color: 'var(--gd-text-3)' }">{{ branchSummary }}</span>
      </span>
    </div>
  </div>
</template>
