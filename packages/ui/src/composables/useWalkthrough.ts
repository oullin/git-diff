import { ref, type Ref } from 'vue';
import type { RepositoryState, WalkthroughRecord } from '@git-diff/domain';

// useWalkthrough owns the LLM walkthrough panel state: the currently rendered
// record (or null when nothing has been generated for the active repo), the
// in-flight flag, and any error message from the last attempt. The single
// `generate` action snapshots the supplied repository state, so callers can
// invoke it without re-deriving the request payload.
export interface UseWalkthrough {
	record: Ref<WalkthroughRecord | null>;
	loading: Ref<boolean>;
	error: Ref<string>;
	generate: (refresh?: boolean) => Promise<void>;
	dismiss: () => void;
}

export function useWalkthrough(state: Ref<RepositoryState | null>): UseWalkthrough {
	const record = ref<WalkthroughRecord | null>(null);

	const loading = ref(false);

	const error = ref('');

	async function generate(refresh = false): Promise<void> {
		const current = state.value;

		if (!current) {
			return;
		}

		loading.value = true;
		error.value = '';

		try {
			record.value = await window.diffApp.generateWalkthrough({
				path: current.root,
				kind: current.mode,
				sha: current.commitSha,
				refresh,
			});
		} catch (cause) {
			error.value = cause instanceof Error ? cause.message : String(cause);
			record.value = null;
		} finally {
			loading.value = false;
		}
	}

	function dismiss(): void {
		record.value = null;
		error.value = '';
	}

	return { record, loading, error, generate, dismiss };
}
