import { afterEach, describe, expect, test, vi } from 'vitest';
import { useCommandRegistry } from '@composables/useCommandRegistry.js';

// The registry is a module singleton — clear it between tests so leftover
// commands from one case don't contaminate the next.
afterEach(() => {
	const { commands } = useCommandRegistry();

	commands.value = [];
});

describe('useCommandRegistry', () => {
	test('register adds a command and returns a disposer', () => {
		const { register, list } = useCommandRegistry();

		const dispose = register({
			id: 'test.cmd',
			title: 'Test',
			run: () => {},
		});

		expect(list.value.find((c) => c.id === 'test.cmd')?.title).toBe('Test');

		dispose();

		expect(list.value.find((c) => c.id === 'test.cmd')).toBeUndefined();
	});

	test('re-registering the same id replaces the existing entry', () => {
		const { register, list } = useCommandRegistry();

		register({ id: 'cmd.x', title: 'v1', run: () => {} });
		register({ id: 'cmd.x', title: 'v2', run: () => {} });

		const matches = list.value.filter((c) => c.id === 'cmd.x');

		expect(matches).toHaveLength(1);
		expect(matches[0]!.title).toBe('v2');
	});

	test('deregister is a no-op for unknown ids', () => {
		const { deregister, list } = useCommandRegistry();

		const before = list.value.length;

		deregister('never.registered');

		expect(list.value.length).toBe(before);
	});

	test('running a command invokes the registered handler', async () => {
		const { register, list } = useCommandRegistry();
		const handler = vi.fn();

		register({ id: 'cmd.run', title: 'Run', run: handler });

		await list.value.find((c) => c.id === 'cmd.run')!.run();

		expect(handler).toHaveBeenCalledOnce();
	});
});
