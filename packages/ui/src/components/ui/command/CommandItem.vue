<script setup lang="ts">
import type { AcceptableValue, ListboxItemEmits } from "reka-ui";
import type { HTMLAttributes } from "vue";
import { ListboxItem } from "reka-ui";
import { cn } from "@lib/utils";

const props = defineProps<{
  value: AcceptableValue;
  disabled?: boolean;
  class?: HTMLAttributes["class"];
}>();

const emits = defineEmits<ListboxItemEmits>();
</script>

<template>
  <ListboxItem
    :value="props.value"
    :disabled="props.disabled"
    :class="
      cn(
        'relative flex cursor-pointer select-none items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
        props.class,
      )
    "
    @select="(event) => emits('select', event)"
  >
    <slot />
  </ListboxItem>
</template>
