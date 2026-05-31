<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, type HTMLAttributes } from 'vue';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import type { RichTextFeatures } from '@rich-text-editor/features';
import type { MentionItem } from '@rich-text-editor/plugins/mentions';

type Props = {
	modelValue: string;
	placeholder?: string;
	disabled?: boolean;
	ariaLabel?: string;
	autofocus?: boolean;
	minHeight?: string;
	uploadImage?: (file: File) => Promise<string>;
	mentionLookup?: (query: string) => Promise<MentionItem[]>;
	namespace?: string;
	features?: RichTextFeatures;
	class?: HTMLAttributes['class'];
};

const props = withDefaults(defineProps<Props>(), {
	minHeight: '12rem',
});

const emit = defineEmits<{ 'update:modelValue': [string] }>();

const RichTextEditor = defineAsyncComponent(() => import('@rich-text-editor/RichTextEditor.vue').then((m) => m.default));

const ready = ref(false);

onMounted(() => {
	const schedule = typeof window !== 'undefined' && typeof window.requestIdleCallback === 'function' ? window.requestIdleCallback.bind(window) : (cb: () => void) => window.setTimeout(cb, 0);

	schedule(() => {
		ready.value = true;
	});
});

const skeletonStyle = computed(() => ({ minHeight: props.minHeight }));
</script>

<template>
	<RichTextEditor
		v-if="ready"
		:model-value="props.modelValue"
		:placeholder="props.placeholder"
		:disabled="props.disabled"
		:aria-label="props.ariaLabel"
		:autofocus="props.autofocus"
		:min-height="props.minHeight"
		:upload-image="props.uploadImage"
		:mention-lookup="props.mentionLookup"
		:namespace="props.namespace"
		:features="props.features"
		:class="props.class"
		@update:model-value="(value: string) => emit('update:modelValue', value)"
	/>
	<div
		v-else
		:class="cn('flex flex-col gap-2 rounded-md border border-input bg-background p-3', props.class)"
		:style="skeletonStyle"
		aria-busy="true"
		:aria-label="props.ariaLabel ?? 'Loading editor'"
	>
		<div class="flex items-center gap-2">
			<Skeleton class="h-7 w-24" />
			<Skeleton class="h-7 w-7" />
			<Skeleton class="h-7 w-7" />
			<Skeleton class="h-7 w-7" />
			<Skeleton class="h-7 w-7" />
			<Skeleton class="h-7 w-7" />
		</div>
		<div class="space-y-2 pt-2">
			<Skeleton class="h-3 w-11/12" />
			<Skeleton class="h-3 w-9/12" />
			<Skeleton class="h-3 w-10/12" />
		</div>
	</div>
</template>
