<script setup lang="ts">
import { computed } from 'vue';
import { Check, ChevronDown, Folder, Plus, Settings2, Trash2 } from 'lucide-vue-next';
import DiffLogo from '@diff/DiffLogo.vue';
import TweaksPanel from '@components/diff/TweaksPanel.vue';
import UserMenu from '@components/diff/UserMenu.vue';
import { Popover, PopoverContent, PopoverTrigger } from '@ui/popover';
import { Skeleton } from '@ui/skeleton';
import type { Tweaks } from '@composables/useTweaks';
import type { AuthUser, Repository, RepositoryState } from '@git-diff/domain';

import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '@ui/dropdown-menu';

const props = defineProps<{
	state: RepositoryState | null;
	repositories: Repository[];
	repositoriesLoading: boolean;
	activeRepoPath: string;
	currentUser: AuthUser | null;
	userInitials: string;
	tweaks: Tweaks;
}>();

const emit = defineEmits<{
	'select-repo': [path: string];
	'add-repo': [];
	'remove-repo': [path: string];
	'refresh-repos': [];
	'update:tweak': [key: keyof Tweaks, value: Tweaks[keyof Tweaks]];
	'log-out': [];
}>();

function onRepoMenuOpen(open: boolean) {
	if (open) {
		emit('refresh-repos');
	}
}

const workspaceLabel = computed(() => {
	const root = props.state?.root ?? props.activeRepoPath;

	if (!root) {
		return 'Select repository';
	}

	const parts = root.split('/').filter(Boolean);

	return parts[parts.length - 1] ?? root;
});
</script>

<template>
	<div
		class="gd-titlebar relative flex items-center"
		:style="{
			height: '56px',
			gap: '12px',
			padding: '0 14px',
			borderBottom: '1px solid var(--gd-border)',
			background: 'var(--gd-bg)',
			flexShrink: 0,
		}"
	>
		<div class="flex items-center" style="gap: 9px" data-no-drag>
			<DiffLogo />
			<span
				:style="{
					fontSize: '14px',
					fontWeight: 700,
					color: 'var(--gd-text)',
					letterSpacing: '-0.1px',
					whiteSpace: 'nowrap',
				}"
				>Git Diff Review</span
			>
			<span
				:style="{
					fontSize: '10px',
					fontWeight: 600,
					letterSpacing: '0.6px',
					textTransform: 'uppercase',
					color: 'var(--gd-text-3)',
					padding: '2px 6px',
					background: 'var(--gd-panel-2)',
					border: '1px solid var(--gd-border)',
					borderRadius: '4px',
					marginLeft: '2px',
				}"
				>Beta</span
			>
		</div>

		<DropdownMenu @update:open="onRepoMenuOpen">
			<DropdownMenuTrigger as-child>
				<button
					type="button"
					class="inline-flex items-center"
					:style="{
						gap: '6px',
						paddingLeft: '6px',
						background: 'transparent',
						border: 0,
						color: 'var(--gd-text-2)',
						cursor: 'pointer',
						whiteSpace: 'nowrap',
					}"
					:title="state?.root ?? ''"
					data-no-drag
				>
					<span :style="{ color: 'var(--gd-text-muted)', fontSize: '13px' }">/</span>
					<Folder :size="12" :style="{ color: 'var(--gd-text-3)' }" />
					<span
						:style="{
							fontSize: '13px',
							color: 'var(--gd-text-2)',
							fontFamily: 'var(--font-mono)',
							whiteSpace: 'nowrap',
						}"
						>{{ workspaceLabel }}</span
					>
					<ChevronDown :size="12" :style="{ color: 'var(--gd-text-3)' }" />
				</button>
			</DropdownMenuTrigger>
			<DropdownMenuContent align="start" class="w-[320px]">
				<DropdownMenuLabel>Repositories</DropdownMenuLabel>
				<DropdownMenuSeparator />
				<template v-if="repositoriesLoading">
					<div aria-busy="true" aria-label="Loading repositories">
						<DropdownMenuItem v-for="(width, i) in ['70%', '55%', '80%']" :key="`repo-skeleton-${i}`" class="group flex items-center gap-2" disabled aria-hidden="true" @select.prevent>
							<span class="h-3.5 w-3.5 shrink-0" />
							<Skeleton class="h-4 min-w-0 flex-1" :style="{ width }" />
							<span class="h-3.5 w-3.5 shrink-0 opacity-0" />
						</DropdownMenuItem>
					</div>
				</template>
				<template v-else>
					<div v-if="repositories.length === 0" class="px-2 py-3 text-xs text-muted-foreground">No repositories yet.</div>
					<DropdownMenuItem v-for="repo in repositories" :key="repo.path" class="group flex items-center gap-2" @select="emit('select-repo', repo.path)">
						<Check :class="['h-3.5 w-3.5 shrink-0', activeRepoPath === repo.path ? 'opacity-100' : 'opacity-0']" />
						<span class="min-w-0 flex-1 truncate" :title="repo.path">{{ repo.name }}</span>
						<button class="icon-btn opacity-0 group-hover:opacity-100" type="button" title="Remove from list" @click.stop="emit('remove-repo', repo.path)">
							<Trash2 class="h-3.5 w-3.5" />
						</button>
					</DropdownMenuItem>
				</template>
				<DropdownMenuSeparator />
				<DropdownMenuItem @select="emit('add-repo')">
					<Plus class="h-4 w-4" />
					Add repository…
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>

		<div class="flex-1" />

		<Popover>
			<PopoverTrigger as-child>
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
						whiteSpace: 'nowrap',
						cursor: 'pointer',
					}"
					data-no-drag
				>
					<Settings2 :size="14" />
					Tweaks
				</button>
			</PopoverTrigger>
			<PopoverContent align="end" :side-offset="8" class="w-[280px] p-0 border-0 shadow-none bg-transparent">
				<div
					:style="{
						width: '280px',
						background: 'var(--gd-panel)',
						border: '1px solid var(--gd-border)',
						borderRadius: '10px',
						boxShadow: 'var(--gd-shadow-lg)',
						color: 'var(--gd-text)',
						fontFamily: 'var(--font-sans)',
						fontSize: '13px',
						overflow: 'hidden',
					}"
				>
					<div
						:style="{
							padding: '10px 12px',
							borderBottom: '1px solid var(--gd-border)',
							fontSize: '13px',
							fontWeight: 600,
							color: 'var(--gd-text)',
						}"
					>
						Tweaks
					</div>
					<div :style="{ padding: '12px' }">
						<TweaksPanel :tweaks="tweaks" @update:tweak="(key, value) => emit('update:tweak', key, value)" />
					</div>
				</div>
			</PopoverContent>
		</Popover>

		<div
			:style="{
				width: '1px',
				height: '22px',
				background: 'var(--gd-border)',
				margin: '0 2px',
			}"
			data-no-drag
		/>

		<UserMenu :current-user="currentUser" :user-initials="userInitials" @log-out="emit('log-out')" />
	</div>
</template>
