import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { ReviewDetail, ReviewSession } from '@git-diff/domain';

// useReviewsStore holds the review list scoped to the active repository,
// the detail record for the review currently displayed (or null), and the
// rich-text summary the user is composing in the ReviewPanel. Adding /
// updating / starting reviews still lives in App.vue because those flows
// also promote pending comments and mutate repo state.
export const useReviewsStore = defineStore('reviews', () => {
	const items = ref<ReviewSession[]>([]);

	const active = ref<ReviewDetail | null>(null);

	const summaryDraft = ref('');

	function setItems(reviews: ReviewSession[]): void {
		items.value = reviews;
	}

	function setActive(detail: ReviewDetail | null): void {
		active.value = detail;
	}

	function setSummaryDraft(value: string): void {
		summaryDraft.value = value;
	}

	// upsertActive replaces the matching session in `items` with the latest
	// copy (created or updated) and bumps it to the front. Used by the
	// startReview flow so the picker shows the freshly created session
	// without a round-trip to listReviews.
	function upsertActive(review: ReviewSession): void {
		items.value = [review, ...items.value.filter((item) => item.id !== review.id)];
	}

	function clear(): void {
		items.value = [];
		active.value = null;
		summaryDraft.value = '';
	}

	return {
		items,
		active,
		summaryDraft,
		setItems,
		setActive,
		setSummaryDraft,
		upsertActive,
		clear,
	};
});
