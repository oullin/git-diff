<script setup lang="ts">
import { computed } from "vue";
import type { GitFileStatus } from "@git-diff/contracts";

const props = defineProps<{ status: GitFileStatus }>();

const meta = computed(() => {
    switch (props.status) {
        case "added":
        case "untracked":
            return {
                label: "A",
                bg: "var(--success-subtle)",
                fg: "var(--success-fg)",
                title: "Added",
            };
        case "deleted":
            return {
                label: "D",
                bg: "var(--danger-subtle)",
                fg: "var(--danger-fg)",
                title: "Deleted",
            };
        case "renamed":
            return {
                label: "R",
                bg: "var(--done-subtle)",
                fg: "var(--done-fg)",
                title: "Renamed",
            };
        case "modified":
        default:
            return {
                label: "M",
                bg: "var(--attention-subtle)",
                fg: "var(--attention-fg)",
                title: "Modified",
            };
    }
});
</script>

<template>
    <span
        :title="meta.title"
        class="inline-flex items-center justify-center font-mono font-semibold"
        :style="{
            width: '18px',
            height: '18px',
            borderRadius: '4px',
            background: meta.bg,
            color: meta.fg,
            fontSize: '11px',
        }"
        >{{ meta.label }}</span
    >
</template>
