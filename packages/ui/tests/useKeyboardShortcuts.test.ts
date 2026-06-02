import { mount } from '@vue/test-utils';
import { createPinia } from 'pinia';
import { ref, defineComponent, h, type Ref } from 'vue';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useKeyboardShortcuts, type UseKeyboardShortcutsOptions } from '@composables/useKeyboardShortcuts.js';

// Mounts a trivial host that wires the shortcuts in setup, mirroring how
// useAppSession attaches them synchronously during component setup.
function mountHost(overrides: Partial<UseKeyboardShortcutsOptions> = {}, enabled: Ref<boolean> = ref(true)) {
	const opts: UseKeyboardShortcutsOptions = {
		enabled,
		onSelectAdjacent: vi.fn(),
		onJumpToHunk: vi.fn(),
		onToggleViewed: vi.fn(),
		onStartReview: vi.fn(),
		onOpenSearch: vi.fn(),
		...overrides,
	};

	const Host = defineComponent({
		setup() {
			useKeyboardShortcuts(opts);

			return () => h('div');
		},
	});

	const wrapper = mount(Host, { global: { plugins: [createPinia()] } });

	return { wrapper, opts };
}

function press(key: string, init: KeyboardEventInit = {}): void {
	document.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true, ...init }));
}

describe('useKeyboardShortcuts', () => {
	beforeEach(() => {
		document.body.innerHTML = '';
	});

	it('fires the navigation action for the default next_file binding', () => {
		const { opts, wrapper } = mountHost();

		press('j');

		expect(opts.onSelectAdjacent).toHaveBeenCalledWith(1);

		wrapper.unmount();
	});

	it('fires for ArrowDown / ArrowUp aliases', () => {
		const { opts, wrapper } = mountHost();

		press('ArrowDown');
		press('ArrowUp');

		expect(opts.onSelectAdjacent).toHaveBeenNthCalledWith(1, 1);
		expect(opts.onSelectAdjacent).toHaveBeenNthCalledWith(2, -1);

		wrapper.unmount();
	});

	it('does nothing when disabled', () => {
		const enabled = ref(false);
		const { opts, wrapper } = mountHost({}, enabled);

		press('j');

		expect(opts.onSelectAdjacent).not.toHaveBeenCalled();

		wrapper.unmount();
	});

	it('ignores keystrokes typed into inputs', () => {
		const { opts, wrapper } = mountHost();
		const input = document.createElement('input');

		document.body.appendChild(input);

		input.dispatchEvent(new KeyboardEvent('keydown', { key: 'j', bubbles: true }));

		expect(opts.onSelectAdjacent).not.toHaveBeenCalled();

		wrapper.unmount();
	});

	it('detaches the listener on unmount', () => {
		const { opts, wrapper } = mountHost();

		wrapper.unmount();
		press('j');

		expect(opts.onSelectAdjacent).not.toHaveBeenCalled();
	});
});
