<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import {
  Check,
  ChevronDown,
  GitBranch,
  Loader2,
  LogOut,
  Plus,
  RefreshCw,
  Search,
  Settings2,
} from "lucide-vue-next";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@ui/dropdown-menu";
import { Command, CommandEmpty, CommandInput, CommandItem, CommandList } from "@ui/command";
import { Popover, PopoverAnchor, PopoverContent, PopoverTrigger } from "@ui/popover";
import {
  Dialog,
  DialogBody,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@ui/dialog";
import DiffStat from "./DiffStat.vue";
import Kbd from "./Kbd.vue";
import TweaksPanel from "./TweaksPanel.vue";
import type { Tweaks } from "@composables/useTweaks";
import { useFileSearch } from "@composables/useFileSearch";
import type { AuthUser, FileSearchResult, RepositoryState } from "@git-diff/contracts";

const props = defineProps<{
  state: RepositoryState | null;
  currentUser: AuthUser | null;
  userInitials: string;
  tweaks: Tweaks;
  creatingBranch: boolean;
  branchCreateError: string;
}>();

const emit = defineEmits<{
  "update:tweak": [key: keyof Tweaks, value: Tweaks[keyof Tweaks]];
  refresh: [];
  "log-out": [];
  "switch-branch": [branch: string];
  "create-branch": [name: string];
  "select-result": [result: FileSearchResult];
}>();

const branches = ref<string[]>([]);
const branchesLoading = ref(false);
const branchesError = ref("");

const createOpen = ref(false);
const createName = ref("");
const createInputRef = ref<HTMLInputElement | null>(null);

const {
  query: searchQuery,
  results: searchResults,
  loading: searchLoading,
  error: searchError,
  open: searchOpen,
  onInput: onSearchInput,
  reset: resetSearch,
} = useFileSearch();

function basename(path: string): string {
  const slash = path.lastIndexOf("/");
  return slash >= 0 ? path.slice(slash + 1) : path;
}

function dirname(path: string): string {
  const slash = path.lastIndexOf("/");
  return slash >= 0 ? path.slice(0, slash) : "";
}

function onSelectResult(result: FileSearchResult) {
  emit("select-result", result);
  resetSearch();
}

onBeforeUnmount(() => {
  resetSearch();
});

async function onBranchMenuOpen(open: boolean) {
  if (!open || !props.state) {
    return;
  }

  branchesLoading.value = true;
  branchesError.value = "";

  try {
    const result = await window.diffApp.listBranches(props.state.root);
    branches.value = result.branches;
  } catch (cause) {
    branchesError.value = cause instanceof Error ? cause.message : String(cause);
  } finally {
    branchesLoading.value = false;
  }
}

function openCreateDialog() {
  createName.value = "";
  createOpen.value = true;
  void nextTick(() => createInputRef.value?.focus());
}

function onCreateSubmit() {
  const name = createName.value.trim();

  if (!name) {
    return;
  }

  emit("create-branch", name);
}

watch(
  () => props.creatingBranch,
  (busy, wasBusy) => {
    if (wasBusy && !busy && !props.branchCreateError) {
      createOpen.value = false;
    }
  },
);
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
    <DropdownMenu @update:open="onBranchMenuOpen">
      <DropdownMenuTrigger as-child>
        <button
          type="button"
          class="inline-flex items-center"
          :disabled="!state"
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
            cursor: state ? 'pointer' : 'not-allowed',
          }"
        >
          <GitBranch :size="13" :style="{ color: 'var(--gd-text-3)' }" />
          <span :style="{ fontFamily: 'var(--font-mono)' }">{{ state?.branch || "detached" }}</span>
          <ChevronDown :size="12" :style="{ color: 'var(--gd-text-3)', marginLeft: '2px' }" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" class="w-[260px] max-h-[320px] overflow-auto">
        <DropdownMenuItem v-if="branchesLoading" disabled>Loading…</DropdownMenuItem>
        <DropdownMenuItem v-else-if="branchesError" disabled>{{ branchesError }}</DropdownMenuItem>
        <DropdownMenuItem
          v-for="b in branches"
          :key="b"
          :disabled="b === state?.branch"
          @select="emit('switch-branch', b)"
        >
          <GitBranch class="h-3.5 w-3.5" />
          <span class="font-mono">{{ b }}</span>
          <Check v-if="b === state?.branch" class="ml-auto h-3.5 w-3.5" />
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem @select="openCreateDialog">
          <Plus class="h-3.5 w-3.5" />
          New branch…
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>

    <Dialog :show="createOpen" max-width="md" @close="createOpen = false">
      <DialogHeader>
        <DialogTitle>New branch</DialogTitle>
        <DialogDescription>
          Branch from <span class="font-mono">{{ state?.branch }}</span> and switch to it.
        </DialogDescription>
      </DialogHeader>
      <form @submit.prevent="onCreateSubmit">
        <DialogBody class="space-y-2">
          <label class="block text-xs font-medium text-muted-foreground">Branch name</label>
          <input
            ref="createInputRef"
            v-model="createName"
            type="text"
            placeholder="feature/new-thing"
            :disabled="creatingBranch"
            class="block w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm font-mono focus:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          />
          <p v-if="branchCreateError" class="text-xs text-destructive">{{ branchCreateError }}</p>
        </DialogBody>
        <DialogFooter>
          <button
            type="button"
            class="rounded-md border border-border bg-background px-3 py-1.5 text-sm"
            :disabled="creatingBranch"
            @click="createOpen = false"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="rounded-md bg-primary px-3 py-1.5 text-sm text-primary-foreground disabled:opacity-50"
            :disabled="!createName.trim() || creatingBranch"
          >
            {{ creatingBranch ? "Creating…" : "Create" }}
          </button>
        </DialogFooter>
      </form>
    </Dialog>

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

    <div
      :style="{ width: '1px', height: '22px', background: 'var(--gd-border)', margin: '0 6px' }"
    />

    <Popover :open="searchOpen">
      <PopoverAnchor as-child>
        <Command
          :model-value="undefined"
          class="flex flex-row items-center"
          :style="{
            flex: 1,
            minWidth: 0,
            gap: '6px',
            height: '32px',
            padding: '0 10px',
            borderRadius: '8px',
            background: 'var(--gd-panel-2)',
            border: '1px solid var(--gd-border)',
          }"
          @update:model-value="(value: any) => value && onSelectResult(value as FileSearchResult)"
        >
          <Search :size="13" :style="{ color: 'var(--gd-text-3)', flexShrink: 0 }" />
          <CommandInput
            :model-value="searchQuery"
            placeholder="Search files across all repos…"
            :style="{
              flex: 1,
              minWidth: 0,
              background: 'transparent',
              color: 'var(--gd-text)',
              fontSize: '13.5px',
            }"
            @update:model-value="onSearchInput"
          />
          <Loader2
            v-if="searchLoading"
            :size="13"
            class="animate-spin"
            :style="{ color: 'var(--gd-text-3)', flexShrink: 0 }"
          />
          <Kbd>⌘P</Kbd>

          <PopoverContent
            align="start"
            :side-offset="6"
            class="p-0 w-[var(--reka-popover-trigger-width)] max-w-[640px]"
            @open-auto-focus="(event: Event) => event.preventDefault()"
          >
            <CommandList class="max-h-[360px]">
              <CommandEmpty v-if="searchError">{{ searchError }}</CommandEmpty>
              <CommandEmpty v-else-if="!searchLoading && searchResults.length === 0">
                No matching files
              </CommandEmpty>
              <CommandItem
                v-for="result in searchResults"
                :key="`${result.repoPath}::${result.filePath}`"
                :value="result"
                class="flex items-center gap-2"
              >
                <span class="font-mono text-sm truncate">{{ basename(result.filePath) }}</span>
                <span
                  v-if="dirname(result.filePath)"
                  class="font-mono text-xs text-muted-foreground truncate"
                  >{{ dirname(result.filePath) }}</span
                >
                <span class="ml-auto pl-2 text-xs text-muted-foreground shrink-0">
                  {{ result.repoName }}
                </span>
              </CommandItem>
            </CommandList>
          </PopoverContent>
        </Command>
      </PopoverAnchor>
    </Popover>

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

    <Popover>
      <PopoverTrigger as-child>
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
            whiteSpace: 'nowrap',
            cursor: 'pointer',
          }"
        >
          <Settings2 :size="14" />
          Tweaks
        </button>
      </PopoverTrigger>
      <PopoverContent
        align="end"
        :side-offset="8"
        class="w-[280px] p-0 border-0 shadow-none bg-transparent"
      >
        <div
          :style="{
            width: '280px',
            background: 'var(--gd-panel)',
            border: '1px solid var(--gd-border)',
            borderRadius: '10px',
            boxShadow: 'var(--gd-shadow-lg)',
            color: 'var(--gd-text)',
            fontFamily: 'var(--font-sans)',
            fontSize: '13px',
            overflow: 'hidden',
          }"
        >
          <div
            :style="{
              padding: '10px 12px',
              borderBottom: '1px solid var(--gd-border)',
              fontSize: '13px',
              fontWeight: 600,
              color: 'var(--gd-text)',
            }"
          >
            Tweaks
          </div>
          <div :style="{ padding: '12px' }">
            <TweaksPanel
              :tweaks="tweaks"
              @update:tweak="(key, value) => emit('update:tweak', key, value)"
            />
          </div>
        </div>
      </PopoverContent>
    </Popover>

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
