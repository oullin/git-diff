import { client } from '#electron/bridge.js';
import type { IpcRouter } from '#electron/ipc/router.js';

export function register(router: IpcRouter): void {
	router.on('reviews:create', async (_event, request: Record<string, unknown>) => (await client()).reviews.create(request));

	router.on('reviews:list', async (_event, limit?: number) => (await client()).reviews.list({ limit }));

	router.on('reviews:detail', async (_event, id: number) => (await client()).reviews.detail({ id }));

	router.on(
		'reviews:event',
		async (
			_event,
			request: {
				reviewId: number;
				type: string;
				filePath?: string;
				message?: string;
				metadata?: string;
			},
		) => (await client()).reviews.addEvent(request),
	);

	router.on(
		'reviews:comment:create',
		async (
			_event,
			request: {
				reviewId: number;
				filePath: string;
				diffSection: string;
				side: string;
				lineNumber: number;
				authorLabel: string;
				bodyHtml: string;
			},
		) => (await client()).reviews.createComment(request),
	);

	router.on(
		'reviews:comment:update',
		async (
			_event,
			request: {
				reviewId: number;
				commentId: number;
				bodyHtml: string;
			},
		) => (await client()).reviews.updateComment(request),
	);

	router.on(
		'reviews:comment:delete',
		async (
			_event,
			request: {
				reviewId: number;
				commentId: number;
			},
		) => (await client()).reviews.deleteComment(request),
	);

	router.on(
		'reviews:comment:resolve',
		async (
			_event,
			request: {
				reviewId: number;
				commentId: number;
				resolved: boolean;
			},
		) => (await client()).reviews.setCommentResolved(request),
	);
}
