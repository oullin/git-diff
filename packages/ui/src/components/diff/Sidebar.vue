<script setup lang="ts">
import { computed, toRef } from "vue";
import { Search, X } from "lucide-vue-next";
import FileListView from "@components/diff/FileListView.vue";
import SegGroup from "@components/diff/SegGroup.vue";
import SidebarActions from "@components/diff/SidebarActions.vue";
import RepoFileTree from "@components/RepoFileTree.vue";
import { useFileListFilter } from "@composables/useFileListFilter";
import type { ChangedFile, ReviewComment } from "@git-diff/contracts";

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

const { totalCount, viewedCount, filteredFiles, filteredAllPaths, progressPct } = useFileListFilter(
    {
        files: toRef(props, "files"),
        allPaths: toRef(props, "allPaths"),
        searchQuery: toRef(props, "searchQuery"),
        isViewed: (file) => props.isViewed(file),
    },
);

const filteredCountLabel = computed(() => filteredFiles.value.length);

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
            boxShadow:
                '1px 0 0 var(--gd-edge-hi-2) inset, inset -8px 0 16px -16px var(--gd-edge-lo)',
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
                    <span :style="{ color: 'var(--gd-text-2)' }">{{ filteredCountLabel }}</span>
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
                    @update:model-value="
                        (value) => emit('update:scope', value as 'changed' | 'all')
                    "
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
                        background:
                            'linear-gradient(90deg, var(--gd-accent-strong), var(--gd-accent))',
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
        <FileListView
            v-if="scope === 'changed'"
            :files="filteredFiles"
            :selected-path="selectedPath"
            :is-viewed="isViewed"
            :threads-for-file="threadsForFile"
            @select="(path) => emit('select', path)"
            @toggle-viewed="(file) => emit('toggle-viewed', file)"
        />
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

        <SidebarActions
            @start-review="emit('start-review')"
            @open-review-panel="(intent) => emit('open-review-panel', intent)"
        />
    </div>
</template>
