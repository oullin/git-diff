<script setup lang="ts">
import { ref, onBeforeUnmount } from "vue";
import { GripVertical } from "lucide-vue-next";

const props = defineProps<{
  ratio: number;
  containerEl: HTMLElement | null;
  min?: number;
  max?: number;
}>();

const emit = defineEmits<{
  "update:ratio": [value: number];
}>();

const dragging = ref(false);

function clamp(value: number): number {
  const min = props.min ?? 0.2;
  const max = props.max ?? 0.8;
  return Math.min(max, Math.max(min, value));
}

function onPointerMove(event: PointerEvent): void {
  const el = props.containerEl;
  if (!el) return;
  const rect = el.getBoundingClientRect();
  if (rect.width <= 0) return;
  const next = (event.clientX - rect.left) / rect.width;
  emit("update:ratio", clamp(next));
}

function endDrag(): void {
  dragging.value = false;
  document.removeEventListener("pointermove", onPointerMove);
  document.removeEventListener("pointerup", endDrag);
  document.removeEventListener("pointercancel", endDrag);
  document.body.style.removeProperty("user-select");
  document.body.style.removeProperty("cursor");
}

function startDrag(event: PointerEvent): void {
  event.preventDefault();
  dragging.value = true;
  document.addEventListener("pointermove", onPointerMove);
  document.addEventListener("pointerup", endDrag);
  document.addEventListener("pointercancel", endDrag);
  document.body.style.userSelect = "none";
  document.body.style.cursor = "col-resize";
}

onBeforeUnmount(endDrag);
</script>

<template>
  <div
    class="gd-split-handle"
    :class="{ 'gd-split-handle--active': dragging }"
    :style="{ left: `calc(${ratio * 100}% - 3px)` }"
    role="separator"
    aria-orientation="vertical"
    :aria-valuenow="Math.round(ratio * 100)"
    @pointerdown="startDrag"
  >
    <div class="gd-split-handle__line" />
    <div class="gd-split-handle__grip">
      <GripVertical :size="10" />
    </div>
  </div>
</template>

<style scoped>
.gd-split-handle {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 6px;
  z-index: 5;
  cursor: col-resize;
  display: flex;
  align-items: stretch;
  justify-content: center;
  touch-action: none;
}

.gd-split-handle__line {
  width: 1px;
  background: var(--gd-border-soft);
  opacity: 0;
  transition:
    opacity 120ms ease,
    background 120ms ease,
    width 120ms ease;
}

.gd-split-handle:hover .gd-split-handle__line,
.gd-split-handle--active .gd-split-handle__line {
  width: 2px;
  opacity: 1;
  background: var(--gd-accent, var(--gd-border-soft));
}

.gd-split-handle__grip {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 22px;
  border-radius: 4px;
  background: var(--gd-panel);
  border: 1px solid var(--gd-border-soft);
  color: var(--gd-text-3);
  opacity: 0;
  pointer-events: none;
  transition: opacity 120ms ease;
}

.gd-split-handle:hover .gd-split-handle__grip,
.gd-split-handle--active .gd-split-handle__grip {
  opacity: 1;
}
</style>
