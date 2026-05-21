<script setup lang="ts">
import type { AcceptableValue, ListboxRootEmits, ListboxRootProps } from "reka-ui";
import type { HTMLAttributes } from "vue";
import { reactiveOmit } from "@vueuse/core";
import { ListboxRoot, useForwardPropsEmits } from "reka-ui";
import { cn } from "@lib/utils";

const props = defineProps<
    ListboxRootProps<AcceptableValue> & { class?: HTMLAttributes["class"] }
>();
const emits = defineEmits<ListboxRootEmits<AcceptableValue>>();

const delegatedProps = reactiveOmit(props, "class");
const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
    <ListboxRoot
        v-bind="forwarded"
        :class="
            cn(
                'flex h-full w-full flex-col overflow-hidden rounded-md bg-popover text-popover-foreground',
                props.class,
            )
        "
    >
        <slot />
    </ListboxRoot>
</template>
