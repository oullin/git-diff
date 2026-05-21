<script setup lang="ts">
import type { AccordionRootEmits, AccordionRootProps } from "reka-ui";
import { reactiveOmit } from "@vueuse/core";
import { AccordionRoot, useForwardPropsEmits } from "reka-ui";
import type { HTMLAttributes } from "vue";
import { cn } from "@lib/utils";

const props = defineProps<AccordionRootProps & { class?: HTMLAttributes["class"] }>();
const emits = defineEmits<AccordionRootEmits>();

const delegatedProps = reactiveOmit(props, "class");
const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
    <AccordionRoot v-bind="forwarded" :class="cn(props.class)">
        <slot />
    </AccordionRoot>
</template>
