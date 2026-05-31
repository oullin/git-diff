<script setup lang="ts">
import { computed, type HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";

type Props = {
  pressed?: boolean;
  disabled?: boolean;
  title?: string;
  ariaLabel?: string;
  class?: HTMLAttributes["class"];
};

const props = withDefaults(defineProps<Props>(), {
  pressed: false,
  disabled: false,
});

const emit = defineEmits<{ click: [MouseEvent] }>();

const classes = computed(() =>
  cn(
    "inline-flex h-8 w-8 items-center justify-center rounded-md text-sm transition-colors",
    "hover:bg-accent hover:text-accent-foreground",
    "focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
    "disabled:pointer-events-none disabled:opacity-50",
    props.pressed && "bg-accent text-accent-foreground",
    props.class,
  ),
);

function handleClick(event: MouseEvent): void {
  event.preventDefault();
  if (props.disabled) {
    return;
  }

  emit("click", event);
}
</script>

<template>
  <button
    type="button"
    :class="classes"
    :disabled="disabled"
    :title="title"
    :aria-label="ariaLabel ?? title"
    :aria-pressed="pressed"
    @mousedown.prevent
    @click="handleClick"
  >
    <slot />
  </button>
</template>
