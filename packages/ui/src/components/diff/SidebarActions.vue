<script setup lang="ts">
import { Check, Eye, MessageSquare } from 'lucide-vue-next';
import Kbd from '@components/diff/Kbd.vue';

const emit = defineEmits<{
	'start-review': [];
}>();

const secondaryButton = {
	flex: 1,
	height: '32px',
	borderRadius: '6px',
	border: '1px solid var(--gd-border)',
	background: 'var(--gd-bg)',
	color: 'var(--gd-text)',
	fontSize: '13px',
	fontWeight: 500,
	display: 'flex',
	alignItems: 'center',
	justifyContent: 'center',
	gap: '6px',
	cursor: 'pointer',
} as const;
</script>

<template>
	<div
		:style="{
			padding: '6px 6px 14px',
			borderTop: '1px solid var(--gd-border)',
			background: 'var(--gd-panel)',
			display: 'flex',
			flexDirection: 'column',
			gap: '14px',
		}"
	>
		<button
			type="button"
			:style="{
				height: '36px',
				borderRadius: '8px',
				border: '1px solid rgba(31, 35, 40, 0.15)',
				background: 'var(--success)',
				color: '#fff',
				fontSize: '13.5px',
				fontWeight: 600,
				display: 'flex',
				alignItems: 'center',
				justifyContent: 'center',
				gap: '8px',
				whiteSpace: 'nowrap',
				cursor: 'pointer',
			}"
			@click="emit('start-review')"
		>
			<Check :size="13" />
			Submit review
			<Kbd tone="on-accent">⌘↵</Kbd>
		</button>

		<div :style="{ display: 'flex', gap: '8px' }">
			<button type="button" :style="secondaryButton">
				<MessageSquare :size="14" />
				Comment
			</button>
			<button type="button" :style="secondaryButton">
				<Eye :size="14" />
				Request
			</button>
		</div>
	</div>
</template>
