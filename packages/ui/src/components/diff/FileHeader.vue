<script setup lang="ts">
import { computed } from "vue";
import { Check, ChevronDown, ChevronRight, Copy } from "lucide-vue-next";
import StatusBadge from "./StatusBadge.vue";
import type { ChangedFile, DiffSection } from "@api";

const props = defineProps<{
  file: ChangedFile;
  collapsed: boolean;
  viewed: boolean;
}>();

const emit = defineEmits<{
  "toggle-collapsed": [];
  "toggle-viewed": [];
  copy: [];
}>();

const parts = computed(() => props.file.path.split("/"));
const hasUnstaged = computed(() =>
  props.file.sections.some((s: DiffSection) => s.kind === "unstaged" || s.kind === "untracked"),
);
</script>

<template>
  <div
    class="flex items-center sticky top-0 z-[5]"
    :style="{
      gap: '10px',
      padding: '10px 16px',
      height: '52px',
      background: 'var(--gd-panel)',
      borderBottom: '1px solid var(--gd-border)',
    }"
  >
    <button
      type="button"
      :style="{
        width: '26px',
        height: '26px',
        borderRadius: '6px',
        background: 'transparent',
        border: '1px solid transparent',
        color: 'var(--gd-text-3)',
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 0,
        cursor: 'pointer',
      }"
      @click="emit('toggle-collapsed')"
    >
      <ChevronRight v-if="collapsed" :size="14" />
      <ChevronDown v-else :size="14" />
    </button>
    <StatusBadge :status="file.status" />
    <div
      class="flex items-center min-w-0"
      :style="{ gap: '2px', fontFamily: 'var(--font-mono)', fontSize: '13.5px' }"
    >
      <template v-for="(part, i) in parts" :key="i">
        <span v-if="i > 0" :style="{ color: 'var(--gd-text-muted)', padding: '0 2px' }">/</span>
        <span
          :style="{
            color: i === parts.length - 1 ? 'var(--gd-text)' : 'var(--gd-text-3)',
            fontWeight: i === parts.length - 1 ? 600 : 400,
            whiteSpace: 'nowrap',
          }"
          >{{ part }}</span
        >
      </template>
    </div>
    <span
      v-if="hasUnstaged"
      :style="{
        fontSize: '10.5px',
        fontWeight: 600,
        letterSpacing: '0.4px',
        textTransform: 'uppercase',
        color: 'var(--gd-warn)',
        padding: '2px 6px',
        background: 'rgb(251 191 36 / 0.10)',
        border: '1px solid rgb(251 191 36 / 0.22)',
        borderRadius: '4px',
      }"
      >Unstaged</span
    >

    <div class="flex-1" />

    <button
      type="button"
      :style="{
        height: '26px',
        padding: '0 8px',
        borderRadius: '8px',
        border: '1px solid transparent',
        background: 'transparent',
        color: 'var(--gd-text-2)',
        display: 'inline-flex',
        alignItems: 'center',
        gap: '4px',
        cursor: 'pointer',
      }"
      title="Copy path"
      @click="emit('copy')"
    >
      <Copy :size="14" />
    </button>
    <button
      type="button"
      :style="{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '6px',
        height: '28px',
        padding: '0 10px',
        borderRadius: '7px',
        background: viewed ? 'var(--gd-accent-soft)' : 'var(--gd-panel-2)',
        border: '1px solid',
        borderColor: viewed ? 'var(--gd-accent-strong)' : 'var(--gd-border)',
        color: viewed ? 'var(--gd-accent)' : 'var(--gd-text-2)',
        fontSize: '13px',
        fontWeight: 500,
        cursor: 'pointer',
      }"
      @click="emit('toggle-viewed')"
    >
      <span
        :style="{
          width: '14px',
          height: '14px',
          borderRadius: '3px',
          background: viewed ? 'var(--gd-accent)' : 'transparent',
          border: viewed ? 'none' : '1.5px solid var(--gd-border-strong)',
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#0a0a0c',
        }"
      >
        <Check v-if="viewed" :size="10" :stroke-width="3.5" />
      </span>
      {{ viewed ? "Viewed" : "Mark viewed" }}
    </button>
  </div>
</template>
