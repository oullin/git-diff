<script setup lang="ts">
import { ref, watch } from 'vue';
import { Teleport } from 'vue';
import type { TextNode } from 'lexical';
import { $createMentionNode } from '@rich-text-editor/nodes/MentionNode';
import type { MentionItem } from '@rich-text-editor/plugins/mentions';

import { useLexicalComposer, TypeaheadMenuPlugin, MenuOption, useBasicTypeaheadTriggerMatch } from 'lexical-vue';

type Props = {
	lookup?: (query: string) => Promise<MentionItem[]>;
};

const props = defineProps<Props>();

class MentionOption extends MenuOption {
	item: MentionItem;

	constructor(item: MentionItem) {
		super(item.id);
		this.item = item;
	}
}

const editor = useLexicalComposer();

const options = ref<MentionOption[]>([]);

const query = ref<string | null>(null);

const triggerFn = useBasicTypeaheadTriggerMatch('@', { minLength: 0, maxLength: 50 });

watch(query, async (next, _previous, onCleanup) => {
	let isCurrent = true;

	onCleanup(() => {
		isCurrent = false;
	});

	if (next === null || !props.lookup) {
		options.value = [];

		return;
	}

	try {
		const items = await props.lookup(next);

		if (!isCurrent) {
			return;
		}

		options.value = items.map((item) => new MentionOption(item));
	} catch {
		if (!isCurrent) {
			return;
		}

		options.value = [];
	}
});

function onSelectOption({ option, closeMenu, textNodeContainingQuery }: { option: MentionOption; closeMenu: () => void; textNodeContainingQuery: TextNode | null; matchingString: string }): void {
	editor.update(() => {
		const mention = $createMentionNode(option.item.id, option.item.label);

		if (textNodeContainingQuery) {
			textNodeContainingQuery.replace(mention);
			mention.select();
		}

		closeMenu();
	});
}
</script>

<template>
	<TypeaheadMenuPlugin :options="options" :trigger-fn="triggerFn" @query-change="(value) => (query = value)" @select-option="onSelectOption">
		<template #default="{ anchorElementRef, itemProps, matchingString }">
			<Teleport v-if="anchorElementRef && options.length > 0" :to="anchorElementRef">
				<div class="z-50 max-h-72 w-56 overflow-y-auto rounded-md border border-border bg-popover p-1 text-sm shadow-md">
					<button
						v-for="(option, index) in itemProps.options"
						:key="option.key"
						type="button"
						:ref="(el) => option.setRefElement(el as Element | null)"
						class="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left hover:bg-accent"
						:class="itemProps.selectedIndex === index ? 'bg-accent text-accent-foreground' : ''"
						@mousedown.prevent
						@click="itemProps.selectOptionAndCleanUp(option as MentionOption)"
						@mouseenter="itemProps.setHighlightedIndex(index)"
					>
						<span>@{{ (option as MentionOption).item.label }}</span>
					</button>
					<div v-if="itemProps.options.length === 0 && matchingString" class="px-2 py-1.5 text-xs text-muted-foreground">No matches for "{{ matchingString }}"</div>
				</div>
			</Teleport>
		</template>
	</TypeaheadMenuPlugin>
</template>
