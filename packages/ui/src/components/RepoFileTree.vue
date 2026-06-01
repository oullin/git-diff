<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { FileTree, type FileTreeItemHandle, type GitStatusEntry } from '@pierre/trees';
import '@pierre/trees/web-components';

type Props = {
	paths: readonly string[];
	selectedPath?: string;
	gitStatus?: readonly GitStatusEntry[];
	initialExpansion?: 'open' | 'closed';
	expandAllOnReset?: boolean;
};

const props = withDefaults(defineProps<Props>(), {
	gitStatus: () => [],
	initialExpansion: 'open',
	expandAllOnReset: false,
});

function ancestorDirs(paths: readonly string[]): string[] {
	const set = new Set<string>();

	for (const path of paths) {
		let idx = path.indexOf('/');

		while (idx !== -1) {
			set.add(path.slice(0, idx));
			idx = path.indexOf('/', idx + 1);
		}
	}

	return Array.from(set);
}

const emit = defineEmits<{ select: [path: string] }>();

const mountEl = ref<HTMLDivElement | null>(null);

let tree: FileTree | null = null;

onMounted(() => {
	if (!mountEl.value) {
		return;
	}

	tree = new FileTree({
		paths: [...props.paths],
		gitStatus: [...props.gitStatus],
		flattenEmptyDirectories: true,
		initialExpansion: props.initialExpansion,
		search: false,
	});
	tree.subscribe(() => {
		const focused = tree?.getFocusedItem();

		if (focused && !focused.isDirectory()) {
			emit('select', focused.getPath());
		}
	});
	tree.render({ containerWrapper: mountEl.value });
});

watch(
	() => props.paths,
	(next) => {
		if (!tree) {
			return;
		}

		const options = props.expandAllOnReset ? { initialExpandedPaths: ancestorDirs(next) } : undefined;

		tree.resetPaths([...next], options);
		tree.setGitStatus([...props.gitStatus]);
	},
);

watch(
	() => props.gitStatus,
	(next) => {
		tree?.setGitStatus([...next]);
	},
);

watch(
	() => props.selectedPath,
	(path) => {
		if (!tree || !path) {
			return;
		}

		const item: FileTreeItemHandle | null = tree.getItem(path);

		if (item && !item.isSelected()) {
			item.select();
		}
	},
);

onBeforeUnmount(() => {
	tree?.cleanUp();
	tree = null;
});
</script>

<template>
	<div ref="mountEl" class="repo-file-tree h-full w-full" />
</template>

<style scoped>
.repo-file-tree {
	min-height: 120px;
	font-family: var(--font-code);
	font-size: 13px;
}
</style>
