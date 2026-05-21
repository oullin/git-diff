<script setup lang="ts">
defineProps<{ add: number; del: number; mini?: boolean }>();
</script>

<template>
    <span
        class="inline-flex items-center gap-1 font-mono font-semibold"
        :style="{ fontSize: mini ? '10.5px' : '11px' }"
    >
        <span :style="{ color: 'var(--gd-added)' }">+{{ add }}</span>
        <span :style="{ color: 'var(--gd-removed)' }">−{{ del }}</span>
    </span>
</template>
