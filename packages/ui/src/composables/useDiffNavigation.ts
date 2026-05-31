import type { ChangedFile } from '@git-diff/domain';
import type { ComputedRef } from 'vue';
import { forceRenderAllDiffFiles } from '@composables/useLazyRender';

export interface UseDiffNavigationOptions {
	files: ComputedRef<ChangedFile[]>;
	changedIndex: ComputedRef<number>;
	onSelect: (path: string) => void;
}

// Keyboard binding lives in useKeyboardShortcuts; this owns the actions
// only so callers can wire them up outside keyboard contexts (buttons).
export function useDiffNavigation({ files, changedIndex, onSelect }: UseDiffNavigationOptions) {
	function selectAdjacent(delta: number): void {
		if (files.value.length === 0) {
			return;
		}

		const next = Math.min(Math.max(changedIndex.value + delta, 0), files.value.length - 1);
		const file = files.value[next];

		if (file) {
			onSelect(file.path);
		}
	}

	function jumpToHunk(delta: number): void {
		const anchors = Array.from(document.querySelectorAll<HTMLElement>('[data-hunk-anchor]'));

		if (anchors.length === 0) {
			// Every diff is still viewport-deferred — reveal them so hunks exist.
			forceRenderAllDiffFiles();

			return;
		}

		const midpoint = window.innerHeight / 2;

		let current = 0;

		for (let i = 0; i < anchors.length; i++) {
			const rect = anchors[i]!.getBoundingClientRect();

			if (rect.top <= midpoint) {
				current = i;
			} else {
				break;
			}
		}

		const rawTarget = current + delta;

		if (rawTarget < 0 || rawTarget > anchors.length - 1) {
			// Stepping past the rendered hunks: reveal deferred files so a
			// follow-up press can land on their hunks instead of stalling.
			forceRenderAllDiffFiles();
		}

		const target = Math.max(0, Math.min(anchors.length - 1, rawTarget));
		const element = anchors[target]!;

		element.scrollIntoView({ block: 'center', behavior: 'smooth' });
		element.classList.add('gd-hunk-flash');
		window.setTimeout(() => element.classList.remove('gd-hunk-flash'), 350);
	}

	return { selectAdjacent, jumpToHunk };
}
