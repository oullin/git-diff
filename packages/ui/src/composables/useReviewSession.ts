import { type Ref } from 'vue';
import { sanitizeHtml } from '@ui/safe-html';
import type { PatchLine } from '@git-diff/domain/diff';
import type { LineSelectionRange } from '@composables/useLineSelection';
import { useCommentDraft } from '@composables/useCommentDraft';
import { usePendingComments } from '@composables/usePendingComments';
import { useReviewMarkdownCopy } from '@composables/useReviewMarkdownCopy';

import { viewedPrefKey, type AuthUser, type ChangedFile, type DiffSection, type RepositoryState, type ReviewComment, type ReviewDetail, type ReviewSession } from '@git-diff/domain';

/**
 * Owns the review-session domain: starting a review, viewed-state, comment
 * CRUD (review + pending drafts), and markdown export. Wraps the comment-draft,
 * pending-comment, and markdown-copy composables so App.vue is left with wiring
 * only. State (repo state, active review, reviews) is injected so the stores
 * remain the single source of truth.
 */
export interface ReviewSessionOptions {
	state: Ref<RepositoryState | null>;
	activeReview: Ref<ReviewDetail | null>;
	reviews: Ref<ReviewSession[]>;
	summaryDraft: Ref<string>;
	reviewPanelOpen: Ref<boolean>;
	currentUser: Ref<AuthUser | null>;
	prefValues: Ref<Record<string, string>>;
	savePreferences: (patch: Record<string, string>) => Promise<void>;
	onError: (message: string) => void;
}

