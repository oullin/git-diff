import { computed, type Ref } from 'vue';

import type { AuthUser, ChangedFile, RepositoryState, ReviewComment, ReviewDetail } from '@git-diff/domain';

/**
 * Pure derived view-state over the repo/review/auth stores: changed-file maps,
 * the selected file, comment thread counts, user initials, etc. No side
 * effects — every value is a computed of the injected refs, so this is
 * unit-testable without mounting a component.
 */
export interface RepoSelectorsOptions {
	state: Ref<RepositoryState | null>;
	selectedPath: Ref<string>;
	activeReview: Ref<ReviewDetail | null>;
	currentUser: Ref<AuthUser | null>;
}

export function useRepoSelectors(opts: RepoSelectorsOptions) {
	const files = computed<ChangedFile[]>(() => opts.state.value?.files ?? []);

	const changedByPath = computed(() => {
		const map = new Map<string, ChangedFile>();

		for (const file of files.value) {
			map.set(file.path, file);
		}

		return map;
	});

	const changedPathsSet = computed(() => new Set(changedByPath.value.keys()));

	const trackedFiles = computed(() => opts.state.value?.trackedFiles ?? []);

	const repoPaths = computed(() => {
		const set = new Set<string>(trackedFiles.value);

		for (const file of files.value) {
			set.add(file.path);
		}

		return Array.from(set).sort();
	});

	const selectedFile = computed<ChangedFile | null>(() => changedByPath.value.get(opts.selectedPath.value) ?? files.value[0] ?? null);

	const selectedIsChanged = computed(() => !!opts.selectedPath.value && changedByPath.value.has(opts.selectedPath.value));

	const reviewComments = computed<ReviewComment[]>(() => opts.activeReview.value?.comments ?? []);

	const threadsByPath = computed(() => {
		const counts = new Map<string, number>();

		for (const comment of reviewComments.value) {
			counts.set(comment.filePath, (counts.get(comment.filePath) ?? 0) + 1);
		}

		return counts;
	});

	const userInitials = computed(() => {
		const name = opts.currentUser.value?.displayName ?? opts.currentUser.value?.osUsername ?? 'GO';

		return (
			name
				.split(/[\s_-]+/)
				.map((part) => part[0]?.toUpperCase() ?? '')
				.slice(0, 2)
				.join('') || 'GO'
		);
	});

	const changedIndex = computed(() => files.value.findIndex((f) => f.path === opts.selectedPath.value));

	function threadsForFile(path: string): number {
		return threadsByPath.value.get(path) ?? 0;
	}

	return {
		files,
		changedByPath,
		changedPathsSet,
		trackedFiles,
		repoPaths,
		selectedFile,
		selectedIsChanged,
		reviewComments,
		threadsByPath,
		userInitials,
		changedIndex,
		threadsForFile,
	};
}
