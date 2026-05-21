import type {
    AuthLoginResponse,
    AuthStateResponse,
    AuthUser,
    Branch,
    CommitSummary,
    FileSearchResult,
    PendingComment,
    PullRequestSummary,
    Repository,
    RepositoryCollaborator,
    RepositoryFile,
    RepositoryState,
    ReviewComment,
    ReviewDetail,
    ReviewEvent,
    ReviewSession,
    SystemStats,
    UIPreferencesResponse,
    WalkthroughRecord,
} from "@git-diff/contracts";
import type { UnixTarget, WorkflowBridgeClient } from "#bridge/client-types.js";
import * as authClient from "#bridge/clients/auth.js";
import * as branchClient from "#bridge/clients/branches.js";
import * as pendingCommentClient from "#bridge/clients/pending-comments.js";
import * as preferencesClient from "#bridge/clients/preferences.js";
import * as pullRequestClient from "#bridge/clients/pull-requests.js";
import * as repositoriesClient from "#bridge/clients/repositories.js";
import * as repositoryClient from "#bridge/clients/repository.js";
import * as reviewClient from "#bridge/clients/reviews.js";
import * as systemClient from "#bridge/clients/system.js";
import * as walkthroughClient from "#bridge/clients/walkthrough.js";

export function unixTarget(socketPath: string): UnixTarget {
    return { socketPath };
}

export function createWorkflowBridgeClient(target: UnixTarget): WorkflowBridgeClient {
    return new HttpWorkflowBridgeClient(target);
}

