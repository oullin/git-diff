<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { FileDiff, type FileDiffMetadata, type FileDiffOptions } from '@pierre/diffs';
import { resolvedTheme } from '@composables/useTheme';
import { registerPierreThemes } from '@lib/pierreTheme';
import { pierreSurfaceClasses, toPierreDisplayOptions } from '@composables/usePierreDiffOptions';
import { resolveFileDiff, type DiffSourceContext } from '@composables/usePierreFileDiff';
import type { ChangedFile, DiffHunkStyle, DiffSection, DiffViewMode, ReviewComment } from '@git-diff/domain';
import type { PatchLine } from '@git-diff/domain/diff';
import type { LineSelectionRange } from '@composables/useLineSelection';
import type { RichTextFeatures } from '@ui/rich-text-editor';

// Renders a changed file's diff sections with the @pierre/diffs FileDiff
// engine — one instance per DiffSection (staged/unstaged/untracked/commit).
// Phase 3: rendering + theming + native context expansion only. Comments,
// the add-comment affordance, and line selection are layered on in Phase 4
// (the comment-related props/emits are accepted here to preserve the
// DiffBody.vue contract so LazyDiffBody.vue is a drop-in swap).

const props = withDefaults(
	defineProps<{
		file: ChangedFile;
		viewMode: DiffViewMode;
		diffStyle: DiffHunkStyle;
		density: 'comfortable' | 'compact';
		wordHighlight: boolean;
		wrapLongLines: boolean;
		hideWhitespace: boolean;
		hideResolved?: boolean;
		comments: ReviewComment[];
		replyFeatures: RichTextFeatures;
		repoRoot: string;
		commitRef?: string;
		baseRef?: string;
	}>(),
	{ commitRef: undefined, baseRef: undefined, hideResolved: false },
);

// Emits kept for the DiffBody contract; wired up in Phase 4.
defineEmits<{
	'add-comment': [section: DiffSection, line: PatchLine, range?: LineSelectionRange];
	'delete-comment': [comment: ReviewComment];
	'reply-comment': [parent: ReviewComment, bodyHtml: string];
	'resolve-comment': [comment: ReviewComment, resolved: boolean];
}>();

// Themes must be registered before any FileDiff resolves its theme. main.ts
// also calls this, but registering here (idempotent) covers entry points that
// bypass main.ts — e.g. component unit tests that mount the app directly.
registerPierreThemes();

const containers = ref<Record<string, HTMLElement | null>>({});
const instances = new Map<string, FileDiff<undefined>>();
const metas = new Map<string, FileDiffMetadata | undefined>();

const fallbacks = ref<Record<string, boolean>>({});

const surfaceClasses = computed(() => pierreSurfaceClasses({ diffStyle: props.diffStyle, density: props.density }));

function setContainerRef(id: string) {
	return (el: unknown) => {
		containers.value[id] = el instanceof HTMLElement ? el : null;
	};
}

function buildOptions(): FileDiffOptions<undefined> {
	return {
		...toPierreDisplayOptions({
			viewMode: props.viewMode,
			diffStyle: props.diffStyle,
			wordHighlight: props.wordHighlight,
			wrapLongLines: props.wrapLongLines,
			hideWhitespace: props.hideWhitespace,
			themeType: resolvedTheme.value,
		}),
		disableFileHeader: true,
		// Pure-JS Shiki avoids wasm/worker bundling friction under Electron's
		// file:// origin; revisit for perf in a later pass.
		preferredHighlighter: 'shiki-js',
		useCSSClasses: true,
	};
}

function sourceContext(): DiffSourceContext {
	return {
		repoRoot: props.repoRoot,
		commitRef: props.commitRef,
		baseRef: props.baseRef,
		ignoreWhitespace: props.hideWhitespace,
	};
}

async function renderSection(section: ChangedFile['sections'][number]): Promise<void> {
	const { meta, partial } = await resolveFileDiff(props.file, section, sourceContext());

	metas.set(section.id, meta);
	fallbacks.value = { ...fallbacks.value, [section.id]: partial && !meta };

	const el = containers.value[section.id];

	if (!el || !meta) {
		return;
	}

	let fd = instances.get(section.id);

	if (!fd) {
		fd = new FileDiff<undefined>(buildOptions());
		instances.set(section.id, fd);
	}

	fd.render({ fileDiff: meta, fileContainer: el });
}

/** Re-resolve + re-render every section (content/ref/whitespace changed). */
async function renderAll(): Promise<void> {
	await Promise.all(props.file.sections.map((section) => renderSection(section)));
}

/** Cheap option-only update (layout/indicators/word-diff/theme). */
function applyOptions(): void {
	const options = buildOptions();

	for (const fd of instances.values()) {
		fd.setOptions(options);
		fd.rerender();
	}
}

function disposeAll(): void {
	for (const fd of instances.values()) {
		fd.cleanUp();
	}

	instances.clear();
	metas.clear();
}

onMounted(renderAll);

// Content / source identity → full re-resolve (the diff itself changes).
watch(
	() => [props.file.path, props.file.fingerprint, props.commitRef, props.baseRef, props.hideWhitespace, props.file.sections.map((s) => s.id).join(',')].join('|'),
	() => {
		disposeAll();
		void renderAll();
	},
);

// Display-only options → in-place update, no re-fetch.
watch(() => [props.viewMode, props.diffStyle, props.wordHighlight, props.wrapLongLines, resolvedTheme.value].join('|'), applyOptions);

onBeforeUnmount(disposeAll);
</script>

<template>
	<div :class="surfaceClasses" :style="{ background: 'var(--gd-bg-code, var(--gd-bg))' }">
		<section v-for="section in file.sections" :key="section.id">
			<div :ref="setContainerRef(section.id)"></div>
			<div v-if="section.binary || file.binary" :style="{ padding: '12px 16px', color: 'var(--gd-text-3)', fontSize: '13px' }">Binary file not shown</div>
			<div v-else-if="fallbacks[section.id]" :style="{ padding: '12px 16px', color: 'var(--gd-text-3)', fontSize: '13px' }">Diff unavailable</div>
		</section>
	</div>
</template>