export function useReviewSession(opts: ReviewSessionOptions) {
	const { state, activeReview, reviews, summaryDraft, reviewPanelOpen, currentUser, prefValues, savePreferences } = opts;

	const { target: commentTarget, draft: commentDraft, open: commentDialogOpen, begin: beginCommentDraft, cancel: cancelCommentDraft, close: closeCommentDraft } = useCommentDraft();

	const { items: pendingComments, reload: loadPendingComments, clear: clearPendingComments } = usePendingComments(state);

	const { state: copyReviewState, copy: copyActiveReviewAsMarkdown } = useReviewMarkdownCopy({
		onError: (cause) => opts.onError(cause instanceof Error ? cause.message : String(cause)),
	});

	function authorLabel(): string {
		return currentUser.value?.displayName || currentUser.value?.osUsername || 'You';
	}

	async function startReview() {
		if (!state.value) {
			return;
		}

		const ctx = state.value;

		const review = await window.diffApp.createReview({
			repoRoot: ctx.root,
			branch: ctx.branch,
			headSha: ctx.headSha,
			title: ctx.mode === 'commit' ? `Review commit ${ctx.branch || ctx.commitSha?.slice(0, 7) || ctx.headSha}` : `Review ${ctx.branch || ctx.headSha || 'local changes'}`,
			summary: sanitizeHtml(summaryDraft.value),
			filesChanged: ctx.files.length,
			additions: ctx.additions,
			deletions: ctx.deletions,
			contextKind: ctx.mode,
			contextSha: ctx.commitSha,
		});
		// Promote any drafts that were taken before the review session existed.
		try {
			await window.diffApp.promotePendingComments(review.id);
		} catch {
			// Promotion failure shouldn't abort review start.
		}

		clearPendingComments();

		activeReview.value = await window.diffApp.reviewDetail(review.id);

		reviews.value = [review, ...reviews.value.filter((item) => item.id !== review.id)];
		summaryDraft.value = '';
		reviewPanelOpen.value = false;
	}

	async function toggleViewed(file: ChangedFile) {
		if (!state.value) {
			return;
		}

		const key = viewedPrefKey(state.value.root, file.path);
		const currentlyViewed = isViewed(file);

		await savePreferences({ [key]: currentlyViewed ? '' : file.fingerprint });

		if (activeReview.value) {
			await window.diffApp.addReviewEvent({
				reviewId: activeReview.value.review.id,
				type: currentlyViewed ? 'file_unviewed' : 'file_viewed',
				filePath: file.path,
				message: currentlyViewed ? 'Marked unviewed' : 'Marked viewed',
			});

			activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
		}
	}

	function isViewed(file: ChangedFile): boolean {
		if (!state.value) {
			return false;
		}

		return prefValues.value[viewedPrefKey(state.value.root, file.path)] === file.fingerprint;
	}

	function openCommentForLine(file: ChangedFile, section: DiffSection, line: PatchLine, range?: LineSelectionRange) {
		const lineNumber = line.newLine ?? line.oldLine;

		if (!lineNumber || !activeReview.value) {
			reviewPanelOpen.value = true;

			return;
		}

		if (range) {
			beginCommentDraft({
				file,
				section,
				line: range.endLine,
				side: range.endSide,
				startLine: range.startLine,
				startSide: range.startSide,
			});

			return;
		}

		beginCommentDraft({
			file,
			section,
			line: lineNumber,
			side: line.newLine ? 'right' : 'left',
		});
	}

	async function saveComment() {
		if (!commentTarget.value || !commentDraft.value.trim() || !state.value) {
			return;
		}

		const target = commentTarget.value;
		const author = authorLabel();
		const body = sanitizeHtml(commentDraft.value);

		if (activeReview.value) {
			await window.diffApp.createReviewComment({
				reviewId: activeReview.value.review.id,
				filePath: target.file.path,
				diffSection: target.section.kind,
				side: target.side,
				lineNumber: target.line,
				startLineNumber: target.startLine,
				startSide: target.startSide,
				authorLabel: author,
				bodyHtml: body,
			});

			activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
		} else {
			// No active review yet — store as a draft. On startReview, draft comments
			// for this (repo, context) are promoted into the new review session.
			await window.diffApp.createPendingComment({
				repoRoot: state.value.root,
				contextKind: state.value.mode,
				contextSha: state.value.commitSha,
				filePath: target.file.path,
				diffSection: target.section.kind,
				side: target.side,
				lineNumber: target.line,
				startLineNumber: target.startLine,
				startSide: target.startSide,
				authorLabel: author,
				bodyHtml: body,
			});

			await loadPendingComments();
		}

		closeCommentDraft();
	}

	function cancelComment() {
		cancelCommentDraft();
	}

	async function deleteComment(comment: ReviewComment) {
		if (!activeReview.value) {
			return;
		}

		await window.diffApp.deleteReviewComment({
			reviewId: activeReview.value.review.id,
			commentId: comment.id,
		});

		activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
	}

	async function resolveComment(comment: ReviewComment, resolved: boolean) {
		if (!activeReview.value) {
			return;
		}

		await window.diffApp.setReviewCommentResolved({
			reviewId: activeReview.value.review.id,
			commentId: comment.id,
			resolved,
		});

		activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
	}

	async function replyToComment(parent: ReviewComment, bodyHtml: string) {
		if (!activeReview.value) {
			return;
		}

		await window.diffApp.createReviewComment({
			reviewId: activeReview.value.review.id,
			filePath: parent.filePath,
			diffSection: parent.diffSection,
			side: parent.side,
			lineNumber: parent.lineNumber,
			authorLabel: authorLabel(),
			bodyHtml: sanitizeHtml(bodyHtml),
		});

		activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
	}

	async function copyReviewAsMarkdown() {
		if (activeReview.value) {
			await copyActiveReviewAsMarkdown(activeReview.value);
		}
	}

	return {
		commentTarget,
		commentDraft,
		commentDialogOpen,
		pendingComments,
		loadPendingComments,
		clearPendingComments,
		copyReviewState,
		startReview,
		toggleViewed,
		isViewed,
		openCommentForLine,
		saveComment,
		cancelComment,
		deleteComment,
		resolveComment,
		replyToComment,
		copyReviewAsMarkdown,
	};
}
