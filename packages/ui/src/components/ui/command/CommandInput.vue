<script setup lang="ts">
import type { ListboxFilterProps } from "reka-ui";
import type { HTMLAttributes } from "vue";
import { reactiveOmit } from "@vueuse/core";
import { ListboxFilter, useForwardProps } from "reka-ui";
import { cn } from "@lib/utils";

const props = defineProps<ListboxFilterProps & { class?: HTMLAttributes["class"] }>();
const emits = defineEmits<{ "update:modelValue": [value: string] }>();

const delegatedProps = reactiveOmit(props, "class");
const forwarded = useForwardProps(delegatedProps);
</script>

<template>
    <ListboxFilter
        v-bind="forwarded"
        :class="
            cn(
                'flex w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50',
                props.class,
            )
        "
        @update:model-value="(value) => emits('update:modelValue', value)"
    />
</template>
