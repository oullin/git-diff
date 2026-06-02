<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import { Check, ChevronDown, GitBranch, Plus } from 'lucide-vue-next';
import { useBranches } from '@composables/useBranches';
import type { RepositoryState } from '@git-diff/domain';

import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@ui/dropdown-menu';

import { Dialog, DialogBody, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@ui/dialog';

const props = defineProps<{
	state: RepositoryState | null;
	creatingBranch: boolean;
	branchCreateError: string;
}>();

const emit = defineEmits<{
	'switch-branch': [branch: string];
	'create-branch': [name: string];
}>();

const { items: branches, loading: branchesLoading, error: branchesError, load: loadBranches } = useBranches();

const createOpen = ref(false);

const createName = ref('');

const createInputRef = ref<HTMLInputElement | null>(null);

async function onBranchMenuOpen(open: boolean): Promise<void> {
	if (!open || !props.state) {
		return;
	}

	await loadBranches(props.state.root);
}

function openCreateDialog(): void {
	createName.value = '';
	createOpen.value = true;
	void nextTick(() => createInputRef.value?.focus());
}

function onCreateSubmit(): void {
	const name = createName.value.trim();

	if (!name) {
		return;
	}

	emit('create-branch', name);
}

watch(
	() => props.creatingBranch,
	(busy, wasBusy) => {
		if (wasBusy && !busy && !props.branchCreateError) {
			createOpen.value = false;
		}
	},
);
</script>

<template>
	<DropdownMenu @update:open="onBranchMenuOpen">
		<DropdownMenuTrigger as-child>
			<button
				type="button"
				class="inline-flex items-center"
				:disabled="!state"
				:style="{
					gap: '8px',
					/* Right edge aligns with the sidebar below: sidebar width minus the top bar's left padding. */
					width: 'calc(var(--gd-sidebar-w) - var(--gd-topbar-pad-x))',
					height: '34px',
					padding: '0 10px 0 12px',
					borderRadius: '8px',
					background: 'var(--gd-panel-2)',
					border: '1px solid var(--gd-border)',
					color: 'var(--gd-text)',
					fontSize: '13.5px',
					fontWeight: 500,
					whiteSpace: 'nowrap',
					flexShrink: 0,
					cursor: state ? 'pointer' : 'not-allowed',
				}"
			>
				<GitBranch :size="13" :style="{ color: 'var(--gd-text-3)', flexShrink: 0 }" />
				<span
					:style="{
						fontFamily: 'var(--font-mono)',
						flex: 1,
						minWidth: 0,
						overflow: 'hidden',
						textOverflow: 'ellipsis',
						whiteSpace: 'nowrap',
						textAlign: 'left',
					}"
					>{{ state?.branch || 'detached' }}</span
				>
				<ChevronDown :size="12" :style="{ color: 'var(--gd-text-3)', marginLeft: '2px', flexShrink: 0 }" />
			</button>
		</DropdownMenuTrigger>
		<DropdownMenuContent align="start" class="w-[260px] max-h-[320px] overflow-auto">
			<DropdownMenuItem v-if="branchesLoading" disabled>Loading…</DropdownMenuItem>
			<DropdownMenuItem v-else-if="branchesError" disabled>{{ branchesError }}</DropdownMenuItem>
			<DropdownMenuItem v-for="b in branches" :key="b" :disabled="b === state?.branch" @select="emit('switch-branch', b)">
				<GitBranch class="h-3.5 w-3.5" />
				<span class="font-mono">{{ b }}</span>
				<Check v-if="b === state?.branch" class="ml-auto h-3.5 w-3.5" />
			</DropdownMenuItem>
			<DropdownMenuSeparator />
			<DropdownMenuItem @select="openCreateDialog">
				<Plus class="h-3.5 w-3.5" />
				New branch…
			</DropdownMenuItem>
		</DropdownMenuContent>
	</DropdownMenu>

	<Dialog :show="createOpen" max-width="md" @close="createOpen = false">
		<DialogHeader>
			<DialogTitle>New branch</DialogTitle>
			<DialogDescription>
				Branch from <span class="font-mono">{{ state?.branch }}</span> and switch to it.
			</DialogDescription>
		</DialogHeader>
		<form @submit.prevent="onCreateSubmit">
			<DialogBody class="space-y-2">
				<label class="block text-xs font-medium text-muted-foreground">Branch name</label>
				<input
					ref="createInputRef"
					v-model="createName"
					type="text"
					placeholder="feature/new-thing"
					:disabled="creatingBranch"
					class="block w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm font-mono focus:outline-none focus-visible:ring-1 focus-visible:ring-ring"
				/>
				<p v-if="branchCreateError" class="text-xs text-destructive">
					{{ branchCreateError }}
				</p>
			</DialogBody>
			<DialogFooter>
				<button type="button" class="rounded-md border border-border bg-background px-3 py-1.5 text-sm" :disabled="creatingBranch" @click="createOpen = false">Cancel</button>
				<button type="submit" class="rounded-md bg-primary px-3 py-1.5 text-sm text-primary-foreground disabled:opacity-50" :disabled="!createName.trim() || creatingBranch">
					{{ creatingBranch ? 'Creating…' : 'Create' }}
				</button>
			</DialogFooter>
		</form>
	</Dialog>
</template>
