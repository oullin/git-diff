import { AuthClient } from "#bridge/clients/auth.js";
import { BranchClient } from "#bridge/clients/branches.js";
import { PendingCommentClient } from "#bridge/clients/pending-comments.js";
import { PreferenceClient } from "#bridge/clients/preferences.js";
import { PullRequestClient } from "#bridge/clients/pull-requests.js";
import { RepositoriesClient } from "#bridge/clients/repositories.js";
import { RepositoryClient } from "#bridge/clients/repository.js";
import { ReviewClient } from "#bridge/clients/reviews.js";
import { SystemClient } from "#bridge/clients/system.js";
import { WalkthroughClient } from "#bridge/clients/walkthrough.js";
import { type HttpTransport } from "#bridge/http.js";
/**
 * ApiClient bundles every per-domain client. Consumers depend on the
 * shape of this object (one field per domain) rather than the wire
 * format or the transport — this keeps Electron IPC handlers stable
 * across transport changes.
 */
export interface ApiClient {
    readonly auth: AuthClient;
    readonly branches: BranchClient;
    readonly pendingComments: PendingCommentClient;
    readonly preferences: PreferenceClient;
    readonly pullRequests: PullRequestClient;
    readonly repositories: RepositoriesClient;
    readonly repository: RepositoryClient;
    readonly reviews: ReviewClient;
    readonly system: SystemClient;
    readonly walkthroughs: WalkthroughClient;
    close(): void;
}
export declare function createApiClient(socketPath: string): ApiClient;
export declare function createApiClientFromTransport(transport: HttpTransport): ApiClient;
export declare function waitForReady(client: ApiClient, timeoutMs?: number): Promise<void>;
