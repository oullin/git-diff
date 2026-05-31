import type { Ref } from 'vue';
import { PREF_KEYS, type ChangedFile } from '@git-diff/domain';
import { useCommandRegistry } from '@composables/useCommandRegistry';

/**
 * Registers the application's command-palette entries. Pure wiring: each
 * command delegates to an injected app action, so App.vue no longer owns the
 * registration boilerplate.
 */
export interface AppCommandsOptions {
	hideWhitespace: () => boolean;
	selectedFile: Ref<ChangedFile | null>;
	savePreferences: (patch: Record<string, string>) => Promise<void>;
	copyReviewAsMarkdown: () => void | Promise<void>;
	generateWalkthrough: () => void | Promise<void>;
	copyPath: (path: string) => void;
}

export function useAppCommands(opts: AppCommandsOptions): void {
	const palette = useCommandRegistry();

	palette.register({
		id: 'diff.toggle-whitespace',
		title: 'Toggle whitespace-only changes',
		section: 'Diff',
		keymapId: 'toggle_whitespace',
		run: () =>
			opts.savePreferences({
				[PREF_KEYS.diffHideWhitespace]: opts.hideWhitespace() ? '0' : '1',
			}),
	});

	palette.register({
		id: 'review.copy-markdown',
		title: 'Copy active review as Markdown',
		section: 'Review',
		run: () => opts.copyReviewAsMarkdown(),
	});

	palette.register({
		id: 'walkthrough.generate',
		title: 'Generate AI walkthrough',
		section: 'Review',
		run: async () => {
			await opts.generateWalkthrough();
		},
	});

	palette.register({
		id: 'file.copy-path',
		title: 'Copy current file path',
		section: 'File',
		run: () => {
			const file = opts.selectedFile.value;

			if (file) {
				opts.copyPath(file.path);
			}
		},
	});
}