export function waitForReady(client: WorkflowBridgeClient, timeoutMs = 10000): Promise<void> {
    const deadline = Date.now() + timeoutMs;

    return new Promise<void>((resolve, reject) => {
        const attempt = (): void => {
            client
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

// HttpWorkflowBridgeClient is a thin dispatcher: every method delegates to a
// per-domain function under ./clients/. SRP lives in those modules; this
// class exists only to satisfy the flat WorkflowBridgeClient interface used
// by the renderer/electron IPC layer.
class HttpWorkflowBridgeClient implements WorkflowBridgeClient {
    private readonly socketPath: string;

    constructor(target: UnixTarget) {
        this.socketPath = target.socketPath;
    }

    close(): void {}

    healthz(): Promise<void> {
        return systemClient.healthz(this.socketPath);
    }

    getSystemStats(): Promise<SystemStats> {
        return systemClient.getSystemStats(this.socketPath);
    }

    getUIPreferences(): Promise<UIPreferencesResponse> {
        return preferencesClient.getUIPreferences(this.socketPath);
    }

    saveUIPreferences(values: Record<string, string>): Promise<UIPreferencesResponse> {
        return preferencesClient.saveUIPreferences(this.socketPath, values);
    }

    getAuthState(): Promise<AuthStateResponse> {
        return authClient.getAuthState(this.socketPath);
    }

    authSetup(request: { password: string }): Promise<AuthLoginResponse> {
        return authClient.authSetup(this.socketPath, request);
    }

    authLogin(request: { password: string; remember: boolean }): Promise<AuthLoginResponse> {
        return authClient.authLogin(this.socketPath, request);
    }

    authResume(request: { token: string }): Promise<{ user: AuthUser }> {
        return authClient.authResume(this.socketPath, request);
    }

    authLogout(): Promise<void> {
        return authClient.authLogout(this.socketPath);
    }

    authWipe(request: { osUsername?: string }): Promise<void> {
        return authClient.authWipe(this.socketPath, request);
    }

    repositoryState(request: { path?: string }): Promise<RepositoryState> {
        return repositoryClient.repositoryState(this.socketPath, request);
    }

    openRepository(request: { path: string }): Promise<RepositoryState> {
        return repositoryClient.openRepository(this.socketPath, request);
    }

    refreshRepository(request: { path: string }): Promise<RepositoryState> {
        return repositoryClient.refreshRepository(this.socketPath, request);
    }

    readCommit(request: { path?: string; sha: string }): Promise<RepositoryState> {
        return repositoryClient.readCommit(this.socketPath, request);
    }

    listCommits(request: { path?: string; limit?: number }): Promise<{ commits: CommitSummary[] }> {
        return repositoryClient.listCommits(this.socketPath, request);
    }

    readRepositoryFile(request: { root: string; path: string }): Promise<RepositoryFile> {
        return repositoryClient.readRepositoryFile(this.socketPath, request);
    }

    generateWalkthrough(request: {
        path?: string;
        kind?: "working" | "commit";
        sha?: string;
        refresh?: boolean;
    }): Promise<WalkthroughRecord> {
        return walkthroughClient.generateWalkthrough(this.socketPath, request);
    }

    listPullRequests(request: {
        path?: string;
        limit?: number;
    }): Promise<{ pullRequests: PullRequestSummary[] }> {
        return pullRequestClient.listPullRequests(this.socketPath, request);
    }

    readPullRequest(request: { path?: string; number: number }): Promise<RepositoryState> {
        return pullRequestClient.readPullRequest(this.socketPath, request);
    }

    listPendingComments(request: {
        path?: string;
        kind?: "working" | "commit";
        sha?: string;
    }): Promise<{ comments: PendingComment[] }> {
        return pendingCommentClient.listPendingComments(this.socketPath, request);
    }

    createPendingComment(request: {
        repoRoot: string;
        contextKind: "working" | "commit";
        contextSha?: string;
        filePath: string;
        diffSection: string;
        side: string;
        lineNumber: number;
        authorLabel: string;
        bodyHtml: string;
    }): Promise<PendingComment> {
        return pendingCommentClient.createPendingComment(this.socketPath, request);
    }

    updatePendingComment(request: { id: string; bodyHtml: string }): Promise<PendingComment> {
        return pendingCommentClient.updatePendingComment(this.socketPath, request);
    }

    deletePendingComment(request: { id: string }): Promise<void> {
        return pendingCommentClient.deletePendingComment(this.socketPath, request);
    }

    promotePendingComments(request: { reviewId: string }): Promise<{ promoted: number }> {
        return pendingCommentClient.promotePendingComments(this.socketPath, request);
    }

    listBranches(request: { path?: string }): Promise<{ branches: string[]; records?: Branch[] }> {
        return branchClient.listBranches(this.socketPath, request);
    }

    checkoutBranch(request: { path: string; branch: string }): Promise<RepositoryState> {
        return branchClient.checkoutBranch(this.socketPath, request);
    }

    createBranch(request: { path: string; name: string }): Promise<RepositoryState> {
        return branchClient.createBranch(this.socketPath, request);
    }

    deleteBranch(request: { path?: string; name: string }): Promise<void> {
        return branchClient.deleteBranch(this.socketPath, request);
    }

    lockBranch(request: { path?: string; name: string }): Promise<{ branches: Branch[] }> {
        return branchClient.lockBranch(this.socketPath, request);
    }

    unlockBranch(request: { path?: string; name: string }): Promise<{ branches: Branch[] }> {
        return branchClient.unlockBranch(this.socketPath, request);
    }

    createReview(request: Partial<ReviewSession>): Promise<ReviewSession> {
        return reviewClient.createReview(this.socketPath, request);
    }

    listReviews(request: { limit?: number }): Promise<{ reviews: ReviewSession[] }> {
        return reviewClient.listReviews(this.socketPath, request);
    }

    reviewDetail(request: { id: string }): Promise<ReviewDetail> {
        return reviewClient.reviewDetail(this.socketPath, request);
    }

    addReviewEvent(request: {
        reviewId: string;
        type: string;
        filePath?: string;
        message?: string;
        metadata?: string;
    }): Promise<ReviewEvent> {
        return reviewClient.addReviewEvent(this.socketPath, request);
    }

    createReviewComment(request: {
        reviewId: string;
        filePath: string;
        diffSection: string;
        side: string;
        lineNumber: number;
        authorLabel: string;
        bodyHtml: string;
    }): Promise<ReviewComment> {
        return reviewClient.createReviewComment(this.socketPath, request);
    }

    updateReviewComment(request: {
        reviewId: string;
        commentId: string;
        bodyHtml: string;
    }): Promise<ReviewComment> {
        return reviewClient.updateReviewComment(this.socketPath, request);
    }

    deleteReviewComment(request: { reviewId: string; commentId: string }): Promise<void> {
        return reviewClient.deleteReviewComment(this.socketPath, request);
    }

    listRepositories(): Promise<{ repositories: Repository[] }> {
        return repositoriesClient.listRepositories(this.socketPath);
    }

    searchRepositoryFiles(request: {
        query: string;
        limit?: number;
    }): Promise<{ results: FileSearchResult[] }> {
        return repositoriesClient.searchRepositoryFiles(this.socketPath, request);
    }

    upsertRepository(request: { path: string; name?: string }): Promise<Repository> {
        return repositoriesClient.upsertRepository(this.socketPath, request);
    }

    removeRepository(request: { path: string }): Promise<void> {
        return repositoriesClient.removeRepository(this.socketPath, request);
    }

    listCollaborators(request: {
        path: string;
    }): Promise<{ collaborators: RepositoryCollaborator[] }> {
        return repositoriesClient.listCollaborators(this.socketPath, request);
    }

    addCollaborator(request: {
        path: string;
        userId: number;
        role: "write" | "read";
    }): Promise<RepositoryCollaborator> {
        return repositoriesClient.addCollaborator(this.socketPath, request);
    }

    removeCollaborator(request: { path: string; userId: number }): Promise<void> {
        return repositoriesClient.removeCollaborator(this.socketPath, request);
    }
}
