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
    RepositoryFileRange,
    RepositoryMode,
    RepositoryState,
    ReviewComment,
    ReviewDetail,
    ReviewEvent,
    ReviewSession,
    SystemStats,
    UIPreferencesResponse,
    WalkthroughRecord,
} from "@git-diff/contracts";

export interface UnixTarget {
    socketPath: string;
}

// WorkflowBridgeClient was originally the macOS-workflow client that the
// gus-mac UI consumed. The diff-review app keeps the name (so the bridge's
// HTTP wiring stays familiar) but the surface is now scoped to the methods
// the renderer actually uses via diffApp. TODO(phase-12): rename to
// ApiClient + split into per-domain facades.
export interface WorkflowBridgeClient {
    close(): void;
    healthz(): Promise<void>;
    getUIPreferences(): Promise<UIPreferencesResponse>;
    saveUIPreferences(values: Record<string, string>): Promise<UIPreferencesResponse>;
    getAuthState(): Promise<AuthStateResponse>;
    authSetup(request: { password: string }): Promise<AuthLoginResponse>;
    authLogin(request: { password: string; remember: boolean }): Promise<AuthLoginResponse>;
    authResume(request: { token: string }): Promise<{ user: AuthUser }>;
    authLogout(): Promise<void>;
    authWipe(request: { osUsername?: string }): Promise<void>;
    repositoryState(request: { path?: string }): Promise<RepositoryState>;
    openRepository(request: { path: string }): Promise<RepositoryState>;
    refreshRepository(request: { path: string }): Promise<RepositoryState>;
    readCommit(request: { path?: string; sha: string }): Promise<RepositoryState>;
    listCommits(request: { path?: string; limit?: number }): Promise<{ commits: CommitSummary[] }>;
    generateWalkthrough(request: {
        path?: string;
        kind?: RepositoryMode;
        sha?: string;
        refresh?: boolean;
    }): Promise<WalkthroughRecord>;
    listPullRequests(request: {
        path?: string;
        limit?: number;
    }): Promise<{ pullRequests: PullRequestSummary[] }>;
    readPullRequest(request: { path?: string; number: number }): Promise<RepositoryState>;
    listPendingComments(request: {
        path?: string;
        kind?: RepositoryMode;
        sha?: string;
    }): Promise<{ comments: PendingComment[] }>;
    createPendingComment(request: {
        repoRoot: string;
        contextKind: RepositoryMode;
        contextSha?: string;
        filePath: string;
        diffSection: string;
        side: string;
        lineNumber: number;
        startLineNumber?: number;
        startSide?: string;
        authorLabel: string;
        bodyHtml: string;
    }): Promise<PendingComment>;
    updatePendingComment(request: { id: string; bodyHtml: string }): Promise<PendingComment>;
    deletePendingComment(request: { id: string }): Promise<void>;
    promotePendingComments(request: { reviewId: string }): Promise<{ promoted: number }>;
    readRepositoryFile(request: { root: string; path: string }): Promise<RepositoryFile>;
    readRepositoryFileRange(request: {
        root: string;
        path: string;
        ref?: string;
        startLine: number;
        endLine: number;
    }): Promise<RepositoryFileRange>;
    listBranches(request: { path?: string }): Promise<{ branches: string[]; records?: Branch[] }>;
    checkoutBranch(request: { path: string; branch: string }): Promise<RepositoryState>;
    createBranch(request: { path: string; name: string }): Promise<RepositoryState>;
    deleteBranch(request: { path?: string; name: string }): Promise<void>;
    lockBranch(request: { path?: string; name: string }): Promise<{ branches: Branch[] }>;
    unlockBranch(request: { path?: string; name: string }): Promise<{ branches: Branch[] }>;
    listCollaborators(request: {
        path: string;
    }): Promise<{ collaborators: RepositoryCollaborator[] }>;
    addCollaborator(request: {
        path: string;
        userId: number;
        role: "write" | "read";
    }): Promise<RepositoryCollaborator>;
    removeCollaborator(request: { path: string; userId: number }): Promise<void>;
    getSystemStats(): Promise<SystemStats>;
    createReview(request: Partial<ReviewSession>): Promise<ReviewSession>;
    listReviews(request: { limit?: number }): Promise<{ reviews: ReviewSession[] }>;
    reviewDetail(request: { id: string }): Promise<ReviewDetail>;
    addReviewEvent(request: {
        reviewId: string;
        type: string;
        filePath?: string;
        message?: string;
        metadata?: string;
    }): Promise<ReviewEvent>;
    createReviewComment(request: {
        reviewId: string;
        filePath: string;
        diffSection: string;
        side: string;
        lineNumber: number;
        startLineNumber?: number;
        startSide?: string;
        authorLabel: string;
        bodyHtml: string;
    }): Promise<ReviewComment>;
    updateReviewComment(request: {
        reviewId: string;
        commentId: string;
        bodyHtml: string;
    }): Promise<ReviewComment>;
    deleteReviewComment(request: { reviewId: string; commentId: string }): Promise<void>;
    listRepositories(): Promise<{ repositories: Repository[] }>;
    searchRepositoryFiles(request: {
        query: string;
        limit?: number;
    }): Promise<{ results: FileSearchResult[] }>;
    upsertRepository(request: { path: string; name?: string }): Promise<Repository>;
    removeRepository(request: { path: string }): Promise<void>;
}
