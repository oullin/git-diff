import { computed, ref, watch, type Ref } from 'vue';
import { useIntersectionObserver } from '@vueuse/core';
import { isDiffFileForced } from '@composables/useLazyRender';

// Files whose total churn exceeds this render eagerly anyway — placeholders for
// small files cause more flicker than they save.
export const LAZY_DIFF_LINE_THRESHOLD = 400;

const ESTIMATED_LINE_HEIGHT = 20;
const MIN_PLACEHOLDER_HEIGHT = 120;
const MAX_PLACEHOLDER_HEIGHT = 4000;

// Rough pixel height for a not-yet-rendered file so the scrollbar and anchor
// positions stay approximately stable until the real body mounts.
export function estimatedDiffHeight(lineCount: number, lineHeight = ESTIMATED_LINE_HEIGHT): number {
	const raw = lineCount * lineHeight + 48;

	return Math.min(Math.max(raw, MIN_PLACEHOLDER_HEIGHT), MAX_PLACEHOLDER_HEIGHT);
}

// "Mount once" viewport deferral. Returns `rendered`, which latches true the
// first time the target nears the viewport (or deferral is disabled, or the
// file is force-rendered) and never reverts.
export function useLazyDiffFile(target: Ref<HTMLElement | null>, options: { path: Ref<string>; shouldDefer: Ref<boolean> }) {
	const intersected = ref(false);

	const { stop } = useIntersectionObserver(
		target,
		(entries) => {
			if (entries.some((entry) => entry.isIntersecting)) {
				intersected.value = true;
				stop();
			}
		},
		{ rootMargin: '600px' },
	);

	watch(
		() => options.shouldDefer.value,
		(defer) => {
			if (!defer) {
				stop();
			}
		},
	);

	const rendered = computed(() => !options.shouldDefer.value || intersected.value || isDiffFileForced(options.path.value));

	// immediate: a file that is already rendered on init (e.g. small files that
	// never defer) must stop its observer right away, otherwise it leaks.
	watch(
		rendered,
		(value) => {
			if (value) {
				stop();
			}
		},
		{ immediate: true },
	);

	return { rendered };
}
