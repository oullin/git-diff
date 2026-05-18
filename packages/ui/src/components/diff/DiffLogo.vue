<script setup lang="ts">
defineProps<{ size?: number }>();
</script>

<template>
  <div
    class="relative flex flex-col items-center justify-center shrink-0"
    :style="{
      width: `${size ?? 22}px`,
      height: `${size ?? 22}px`,
      borderRadius: '6px',
      background: `linear-gradient(135deg, var(--gd-accent-strong), var(--gd-accent))`,
      boxShadow: '0 0 0 1px rgba(255,255,255,0.06) inset, 0 4px 10px -2px var(--gd-accent-soft)',
      gap: '2.5px',
    }"
  >
    <div
      :style="{
        width: '11px',
        height: '2.5px',
        borderRadius: '1px',
        background: 'rgba(255,255,255,0.92)',
      }"
    />
    <div
      :style="{
        width: '7px',
        height: '2.5px',
        borderRadius: '1px',
        background: 'rgba(255,255,255,0.55)',
      }"
    />
  </div>
</template>
