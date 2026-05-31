<script setup lang="ts">
defineProps<{ tone?: "default" | "on-accent" }>();
</script>

<template>
  <span :class="['gd-kbd', tone === 'on-accent' && 'gd-kbd--on-accent']">
    <slot />
  </span>
</template>
