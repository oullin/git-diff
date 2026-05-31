<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue';
import Dialog from '@ui/dialog/Dialog.vue';
import { formatShortcut } from '@/composables/formatShortcut';
import { matchesBinding, parseBinding } from '@/composables/keymapMatcher';
import { useCommandRegistry, type PaletteCommand } from '@/composables/useCommandRegistry';
import { useKeymap } from '@/composables/useKeymap';

import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@ui/command';

const open = ref(false);

const query = ref('');
const { list } = useCommandRegistry();
const { keymap, keyFor } = useKeymap();

const commandBarBinding = computed(() => keyFor('command_bar'));

const grouped = computed<Array<{ heading: string; items: PaletteCommand[] }>>(() => {
	const buckets = new Map<string, PaletteCommand[]>();

	for (const cmd of list.value) {
		const section = cmd.section ?? 'General';
		const items = buckets.get(section) ?? [];

		items.push(cmd);
		buckets.set(section, items);
	}

	return [...buckets.entries()].sort(([a], [b]) => a.localeCompare(b)).map(([heading, items]) => ({ heading, items }));
});

function openPalette(): void {
	query.value = '';
	open.value = true;
}

function close(): void {
	open.value = false;
}

function shortcutFor(cmd: PaletteCommand): string {
	if (!cmd.keymapId) {
		return '';
	}

	return formatShortcut(keymap.value[cmd.keymapId] ?? '');
}

function runCommand(cmd: PaletteCommand): void {
	close();
	void cmd.run();
}

function onKeydown(event: KeyboardEvent): void {
	// Closing is handled by Dialog's own Esc listener.
	if (matchesBindingString(event, commandBarBinding.value)) {
		event.preventDefault();
		openPalette();
	}
}

document.addEventListener('keydown', onKeydown);

onUnmounted(() => document.removeEventListener('keydown', onKeydown));

watch(open, (next) => {
	if (!next) {
		query.value = '';
	}
});

function matchesBindingString(event: KeyboardEvent, binding: string): boolean {
	const parsed = parseBinding(binding);

	if (!parsed) {
		return false;
	}

	return matchesBinding(event, parsed);
}
</script>

<template>
	<Dialog :show="open" max-width="lg" @close="close">
		<Command class="bg-popover" :model-value="null">
			<div class="flex items-center border-b border-border px-3">
				<CommandInput v-model="query" placeholder="Search commands..." class="h-11 py-2" autofocus />
				<span v-if="commandBarBinding" class="ml-3 rounded border border-border bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
					{{ formatShortcut(commandBarBinding) }}
				</span>
			</div>
			<CommandList class="max-h-80 overflow-y-auto p-1">
				<CommandEmpty class="px-3 py-6 text-center text-sm text-muted-foreground"> No commands match "{{ query }}". </CommandEmpty>
				<CommandGroup v-for="group in grouped" :key="group.heading" :heading="group.heading">
					<CommandItem v-for="cmd in group.items" :key="cmd.id" :value="cmd.id" @select="runCommand(cmd)">
						<span class="flex-1 truncate">{{ cmd.title }}</span>
						<span v-if="shortcutFor(cmd)" class="ml-3 text-xs text-muted-foreground">
							{{ shortcutFor(cmd) }}
						</span>
					</CommandItem>
				</CommandGroup>
			</CommandList>
		</Command>
	</Dialog>
</template>
