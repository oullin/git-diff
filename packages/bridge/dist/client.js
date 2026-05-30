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
import { SocketHttpTransport } from "#bridge/http.js";
class HttpApiClient {
    auth;
    branches;
    pendingComments;
    preferences;
    pullRequests;
    repositories;
    repository;
    reviews;
    system;
    userConfig;
    walkthroughs;
    constructor(transport) {
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
    close() {}
}
export function createApiClient(socketPath) {
    return new HttpApiClient(new SocketHttpTransport(socketPath));
}
export function createApiClientFromTransport(transport) {
    return new HttpApiClient(transport);
}
export function waitForReady(client, timeoutMs = 10000) {
    const deadline = Date.now() + timeoutMs;
    return new Promise((resolve, reject) => {
        const attempt = () => {
            client.system
                .healthz()
                .then(resolve)
                .catch((error) => {
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
