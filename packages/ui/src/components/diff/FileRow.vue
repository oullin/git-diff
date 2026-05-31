<script setup lang="ts">
import { computed } from 'vue';
import { Check, MessageSquare } from 'lucide-vue-next';
import StatusBadge from '@diff/StatusBadge.vue';
import DiffStat from '@diff/DiffStat.vue';
import type { ChangedFile } from '@git-diff/domain';

const props = defineProps<{
	file: ChangedFile;
	selected: boolean;
	viewed: boolean;
	threads: number;
}>();

const emit = defineEmits<{
	select: [];
	'toggle-viewed': [];
}>();

const parts = computed(() => props.file.path.split('/'));

const name = computed(() => parts.value[parts.value.length - 1] ?? props.file.path);

const dir = computed(() => parts.value.slice(0, -1).join('/'));
</script>

<template>
	<div
		class="gd-row relative flex items-center cursor-pointer"
		:style="{
			gap: '8px',
			padding: '8px 8px 8px 10px',
			margin: '1px 4px',
			borderRadius: '7px',
			background: selected ? 'var(--gd-accent-soft)' : 'transparent',
			border: '1px solid',
			borderColor: 'transparent',
			boxShadow: 'none',
		}"
		@click="emit('select')"
	>
		<div
			v-if="selected"
			:style="{
				position: 'absolute',
				left: '0',
				top: '6px',
				bottom: '6px',
				width: '3px',
				background: 'var(--gd-accent)',
				borderRadius: '3px',
			}"
		/>
		<StatusBadge :status="file.status" />
		<div class="flex-1 min-w-0">
			<div class="flex items-baseline" :style="{ gap: '6px' }">
				<span
					:style="{
						fontSize: '13.5px',
						fontWeight: 500,
						color: viewed ? 'var(--gd-text-3)' : 'var(--gd-text)',
						textDecoration: viewed ? 'line-through' : 'none',
						fontFamily: 'var(--font-mono)',
						whiteSpace: 'nowrap',
						overflow: 'hidden',
						textOverflow: 'ellipsis',
					}"
					>{{ name }}</span
				>
			</div>
			<div
				:style="{
					fontSize: '11px',
					color: 'var(--gd-text-muted)',
					whiteSpace: 'nowrap',
					overflow: 'hidden',
					textOverflow: 'ellipsis',
					marginTop: '1px',
				}"
			>
				{{ dir }}
			</div>
		</div>
		<span
			v-if="threads > 0"
			:title="`${threads} comment thread`"
			class="inline-flex items-center"
			:style="{
				gap: '3px',
				fontSize: '11px',
				fontWeight: 600,
				color: 'var(--gd-accent)',
				padding: '2px 5px',
				borderRadius: '4px',
				background: 'var(--gd-accent-soft)',
			}"
		>
			<MessageSquare :size="10" />
			{{ threads }}
		</span>
		<DiffStat :add="file.additions" :del="file.deletions" mini />
		<button
			type="button"
			:title="viewed ? 'Viewed' : 'Mark viewed'"
			:style="{
				width: '16px',
				height: '16px',
				borderRadius: '4px',
				border: viewed ? 'none' : '1px solid var(--gd-border)',
				background: viewed ? 'var(--success-emphasis)' : 'transparent',
				color: '#ffffff',
				padding: 0,
				display: 'inline-flex',
				alignItems: 'center',
				justifyContent: 'center',
				cursor: 'pointer',
			}"
			@click.stop="emit('toggle-viewed')"
		>
			<Check v-if="viewed" :size="11" :stroke-width="3" />
		</button>
	</div>
</template>
