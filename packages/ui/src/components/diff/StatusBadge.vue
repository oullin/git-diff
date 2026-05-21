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
                bg: "rgb(74 222 128 / 0.14)",
                fg: "var(--gd-added)",
                title: "Added",
            };
        case "deleted":
            return {
                label: "D",
                bg: "rgb(248 113 113 / 0.14)",
                fg: "var(--gd-removed)",
                title: "Deleted",
            };
        case "renamed":
            return {
                label: "R",
                bg: "rgb(129 140 248 / 0.14)",
                fg: "var(--gd-accent)",
                title: "Renamed",
            };
        case "modified":
        default:
            return {
                label: "M",
                bg: "rgb(251 191 36 / 0.14)",
                fg: "var(--gd-warn)",
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
