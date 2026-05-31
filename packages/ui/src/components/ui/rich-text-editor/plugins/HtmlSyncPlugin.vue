<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue';
import { useLexicalComposer } from 'lexical-vue';
import { $generateHtmlFromNodes, $generateNodesFromDOM } from '@lexical/html';
import { $createParagraphNode, $getRoot, $isElementNode, type LexicalNode } from 'lexical';
import { mergeRegister } from '@lexical/utils';

type Props = {
	modelValue: string;
};

const props = defineProps<Props>();
const emit = defineEmits<{ 'update:modelValue': [string] }>();

const editor = useLexicalComposer();

let lastEmitted = '';
let applyingExternal = false;
let ready = false;

function appendNodesToRoot(nodes: LexicalNode[]): void {
	const root = $getRoot();

	let buffer: LexicalNode[] = [];

	const flush = (): void => {
		if (buffer.length === 0) {
			return;
		}

		const para = $createParagraphNode();

		para.append(...buffer);
		root.append(para);
		buffer = [];
	};

	for (const node of nodes) {
		if ($isElementNode(node)) {
			flush();
			root.append(node);
		} else {
			buffer.push(node);
		}
	}

	flush();
}

function applyHtml(html: string): void {
	applyingExternal = true;
	editor.update(() => {
		const root = $getRoot();

		root.clear();
		if (!html) {
			root.append($createParagraphNode());

			return;
		}

		const parser = new DOMParser();
		const dom = parser.parseFromString(html, 'text/html');
		const nodes = $generateNodesFromDOM(editor, dom);

		appendNodesToRoot(nodes);
		if (root.getChildrenSize() === 0) {
			root.append($createParagraphNode());
		}
	});
	queueMicrotask(() => {
		applyingExternal = false;
		ready = true;
		lastEmitted = html;
	});
}

applyHtml(props.modelValue ?? '');

watch(
	() => props.modelValue,
	(next) => {
		if (next === lastEmitted) {
			return;
		}

		applyHtml(next);
	},
);

const unregister = mergeRegister(
	editor.registerUpdateListener(({ editorState, dirtyElements, dirtyLeaves }) => {
		if (applyingExternal || !ready) {
			return;
		}

		if (dirtyElements.size === 0 && dirtyLeaves.size === 0) {
			return;
		}

		editorState.read(() => {
			const html = $generateHtmlFromNodes(editor, null);

			if (html === lastEmitted) {
				return;
			}

			lastEmitted = html;
			emit('update:modelValue', html);
		});
	}),
);

onBeforeUnmount(() => unregister());
</script>

<template>
	<span class="hidden" />
</template>
