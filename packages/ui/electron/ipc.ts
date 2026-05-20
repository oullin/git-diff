import { register as registerAuth } from "#electron/ipc/auth.ipc.js";
import { register as registerBranches } from "#electron/ipc/branches.ipc.js";
import { register as registerPendingComments } from "#electron/ipc/pending-comments.ipc.js";
import { register as registerPullRequests } from "#electron/ipc/pull-requests.ipc.js";
import { register as registerRepositories } from "#electron/ipc/repositories.ipc.js";
import { register as registerRepository } from "#electron/ipc/repository.ipc.js";
import { register as registerReviews } from "#electron/ipc/reviews.ipc.js";
import { register as registerSystem } from "#electron/ipc/system.ipc.js";
import { register as registerWalkthrough } from "#electron/ipc/walkthrough.ipc.js";
import type { IpcDeps } from "#electron/ipc/types.js";

export function registerIpcHandlers(deps: IpcDeps): void {
  registerRepository(deps);
  registerRepositories();
  registerBranches();
  registerPullRequests();
  registerWalkthrough();
  registerPendingComments();
  registerReviews();
  registerAuth();
  registerSystem(deps);
}
