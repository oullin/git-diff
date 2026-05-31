import { client } from '#electron/bridge.js';
import { IpcChannels } from '#electron/ipc/routes.js';
import type { IpcRouter } from '#electron/ipc/router.js';

export function register(router: IpcRouter): void {
	router.on(IpcChannels.reviews.create, async (_event, request: Record<string, unknown>) => (await client()).reviews.create(request));

	router.on(IpcChannels.reviews.list, async (_event, limit?: number) => (await client()).reviews.list({ limit }));

	router.on(IpcChannels.reviews.detail, async (_event, id: number) => (await client()).reviews.detail({ id }));

	router.on(
		IpcChannels.reviews.event,
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
		IpcChannels.reviews.commentCreate,
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
		IpcChannels.reviews.commentUpdate,
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
		IpcChannels.reviews.commentDelete,
		async (
			_event,
			request: {
				reviewId: number;
				commentId: number;
			},
		) => (await client()).reviews.deleteComment(request),
	);

	router.on(
		IpcChannels.reviews.commentResolve,
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
