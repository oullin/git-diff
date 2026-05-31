<script setup lang="ts">
import type { AccordionContentProps } from 'reka-ui';
import { reactiveOmit } from '@vueuse/core';
import { AccordionContent, useForwardProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { cn } from '@lib/utils';

const props = defineProps<AccordionContentProps & { class?: HTMLAttributes['class'] }>();
const delegatedProps = reactiveOmit(props, 'class');
const forwarded = useForwardProps(delegatedProps);
</script>

<template>
	<AccordionContent v-bind="forwarded" class="overflow-hidden text-sm">
		<div :class="cn('pb-2 pt-0', props.class)">
			<slot />
		</div>
	</AccordionContent>
</template>
