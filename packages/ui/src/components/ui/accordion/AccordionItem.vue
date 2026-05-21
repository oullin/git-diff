<script setup lang="ts">
import type { AccordionItemProps } from "reka-ui";
import { reactiveOmit } from "@vueuse/core";
import { AccordionItem, useForwardProps } from "reka-ui";
import type { HTMLAttributes } from "vue";
import { cn } from "@lib/utils";

const props = defineProps<AccordionItemProps & { class?: HTMLAttributes["class"] }>();
const delegatedProps = reactiveOmit(props, "class");
const forwarded = useForwardProps(delegatedProps);
</script>

<template>
    <AccordionItem v-bind="forwarded" :class="cn('border-b border-border', props.class)">
        <slot />
    </AccordionItem>
</template>
