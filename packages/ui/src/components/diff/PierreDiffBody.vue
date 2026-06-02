<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { FileDiff, type FileDiffMetadata, type FileDiffOptions, type SelectedLineRange } from '@pierre/diffs';
import { resolvedTheme } from '@composables/useTheme';
import { registerPierreThemes } from '@lib/pierreTheme';
import { pierreSurfaceClasses, toPierreDisplayOptions } from '@composables/usePierreDiffOptions';
import { resolveFileDiff, type DiffSourceContext } from '@composables/usePierreFileDiff';
import { commentsForSection, lineSideToSelectionSide, selectionToCommentTarget } from '@composables/usePierreComments';
import CommentThread from '@diff/CommentThread.vue';
import type { ChangedFile, DiffSection, DiffHunkStyle, DiffViewMode, ReviewComment } from '@git-diff/domain';
import type { PatchLine } from '@git-diff/domain/diff';
import type { LineSelectionRange } from '@composables/useLineSelection';
import type { RichTextFeatures } from '@ui/rich-text-editor';

// Renders a changed file's diff sections with the @pierre/diffs FileDiff
// engine — one instance per DiffSection (staged/unstaged/untracked/commit).
//
// Comments: FileDiff renders into a Shadow DOM, where the app's Tailwind-based
// CommentThread can't be styled. So the add-comment affordance uses the
// library's own in-diff gutter "+" (single click or drag-select), while the
// threads themselves render in light DOM beneath each section, labelled by
// line, with reply/resolve/delete intact. A comment is
// "outdated" when its anchored line is no longer present in the rendered diff
// (FileDiff.getLineIndex returns undefined), recomputed after every render.

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

const emit = defineEmits<{
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

const outdatedById = ref<Record<number, boolean>>({});

const surfaceClasses = computed(() => pierreSurfaceClasses({ diffStyle: props.diffStyle, density: props.density }));

function setContainerRef(id: string) {
	return (el: unknown) => {
		containers.value[id] = el instanceof HTMLElement ? el : null;
	};
}

function sectionComments(section: DiffSection): ReviewComment[] {
	return commentsForSection(props.comments, props.file.path, section.kind).filter((c) => !(props.hideResolved && c.resolved));
}

function recomputeOutdated(section: DiffSection, instance: FileDiff<undefined>): void {
	const next = { ...outdatedById.value };

	for (const comment of commentsForSection(props.comments, props.file.path, section.kind)) {
		const present = instance.getLineIndex(comment.lineNumber, lineSideToSelectionSide(comment.side)) != null;

		next[comment.id] = !present;
	}

	outdatedById.value = next;
}

function buildOptions(section: DiffSection): FileDiffOptions<undefined> {
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
		// Library renders its own "+" gutter affordance (single click or drag to
		// select a range); we only handle the resulting selection.
		enableGutterUtility: true,
		onGutterUtilityClick: (range: SelectedLineRange) => {
			const target = selectionToCommentTarget(range, section.id);

			emit('add-comment', section, target.line, target.range);
		},
		onPostRender: (_node, instance) => recomputeOutdated(section, instance as FileDiff<undefined>),
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

// Monotonic id stamped on each renderAll pass. Because resolveFileDiff is
// async, rapid prop changes can launch overlapping passes; a stale one that
// resolves late would otherwise render outdated content into the container.
let currentRenderId = 0;

/** Re-resolve + re-render every section (content/ref/whitespace changed). */
async function renderAll(): Promise<void> {
	const renderId = ++currentRenderId;

	const results = await Promise.all(
		props.file.sections.map(async (section) => ({
			section,
			...(await resolveFileDiff(props.file, section, sourceContext())),
		})),
	);

	// A newer renderAll started while we were resolving — discard this pass so
	// only the latest one mutates the DOM and the instances map.
	if (renderId !== currentRenderId) {
		return;
	}

	for (const { section, meta, partial } of results) {
		metas.set(section.id, meta);
		fallbacks.value = { ...fallbacks.value, [section.id]: partial && !meta };

		const el = containers.value[section.id];

		if (!el || !meta) {
			continue;
		}

		let fd = instances.get(section.id);

		if (!fd) {
			fd = new FileDiff<undefined>(buildOptions(section));
			instances.set(section.id, fd);
		}

		// Pass the wrapper as `containerWrapper` (not `fileContainer`): the library
		// then creates its own <diffs-container> custom element inside `el`, whose
		// constructor adopts the core stylesheet that resolves Shiki's per-token
		// CSS variables to actual colors. Handing it a plain element as
		// `fileContainer` skips that element, so tokens render uncoloured.
		fd.render({ fileDiff: meta, containerWrapper: el });
	}
}

/** Cheap option-only update (layout/indicators/word-diff/theme). */
function applyOptions(): void {
	for (const section of props.file.sections) {
		const fd = instances.get(section.id);

		if (fd) {
			fd.setOptions(buildOptions(section));
			fd.rerender();
		}
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

// Comments may load/update asynchronously after the last render; recompute each
// section's outdated map against its live FileDiff instance so anchors stay accurate.
watch(
	() => props.comments,
	() => {
		for (const [sectionId, fd] of instances) {
			const section = props.file.sections.find((s) => s.id === sectionId);

			if (section) {
				recomputeOutdated(section, fd);
			}
		}
	},
	{ deep: true },
);

onBeforeUnmount(disposeAll);
</script>

<template>
	<div :class="surfaceClasses" :style="{ background: 'var(--gd-bg-code, var(--gd-bg))' }">
		<section v-for="section in file.sections" :key="section.id">
			<div :ref="setContainerRef(section.id)"></div>
			<div v-if="section.binary || file.binary" :style="{ padding: '12px 16px', color: 'var(--gd-text-3)', fontSize: '13px' }">Binary file not shown</div>
			<div v-else-if="fallbacks[section.id]" :style="{ padding: '12px 16px', color: 'var(--gd-text-3)', fontSize: '13px' }">Diff unavailable</div>
			<CommentThread
				v-for="comment in sectionComments(section)"
				:key="comment.id"
				:comment="comment"
				:reply-features="replyFeatures"
				:outdated="outdatedById[comment.id] === true"
				@delete="emit('delete-comment', comment)"
				@reply="(bodyHtml) => emit('reply-comment', comment, bodyHtml)"
				@resolve="(resolved) => emit('resolve-comment', comment, resolved)"
			/>
		</section>
	</div>
</template>
