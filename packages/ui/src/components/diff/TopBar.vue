<script setup lang="ts">
import { RefreshCw, Sparkles } from 'lucide-vue-next';
import BranchPicker from '@components/diff/BranchPicker.vue';
import DiffStat from '@components/diff/DiffStat.vue';
import FileSearchPopover from '@components/diff/FileSearchPopover.vue';
import Kbd from '@components/diff/Kbd.vue';
import CommitPicker from '@entry/components/commits/CommitPicker.vue';
import PullRequestPicker from '@entry/components/commits/PullRequestPicker.vue';
import type { CommitSummary, FileSearchResult, PullRequestSummary, RepositoryState } from '@git-diff/domain';

defineProps<{
	state: RepositoryState | null;
	creatingBranch: boolean;
	branchCreateError: string;
	commits: CommitSummary[];
	commitsLoading: boolean;
	repoMode: 'working' | 'commit';
	pullRequests: PullRequestSummary[];
	pullRequestsLoading: boolean;
	activePullRequest: PullRequestSummary | null;
	walkthroughLoading: boolean;
	hasWalkthrough: boolean;
	hasActiveReview: boolean;
	copyReviewState: 'idle' | 'copied' | 'error';
}>();

const emit = defineEmits<{
	refresh: [];
	'switch-branch': [branch: string];
	'create-branch': [name: string];
	'select-result': [result: FileSearchResult];
	'load-commits': [];
	'open-commit': [sha: string];
	'back-to-working': [];
	'load-pull-requests': [];
	'open-pull-request': [number: number];
	'generate-walkthrough': [];
	'copy-review-as-markdown': [];
}>();

const linkButton = {
	display: 'inline-flex',
	alignItems: 'center',
	gap: '6px',
	height: '30px',
	padding: '0 10px',
	borderRadius: '6px',
	border: '1px solid var(--gd-border)',
	background: 'var(--gd-bg)',
	color: 'var(--gd-text-2)',
	fontSize: '13px',
	fontWeight: 500,
	cursor: 'pointer',
} as const;
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

		<template v-if="state">
			<CommitPicker
				:commits="commits"
				:loading="commitsLoading"
				:active-sha="state.commitSha"
				:mode="repoMode"
				@open="emit('load-commits')"
				@select="(sha: string) => emit('open-commit', sha)"
				@back="emit('back-to-working')"
			/>
			<PullRequestPicker
				:pull-requests="pullRequests"
				:loading="pullRequestsLoading"
				:active-number="activePullRequest?.number"
				@open="emit('load-pull-requests')"
				@select="(number: number) => emit('open-pull-request', number)"
			/>
			<button type="button" :style="linkButton" :disabled="walkthroughLoading" @click="emit('generate-walkthrough')">
				<Sparkles :size="14" />
				{{ walkthroughLoading ? 'Generating…' : hasWalkthrough ? 'Refresh walkthrough' : 'Generate walkthrough' }}
			</button>
			<button v-if="hasActiveReview" type="button" :style="linkButton" :disabled="copyReviewState === 'copied'" @click="emit('copy-review-as-markdown')">
				{{ copyReviewState === 'copied' ? 'Copied!' : 'Copy review as Markdown' }}
			</button>
		</template>

		<button type="button" :style="linkButton" :disabled="!state" @click="emit('refresh')">
			<RefreshCw :size="14" />
			Refresh
			<Kbd>R</Kbd>
		</button>
	</div>
</template>
