import type {
    AuthBootstrapResponse,
    AuthLoginResponse,
    AuthStateResponse,
    Branch,
    CommitSummary,
    FileSearchResult,
    LaunchIntent,
    PendingComment,
    PullRequestSummary,
    Repository,
    RepositoryCollaborator,
    RepositoryFile,
    RepositoryMode,
    RepositoryState,
    ReviewComment,
    ReviewDetail,
    ReviewEvent,
    ReviewSession,
    SystemStats,
    UIPreferences,
    WalkthroughRecord,
} from "@git-diff/contracts";

export interface DiffAppApi {
    takeLaunchIntent(): Promise<LaunchIntent | null>;
    onLaunchIntent(handler: (intent: LaunchIntent) => void): () => void;
    openNewWindow(repoPath?: string): Promise<void>;
    repositoryState(path?: string): Promise<RepositoryState>;
    openRepository(path: string): Promise<RepositoryState>;
    refreshRepository(path: string): Promise<RepositoryState>;
    readCommit(sha: string, path?: string): Promise<RepositoryState>;
    listCommits(path?: string, limit?: number): Promise<{ commits: CommitSummary[] }>;
    generateWalkthrough(request: {
        path?: string;
        kind?: RepositoryMode;
        sha?: string;
        refresh?: boolean;
    }): Promise<WalkthroughRecord>;
    listPullRequests(
        path?: string,
        limit?: number,
    ): Promise<{ pullRequests: PullRequestSummary[] }>;
    readPullRequest(number: number, path?: string): Promise<RepositoryState>;
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
        authorLabel: string;
        bodyHtml: string;
    }): Promise<PendingComment>;
    updatePendingComment(request: { id: string; bodyHtml: string }): Promise<PendingComment>;
    deletePendingComment(id: string): Promise<void>;
    promotePendingComments(reviewId: string): Promise<{ promoted: number }>;
    readRepositoryFile(root: string, path: string): Promise<RepositoryFile>;
    listBranches(path?: string): Promise<{ branches: string[]; records?: Branch[] }>;
    checkoutBranch(path: string, branch: string): Promise<RepositoryState>;
    createBranch(path: string, name: string): Promise<RepositoryState>;
    deleteBranch(name: string, path?: string): Promise<void>;
    lockBranch(name: string, path?: string): Promise<{ branches: Branch[] }>;
    unlockBranch(name: string, path?: string): Promise<{ branches: Branch[] }>;
    chooseRepository(defaultPath?: string): Promise<string | null>;
    listRepositories(): Promise<Repository[]>;
    searchRepositoryFiles(query: string, limit?: number): Promise<FileSearchResult[]>;
    upsertRepository(request: { path: string; name?: string }): Promise<Repository>;
    removeRepository(path: string): Promise<void>;
    listCollaborators(path: string): Promise<RepositoryCollaborator[]>;
    addCollaborator(request: {
        path: string;
        userId: number;
        role: "write" | "read";
    }): Promise<RepositoryCollaborator>;
    removeCollaborator(request: { path: string; userId: number }): Promise<void>;
    createReview(request: Partial<ReviewSession>): Promise<ReviewSession>;
    listReviews(limit?: number): Promise<{ reviews: ReviewSession[] }>;
    reviewDetail(id: string): Promise<ReviewDetail>;
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
        authorLabel: string;
        bodyHtml: string;
    }): Promise<ReviewComment>;
    updateReviewComment(request: {
        reviewId: string;
        commentId: string;
        bodyHtml: string;
    }): Promise<ReviewComment>;
    deleteReviewComment(request: { reviewId: string; commentId: string }): Promise<void>;
    getUIPreferences(): Promise<UIPreferences>;
    saveUIPreferences(patch: Record<string, string>): Promise<UIPreferences>;
    getAuthState(): Promise<AuthStateResponse>;
    authBootstrap(): Promise<AuthBootstrapResponse>;
    authSetup(password: string): Promise<AuthLoginResponse>;
    authLogin(password: string, remember: boolean): Promise<AuthLoginResponse>;
    authLogout(): Promise<void>;
    authWipe(osUsername?: string): Promise<void>;
    openDevTools(): Promise<void>;
    getSystemStats(): Promise<SystemStats>;
}

declare global {
    interface Window {
        diffApp: DiffAppApi;
    }
}
