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
export function unixTarget(socketPath) {
    return { socketPath };
}
export function createWorkflowBridgeClient(target) {
    return new HttpWorkflowBridgeClient(target);
}
export function waitForReady(client, timeoutMs = 10000) {
    const deadline = Date.now() + timeoutMs;
    return new Promise((resolve, reject) => {
        const attempt = () => {
            client
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
// HttpWorkflowBridgeClient is a thin dispatcher: every method delegates to a
// per-domain function under ./clients/. SRP lives in those modules; this
// class exists only to satisfy the flat WorkflowBridgeClient interface used
// by the renderer/electron IPC layer.
class HttpWorkflowBridgeClient {
    socketPath;
    constructor(target) {
        this.socketPath = target.socketPath;
    }
    close() {}
    healthz() {
        return systemClient.healthz(this.socketPath);
    }
    getSystemStats() {
        return systemClient.getSystemStats(this.socketPath);
    }
    getUIPreferences() {
        return preferencesClient.getUIPreferences(this.socketPath);
    }
    saveUIPreferences(values) {
        return preferencesClient.saveUIPreferences(this.socketPath, values);
    }
    getAuthState() {
        return authClient.getAuthState(this.socketPath);
    }
    authSetup(request) {
        return authClient.authSetup(this.socketPath, request);
    }
    authLogin(request) {
        return authClient.authLogin(this.socketPath, request);
    }
    authResume(request) {
        return authClient.authResume(this.socketPath, request);
    }
    authLogout() {
        return authClient.authLogout(this.socketPath);
    }
    authWipe(request) {
        return authClient.authWipe(this.socketPath, request);
    }
    repositoryState(request) {
        return repositoryClient.repositoryState(this.socketPath, request);
    }
    openRepository(request) {
        return repositoryClient.openRepository(this.socketPath, request);
    }
    refreshRepository(request) {
        return repositoryClient.refreshRepository(this.socketPath, request);
    }
    readCommit(request) {
        return repositoryClient.readCommit(this.socketPath, request);
    }
    listCommits(request) {
        return repositoryClient.listCommits(this.socketPath, request);
    }
    readRepositoryFile(request) {
        return repositoryClient.readRepositoryFile(this.socketPath, request);
    }
    generateWalkthrough(request) {
        return walkthroughClient.generateWalkthrough(this.socketPath, request);
    }
    listPullRequests(request) {
        return pullRequestClient.listPullRequests(this.socketPath, request);
    }
    readPullRequest(request) {
        return pullRequestClient.readPullRequest(this.socketPath, request);
    }
    listPendingComments(request) {
        return pendingCommentClient.listPendingComments(this.socketPath, request);
    }
    createPendingComment(request) {
        return pendingCommentClient.createPendingComment(this.socketPath, request);
    }
    updatePendingComment(request) {
        return pendingCommentClient.updatePendingComment(this.socketPath, request);
    }
    deletePendingComment(request) {
        return pendingCommentClient.deletePendingComment(this.socketPath, request);
    }
    promotePendingComments(request) {
        return pendingCommentClient.promotePendingComments(this.socketPath, request);
    }
    listBranches(request) {
        return branchClient.listBranches(this.socketPath, request);
    }
    checkoutBranch(request) {
        return branchClient.checkoutBranch(this.socketPath, request);
    }
    createBranch(request) {
        return branchClient.createBranch(this.socketPath, request);
    }
    deleteBranch(request) {
        return branchClient.deleteBranch(this.socketPath, request);
    }
    lockBranch(request) {
        return branchClient.lockBranch(this.socketPath, request);
    }
    unlockBranch(request) {
        return branchClient.unlockBranch(this.socketPath, request);
    }
    createReview(request) {
        return reviewClient.createReview(this.socketPath, request);
    }
    listReviews(request) {
        return reviewClient.listReviews(this.socketPath, request);
    }
    reviewDetail(request) {
        return reviewClient.reviewDetail(this.socketPath, request);
    }
    addReviewEvent(request) {
        return reviewClient.addReviewEvent(this.socketPath, request);
    }
    createReviewComment(request) {
        return reviewClient.createReviewComment(this.socketPath, request);
    }
    updateReviewComment(request) {
        return reviewClient.updateReviewComment(this.socketPath, request);
    }
    deleteReviewComment(request) {
        return reviewClient.deleteReviewComment(this.socketPath, request);
    }
    listRepositories() {
        return repositoriesClient.listRepositories(this.socketPath);
    }
    searchRepositoryFiles(request) {
        return repositoriesClient.searchRepositoryFiles(this.socketPath, request);
    }
    upsertRepository(request) {
        return repositoriesClient.upsertRepository(this.socketPath, request);
    }
    removeRepository(request) {
        return repositoriesClient.removeRepository(this.socketPath, request);
    }
    listCollaborators(request) {
        return repositoriesClient.listCollaborators(this.socketPath, request);
    }
    addCollaborator(request) {
        return repositoriesClient.addCollaborator(this.socketPath, request);
    }
    removeCollaborator(request) {
        return repositoriesClient.removeCollaborator(this.socketPath, request);
    }
}
