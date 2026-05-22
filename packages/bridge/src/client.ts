import { AuthClient } from "#bridge/clients/auth.js";
import { BranchClient } from "#bridge/clients/branches.js";
import { PendingCommentClient } from "#bridge/clients/pending-comments.js";
import { PreferenceClient } from "#bridge/clients/preferences.js";
import { PullRequestClient } from "#bridge/clients/pull-requests.js";
import { RepositoriesClient } from "#bridge/clients/repositories.js";
import { RepositoryClient } from "#bridge/clients/repository.js";
import { ReviewClient } from "#bridge/clients/reviews.js";
import { SystemClient } from "#bridge/clients/system.js";
import { UserConfigClient } from "#bridge/clients/userconfig.js";
import { WalkthroughClient } from "#bridge/clients/walkthrough.js";
import { type HttpTransport, SocketHttpTransport } from "#bridge/http.js";

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
    readonly userConfig: UserConfigClient;
    readonly walkthroughs: WalkthroughClient;
    close(): void;
}

class HttpApiClient implements ApiClient {
    readonly auth: AuthClient;
    readonly branches: BranchClient;
    readonly pendingComments: PendingCommentClient;
    readonly preferences: PreferenceClient;
    readonly pullRequests: PullRequestClient;
    readonly repositories: RepositoriesClient;
    readonly repository: RepositoryClient;
    readonly reviews: ReviewClient;
    readonly system: SystemClient;
    readonly userConfig: UserConfigClient;
    readonly walkthroughs: WalkthroughClient;

    constructor(transport: HttpTransport) {
        this.auth = new AuthClient(transport);
        this.branches = new BranchClient(transport);
        this.pendingComments = new PendingCommentClient(transport);
        this.preferences = new PreferenceClient(transport);
        this.pullRequests = new PullRequestClient(transport);
        this.repositories = new RepositoriesClient(transport);
        this.repository = new RepositoryClient(transport);
        this.reviews = new ReviewClient(transport);
        this.system = new SystemClient(transport);
        this.userConfig = new UserConfigClient(transport);
        this.walkthroughs = new WalkthroughClient(transport);
    }

    close(): void {}
}

export function createApiClient(socketPath: string): ApiClient {
    return new HttpApiClient(new SocketHttpTransport(socketPath));
}

export function createApiClientFromTransport(transport: HttpTransport): ApiClient {
    return new HttpApiClient(transport);
}

export function waitForReady(client: ApiClient, timeoutMs = 10000): Promise<void> {
    const deadline = Date.now() + timeoutMs;

    return new Promise<void>((resolve, reject) => {
        const attempt = (): void => {
            client.system
                .healthz()
                .then(resolve)
                .catch((error: unknown) => {
                    if (Date.now() >= deadline) {
                        reject(error);

                        return;
                    }

                    setTimeout(attempt, 100);
                });
        };

        attempt();
    });
}
