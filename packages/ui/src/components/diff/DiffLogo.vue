<script setup lang="ts">
defineProps<{ size?: number }>();
</script>

<template>
  <img
    src="/icon-192.png"
    alt="Git Diff Review"
    draggable="false"
    class="shrink-0 select-none"
    :style="{
      width: `${size ?? 22}px`,
      height: `${size ?? 22}px`,
      borderRadius: '6px',
      boxShadow: '0 0 0 1px rgba(255,255,255,0.06) inset, 0 4px 10px -2px rgba(0,0,0,0.35)',
      objectFit: 'cover',
      display: 'block',
    }"
  />
</template>
