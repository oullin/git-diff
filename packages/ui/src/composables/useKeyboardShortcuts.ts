import { onMounted, onUnmounted, type Ref } from 'vue';
import { matchesBinding, parseBinding, type ParsedBinding } from '@/composables/keymapMatcher';
import { useKeymap } from '@/composables/useKeymap';

export interface UseKeyboardShortcutsOptions {
	enabled: Ref<boolean>;
	onSelectAdjacent: (delta: number) => void;
	onJumpToHunk: (delta: number) => void;
	onToggleViewed: () => void;
	onStartReview: () => void;
	onOpenSearch: () => void;
}

/** Returns a manual unsubscribe so callers that need to detach before
 *  onUnmounted (e.g. dialogs) can do so. */
export function useKeyboardShortcuts(opts: UseKeyboardShortcutsOptions): () => void {
	const { keymap } = useKeymap();

	function onKeydown(event: KeyboardEvent): void {
		if (!opts.enabled.value) {
			return;
		}

		const target = event.target as HTMLElement | null;

		if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)) {
			return;
		}

		// Snapshot per event so hot-reloaded bindings take effect immediately.
		const bindings = currentBindings(keymap.value);

		// Arrow keys are hard-coded aliases for next_file/prev_file.
		if (event.key === 'ArrowDown' && !event.metaKey && !event.ctrlKey && !event.altKey) {
			event.preventDefault();
			opts.onSelectAdjacent(1);

			return;
		}

		if (event.key === 'ArrowUp' && !event.metaKey && !event.ctrlKey && !event.altKey) {
			event.preventDefault();
			opts.onSelectAdjacent(-1);

			return;
		}

		if (bindings.nextFile && matchesBinding(event, bindings.nextFile)) {
			event.preventDefault();
			opts.onSelectAdjacent(1);

			return;
		}

		if (bindings.prevFile && matchesBinding(event, bindings.prevFile)) {
			event.preventDefault();
			opts.onSelectAdjacent(-1);

			return;
		}

		if (bindings.nextHunk && matchesBinding(event, bindings.nextHunk)) {
			event.preventDefault();
			opts.onJumpToHunk(1);

			return;
		}

		if (bindings.prevHunk && matchesBinding(event, bindings.prevHunk)) {
			event.preventDefault();
			opts.onJumpToHunk(-1);

			return;
		}

		if (bindings.toggleViewed && matchesBinding(event, bindings.toggleViewed)) {
			opts.onToggleViewed();

			return;
		}

		if (bindings.submitComment && matchesBinding(event, bindings.submitComment)) {
			event.preventDefault();
			opts.onStartReview();

			return;
		}

		if ((bindings.fileFilter && matchesBinding(event, bindings.fileFilter)) || (bindings.diffSearch && matchesBinding(event, bindings.diffSearch))) {
			event.preventDefault();
			opts.onOpenSearch();
		}
	}

	let bound = false;

	function bind(): void {
		if (bound) {
			return;
		}

		document.addEventListener('keydown', onKeydown);
		bound = true;
	}

	function unbind(): void {
		if (!bound) {
			return;
		}

		document.removeEventListener('keydown', onKeydown);
		bound = false;
	}

	onMounted(bind);

	onUnmounted(unbind);

	// Bind synchronously too so callers that don't rely on onMounted
	// (e.g. lazy-mounted dialogs) still get the listener.
	bind();

	return unbind;
}

interface CurrentBindings {
	nextFile: ParsedBinding | null;
	prevFile: ParsedBinding | null;
	nextHunk: ParsedBinding | null;
	prevHunk: ParsedBinding | null;
	toggleViewed: ParsedBinding | null;
	submitComment: ParsedBinding | null;
	fileFilter: ParsedBinding | null;
	diffSearch: ParsedBinding | null;
}

function currentBindings(keymap: ReturnType<typeof useKeymap>['keymap']['value']): CurrentBindings {
	return {
		nextFile: parseBinding(keymap.next_file),
		prevFile: parseBinding(keymap.prev_file),
		nextHunk: parseBinding(keymap.next_hunk),
		prevHunk: parseBinding(keymap.prev_hunk),
		toggleViewed: parseBinding(keymap.toggle_viewed),
		submitComment: parseBinding(keymap.submit_comment),
		fileFilter: parseBinding(keymap.file_filter),
		diffSearch: parseBinding(keymap.diff_search),
	};
}
