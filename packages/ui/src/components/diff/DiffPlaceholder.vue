<script setup lang="ts">
import { computed } from "vue";
import type { ChangedFile } from "@git-diff/contracts";
import { estimatedDiffHeight } from "@composables/useLazyDiffFile";

const props = defineProps<{ file: ChangedFile }>();
const emit = defineEmits<{ load: [] }>();

const lineCount = computed(() => props.file.additions + props.file.deletions);
const minHeight = computed(() => `${estimatedDiffHeight(lineCount.value)}px`);
</script>

<template>
    <div
        class="grid place-items-center"
        :style="{
            minHeight,
            background: 'var(--gd-bg-code, var(--gd-bg))',
            borderTop: '1px solid var(--gd-border)',
        }"
    >
        <button
            type="button"
            class="flex flex-col items-center gap-1"
            :style="{
                background: 'transparent',
                border: 0,
                color: 'var(--gd-text-3)',
                fontSize: '12.5px',
                cursor: 'pointer',
            }"
            @click="emit('load')"
        >
            <span>Large file — scroll to load ({{ lineCount.toLocaleString() }} lines)</span>
            <span :style="{ fontSize: '11.5px', color: 'var(--gd-text-muted)' }">Load now</span>
        </button>
    </div>
</template>
