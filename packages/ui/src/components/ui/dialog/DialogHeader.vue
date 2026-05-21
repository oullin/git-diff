<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";

const props = defineProps<{
    class?: HTMLAttributes["class"];
}>();
</script>

<template>
    <div data-slot="dialog-header" :class="cn('space-y-2 px-6 py-6 pb-0', props.class)">
        <slot />
    </div>
</template>
