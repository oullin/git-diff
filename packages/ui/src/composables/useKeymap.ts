import { computed, type ComputedRef } from 'vue';
import type { Keymap, KeymapAction } from '@git-diff/domain';
import { useUserConfig } from '@/composables/useUserConfig';

// Mirrors Go's userconfig.Defaults() — keep in sync. The UI needs *some*
// keymap before the first fetch lands.
const DEFAULT_KEYMAP: Keymap = {
	command_bar: 'cmd+shift+p',
	file_filter: 'cmd+f',
	diff_search: '/',
	submit_comment: 'cmd+enter',
	discard_comment: 'escape',
	toggle_sidebar: 'cmd+\\',
	next_file: 'j',
	prev_file: 'k',
	next_hunk: 'n',
	prev_hunk: 'p',
	toggle_viewed: 'v',
	toggle_whitespace: 'w',
};

export interface UseKeymap {
	keymap: ComputedRef<Keymap>;
	/** Lookup helper that always returns a string (default when missing). */
	keyFor(action: KeymapAction): string;
}

export function useKeymap(): UseKeymap {
	const { config } = useUserConfig();

	const keymap = computed<Keymap>(() => ({
		...DEFAULT_KEYMAP,
		...config.value?.keymap,
	}));

	return {
		keymap,
		keyFor(action) {
			return keymap.value[action] ?? DEFAULT_KEYMAP[action];
		},
	};
}
