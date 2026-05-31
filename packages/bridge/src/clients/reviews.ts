import type { ReviewComment, ReviewDetail, ReviewEvent, ReviewSession } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';
import { HttpRoutes } from '#bridge/routes.js';

export class ReviewClient {
	constructor(private readonly transport: HttpTransport) {}

	create(request: Partial<ReviewSession>): Promise<ReviewSession> {
		return this.transport.request<ReviewSession>('POST', HttpRoutes.reviews.base, request as Record<string, unknown>);
	}

	list(request: { limit?: number } = {}): Promise<{ reviews: ReviewSession[] }> {
		const query = typeof request.limit === 'number' ? `?limit=${request.limit}` : '';

		return this.transport.request<{ reviews: ReviewSession[] }>('GET', `${HttpRoutes.reviews.base}${query}`);
	}

	detail(request: { id: number }): Promise<ReviewDetail> {
		return this.transport.request<ReviewDetail>('GET', HttpRoutes.reviews.detail(request.id));
	}

	addEvent(request: { reviewId: number; type: string; filePath?: string; message?: string; metadata?: string }): Promise<ReviewEvent> {
		return this.transport.request<ReviewEvent>('POST', HttpRoutes.reviews.events(request.reviewId), {
			type: request.type,
			filePath: request.filePath,
			message: request.message,
			metadata: request.metadata,
		});
	}

	createComment(request: {
		reviewId: number;
		filePath: string;
		diffSection: string;
		side: string;
		lineNumber: number;
		startLineNumber?: number;
		startSide?: string;
		authorLabel: string;
		bodyHtml: string;
	}): Promise<ReviewComment> {
		return this.transport.request<ReviewComment>('POST', HttpRoutes.reviews.comments(request.reviewId), {
			filePath: request.filePath,
			diffSection: request.diffSection,
			side: request.side,
			lineNumber: request.lineNumber,
			startLineNumber: request.startLineNumber,
			startSide: request.startSide,
			authorLabel: request.authorLabel,
			bodyHtml: request.bodyHtml,
		});
	}

	updateComment(request: { reviewId: number; commentId: number; bodyHtml: string }): Promise<ReviewComment> {
		return this.transport.request<ReviewComment>('PATCH', HttpRoutes.reviews.comment(request.reviewId, request.commentId), { bodyHtml: request.bodyHtml });
	}

	deleteComment(request: { reviewId: number; commentId: number }): Promise<void> {
		return this.transport.request<void>('DELETE', HttpRoutes.reviews.comment(request.reviewId, request.commentId));
	}

	setCommentResolved(request: { reviewId: number; commentId: number; resolved: boolean }): Promise<ReviewComment> {
		return this.transport.request<ReviewComment>('PATCH', HttpRoutes.reviews.commentResolve(request.reviewId, request.commentId), { resolved: request.resolved });
	}
}
