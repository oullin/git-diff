<script setup lang="ts">
import { onBeforeUnmount } from 'vue';
import { Loader2, Search } from 'lucide-vue-next';
import { Command, CommandEmpty, CommandInput, CommandItem, CommandList } from '@ui/command';
import { Popover, PopoverAnchor, PopoverContent } from '@ui/popover';
import Kbd from '@components/diff/Kbd.vue';
import { useFileSearch } from '@composables/useFileSearch';
import type { FileSearchResult } from '@git-diff/domain';

const emit = defineEmits<{
	'select-result': [result: FileSearchResult];
}>();

const { query: searchQuery, results: searchResults, loading: searchLoading, error: searchError, open: searchOpen, onInput: onSearchInput, reset: resetSearch } = useFileSearch();

function basename(path: string): string {
	const slash = path.lastIndexOf('/');

	return slash >= 0 ? path.slice(slash + 1) : path;
}

function dirname(path: string): string {
	const slash = path.lastIndexOf('/');

	return slash >= 0 ? path.slice(0, slash) : '';
}

function onSelectResult(result: FileSearchResult): void {
	emit('select-result', result);
	resetSearch();
}

onBeforeUnmount(() => {
	resetSearch();
});
</script>

<template>
	<Popover :open="searchOpen">
		<PopoverAnchor as-child>
			<Command
				:model-value="undefined"
				class="flex flex-row items-center"
				:style="{
					flex: 1,
					minWidth: 0,
					gap: '6px',
					height: '32px',
					padding: '0 10px',
					borderRadius: '8px',
					background: 'var(--gd-panel-2)',
					border: '1px solid var(--gd-border)',
				}"
				@update:model-value="(value: any) => value && onSelectResult(value as FileSearchResult)"
			>
				<Search :size="13" :style="{ color: 'var(--gd-text-3)', flexShrink: 0 }" />
				<CommandInput
					:model-value="searchQuery"
					placeholder="Search files across all repos…"
					:style="{
						flex: 1,
						minWidth: 0,
						background: 'transparent',
						color: 'var(--gd-text)',
						fontSize: '13.5px',
					}"
					@update:model-value="onSearchInput"
				/>
				<Loader2 v-if="searchLoading" :size="13" class="animate-spin" :style="{ color: 'var(--gd-text-3)', flexShrink: 0 }" />
				<Kbd>⌘P</Kbd>

				<PopoverContent align="start" :side-offset="6" class="p-0 w-[var(--reka-popover-trigger-width)] max-w-[640px]" @open-auto-focus="(event: Event) => event.preventDefault()">
					<CommandList class="max-h-[360px]">
						<CommandEmpty v-if="searchError">{{ searchError }}</CommandEmpty>
						<CommandEmpty v-else-if="!searchLoading && searchResults.length === 0"> No matching files </CommandEmpty>
						<CommandItem v-for="result in searchResults" :key="`${result.repoPath}::${result.filePath}`" :value="result" class="flex items-center gap-2">
							<span class="font-mono text-sm truncate">{{ basename(result.filePath) }}</span>
							<span v-if="dirname(result.filePath)" class="font-mono text-xs text-muted-foreground truncate">{{ dirname(result.filePath) }}</span>
							<span class="ml-auto pl-2 text-xs text-muted-foreground shrink-0">
								{{ result.repoName }}
							</span>
						</CommandItem>
					</CommandList>
				</PopoverContent>
			</Command>
		</PopoverAnchor>
	</Popover>
</template>
