<script setup lang="ts">
import { computed } from 'vue';
import { ChevronDown, ChevronUp, Loader2 } from 'lucide-vue-next';

const props = defineProps<{
	text: string;
	canUp: boolean;
	canDown: boolean;
	upInflight: boolean;
	downInflight: boolean;
}>();

defineEmits<{
	'expand-up': [];
	'expand-down': [];
}>();

const header = computed(() => {
	const parts = props.text.split('@@');

	if (parts.length < 3) {
		return { range: props.text, trailer: '' };
	}

	return {
		range: `@@${parts[1]}@@`,
		trailer: parts.slice(2).join('@@').trim(),
	};
});

function buttonStyle(enabled: boolean, inflight: boolean) {
	return {
		width: '26px',
		height: '26px',
		borderRadius: '6px',
		border: '1px solid transparent',
		background: 'transparent',
		color: 'var(--gd-text-3)',
		display: 'inline-flex',
		alignItems: 'center',
		justifyContent: 'center',
		padding: 0,
		cursor: enabled && !inflight ? 'pointer' : 'not-allowed',
		opacity: enabled ? 1 : 0.35,
	} as const;
}
</script>

<template>
	<div
		class="flex items-center"
		data-hunk-anchor
		:style="{
			gap: '10px',
			padding: '8px 16px',
			background: 'var(--accent-subtle)',
			color: 'var(--gd-text-3)',
			fontSize: '13px',
			borderTop: '1px solid var(--gd-border-soft)',
			borderBottom: '1px solid var(--gd-border-soft)',
			boxShadow: '0 1px 0 var(--gd-edge-hi-2) inset',
		}"
	>
		<ChevronDown :size="12" />
		<span :style="{ fontFamily: 'var(--font-mono)', color: 'var(--gd-accent)' }">{{ header.range }}</span>
		<span :style="{ color: 'var(--gd-text-3)' }">{{ header.trailer }}</span>
		<div class="flex-1" />
		<button type="button" title="Expand context up" :disabled="!canUp || upInflight" :style="buttonStyle(canUp, upInflight)" @click="$emit('expand-up')">
			<Loader2 v-if="upInflight" :size="12" class="animate-spin" />
			<ChevronUp v-else :size="12" />
		</button>
		<button type="button" title="Expand context down" :disabled="!canDown || downInflight" :style="buttonStyle(canDown, downInflight)" @click="$emit('expand-down')">
			<Loader2 v-if="downInflight" :size="12" class="animate-spin" />
			<ChevronDown v-else :size="12" />
		</button>
	</div>
</template>
