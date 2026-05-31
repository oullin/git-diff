<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";

const props = defineProps<{
  class?: HTMLAttributes["class"];
}>();
</script>

<template>
  <h2 data-slot="dialog-title" :class="cn('text-lg font-semibold text-foreground', props.class)">
    <slot />
  </h2>
</template>
