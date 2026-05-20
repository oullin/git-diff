<script setup lang="ts">
import { computed } from "vue";
import { resolvedTheme } from "@composables/useTheme";

defineProps<{ size?: number }>();

const iconSrc = computed(() =>
  resolvedTheme.value === "light" ? "/icon-192-light.png" : "/icon-192.png",
);
</script>

<template>
  <img
    :src="iconSrc"
    alt="Git Diff Review"
    draggable="false"
    class="shrink-0 select-none"
    :style="{
      width: `${size ?? 22}px`,
      height: `${size ?? 22}px`,
      borderRadius: '6px',
      boxShadow: 'var(--gd-shadow-card)',
      objectFit: 'cover',
      display: 'block',
    }"
  />
</template>
