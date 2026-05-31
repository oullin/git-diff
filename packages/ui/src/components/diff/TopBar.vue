<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next';
import BranchPicker from '@components/diff/BranchPicker.vue';
import DiffStat from '@components/diff/DiffStat.vue';
import FileSearchPopover from '@components/diff/FileSearchPopover.vue';
import Kbd from '@components/diff/Kbd.vue';
import type { FileSearchResult, RepositoryState } from '@git-diff/domain';

defineProps<{
	state: RepositoryState | null;
	creatingBranch: boolean;
	branchCreateError: string;
}>();

const emit = defineEmits<{
	refresh: [];
	'switch-branch': [branch: string];
	'create-branch': [name: string];
	'select-result': [result: FileSearchResult];
}>();
</script>

<template>
	<div
		class="flex items-center"
		:style="{
			height: '48px',
			flexShrink: 0,
			gap: '10px',
			padding: '0 14px',
			borderBottom: '1px solid var(--gd-border)',
			background: 'var(--gd-bg)',
		}"
	>
		<BranchPicker
			:state="state"
			:creating-branch="creatingBranch"
			:branch-create-error="branchCreateError"
			@switch-branch="(branch) => emit('switch-branch', branch)"
			@create-branch="(name) => emit('create-branch', name)"
		/>

		<div class="flex items-center" :style="{ gap: '8px', paddingLeft: '4px' }">
			<span
				:style="{
					fontSize: '11.5px',
					fontFamily: 'var(--font-mono)',
					color: 'var(--gd-text-3)',
					padding: '3px 7px',
					background: 'var(--gd-panel-2)',
					borderRadius: '5px',
					border: '1px solid var(--gd-border)',
				}"
				>{{ state?.headSha?.slice(0, 8) || '—' }}</span
			>
			<span
				v-if="state"
				:style="{
					fontSize: '13px',
					color: 'var(--gd-text-2)',
					maxWidth: '320px',
					overflow: 'hidden',
					textOverflow: 'ellipsis',
					whiteSpace: 'nowrap',
				}"
				>{{ state.files.length }} changed file{{ state.files.length === 1 ? '' : 's' }}</span
			>
			<DiffStat v-if="state" :add="state.additions" :del="state.deletions" />
		</div>

		<div class="flex-1" />

		<FileSearchPopover @select-result="(result) => emit('select-result', result)" />

		<div class="flex-1" />

		<button
			type="button"
			class="inline-flex items-center"
			:style="{
				gap: '6px',
				height: '30px',
				padding: '0 10px',
				borderRadius: '8px',
				border: '1px solid transparent',
				background: 'transparent',
				color: 'var(--gd-text-2)',
				fontSize: '13.5px',
				fontWeight: 500,
				cursor: 'pointer',
			}"
			:disabled="!state"
			@click="emit('refresh')"
		>
			<RefreshCw :size="14" />
			Refresh
			<Kbd>R</Kbd>
		</button>
	</div>
</template>
