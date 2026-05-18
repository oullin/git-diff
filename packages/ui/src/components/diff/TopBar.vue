<script setup lang="ts">
import { ChevronDown, Filter, GitBranch, LogOut, RefreshCw, Search } from "lucide-vue-next";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@ui/dropdown-menu";
import DiffStat from "./DiffStat.vue";
import Kbd from "./Kbd.vue";
import SegGroup from "./SegGroup.vue";
import type { AuthUser, DiffViewMode, RepositoryState } from "@api";

defineProps<{
  state: RepositoryState | null;
  viewMode: DiffViewMode;
  hideWhitespace: boolean;
  searchQuery: string;
  currentUser: AuthUser | null;
  userInitials: string;
}>();

const emit = defineEmits<{
  "update:viewMode": [mode: DiffViewMode];
  "update:searchQuery": [value: string];
  "toggle-whitespace": [];
  refresh: [];
  "log-out": [];
}>();

function onSearchInput(event: Event) {
  emit("update:searchQuery", (event.target as HTMLInputElement).value);
}
</script>

<template>
  <div
    class="flex items-center"
    :style="{
      height: '56px',
      flexShrink: 0,
      gap: '10px',
      padding: '0 14px',
      borderBottom: '1px solid var(--gd-border)',
      background: 'var(--gd-bg)',
    }"
  >
    <button
      type="button"
      class="inline-flex items-center"
      :style="{
        gap: '8px',
        height: '34px',
        padding: '0 10px 0 12px',
        borderRadius: '8px',
        background: 'var(--gd-panel-2)',
        border: '1px solid var(--gd-border)',
        color: 'var(--gd-text)',
        fontSize: '13.5px',
        fontWeight: 500,
        whiteSpace: 'nowrap',
        flexShrink: 0,
        cursor: 'pointer',
      }"
    >
      <GitBranch :size="13" :style="{ color: 'var(--gd-text-3)' }" />
      <span :style="{ fontFamily: 'var(--font-mono)' }">{{ state?.branch || "detached" }}</span>
      <ChevronDown :size="12" :style="{ color: 'var(--gd-text-3)', marginLeft: '2px' }" />
    </button>

    <div class="flex items-center" :style="{ gap: '8px', paddingLeft: '4px' }">
      <span
        :style="{
          fontSize: '11.5px',
          fontFamily: 'var(--font-mono)',
          color: 'var(--gd-text-3)',
          padding: '3px 7px',
          background: 'var(--gd-panel-2)',
          borderRadius: '5px',
          border: '1px solid var(--gd-border)',
        }"
        >{{ state?.headSha?.slice(0, 8) || "—" }}</span
      >
      <span
        v-if="state"
        :style="{
          fontSize: '13px',
          color: 'var(--gd-text-2)',
          maxWidth: '320px',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
        }"
        >{{ state.files.length }} changed file{{ state.files.length === 1 ? "" : "s" }}</span
      >
      <DiffStat v-if="state" :add="state.additions" :del="state.deletions" />
    </div>

    <div class="flex-1" />

    <div
      class="flex items-center"
      :style="{
        gap: '6px',
        height: '32px',
        padding: '0 10px',
        borderRadius: '8px',
        background: 'var(--gd-panel-2)',
        border: '1px solid var(--gd-border)',
        width: '240px',
      }"
    >
      <Search :size="13" :style="{ color: 'var(--gd-text-3)' }" />
      <input
        :value="searchQuery"
        placeholder="Go to file or line…"
        :style="{
          flex: 1,
          background: 'transparent',
          border: 'none',
          outline: 'none',
          color: 'var(--gd-text)',
          fontSize: '13.5px',
        }"
        @input="onSearchInput"
      />
      <Kbd>⌘P</Kbd>
    </div>

    <SegGroup
      :model-value="viewMode"
      :options="[
        { value: 'split', label: 'Split' },
        { value: 'unified', label: 'Unified' },
      ]"
      @update:model-value="(value) => emit('update:viewMode', value as DiffViewMode)"
    />

    <button
      type="button"
      class="inline-flex items-center"
      :style="{
        gap: '6px',
        height: '30px',
        padding: '0 10px',
        borderRadius: '8px',
        border: '1px solid transparent',
        background: hideWhitespace ? 'var(--gd-hover)' : 'transparent',
        color: hideWhitespace ? 'var(--gd-text)' : 'var(--gd-text-2)',
        fontSize: '13.5px',
        fontWeight: 500,
        whiteSpace: 'nowrap',
        cursor: 'pointer',
      }"
      @click="emit('toggle-whitespace')"
    >
      <Filter :size="14" />
      Whitespace
    </button>

    <button
      type="button"
      class="inline-flex items-center"
      :style="{
        gap: '6px',
        height: '30px',
        padding: '0 10px',
        borderRadius: '8px',
        border: '1px solid transparent',
        background: 'transparent',
        color: 'var(--gd-text-2)',
        fontSize: '13.5px',
        fontWeight: 500,
        cursor: 'pointer',
      }"
      :disabled="!state"
      @click="emit('refresh')"
    >
      <RefreshCw :size="14" />
      Refresh
      <Kbd>R</Kbd>
    </button>

    <div
      :style="{ width: '1px', height: '22px', background: 'var(--gd-border)', margin: '0 2px' }"
    />

    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <button
          type="button"
          class="inline-flex items-center"
          :style="{
            gap: '8px',
            height: '32px',
            padding: '3px 10px 3px 3px',
            borderRadius: '999px',
            background: 'var(--gd-panel-2)',
            border: '1px solid var(--gd-border)',
            color: 'var(--gd-text)',
            fontSize: '13.5px',
            fontWeight: 500,
            whiteSpace: 'nowrap',
            flexShrink: 0,
            cursor: 'pointer',
          }"
        >
          <span
            :style="{
              width: '24px',
              height: '24px',
              borderRadius: '50%',
              background: 'linear-gradient(135deg, var(--gd-accent-strong), var(--gd-accent))',
              color: '#fff',
              fontSize: '11px',
              fontWeight: 600,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }"
            >{{ userInitials }}</span
          >
          <span>{{ currentUser?.osUsername || "guest" }}</span>
          <ChevronDown :size="12" :style="{ color: 'var(--gd-text-3)' }" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" class="w-[200px]">
        <DropdownMenuItem disabled>
          {{ currentUser?.displayName || currentUser?.osUsername || "Signed out" }}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem @select="emit('log-out')">
          <LogOut class="h-4 w-4" />
          Log out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
</template>
