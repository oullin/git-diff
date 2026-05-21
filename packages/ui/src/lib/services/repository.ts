import type {
    Branch,
    CommitSummary,
    FileSearchResult,
    PullRequestSummary,
    Repository,
    RepositoryCollaborator,
    RepositoryFile,
    RepositoryFileRange,
    RepositoryState,
} from "@git-diff/contracts";

/**
 * RepositoryService wraps every repository-shaped call on
 * `window.diffApp`. The renderer code consumes this surface rather than
 * the IPC shim, so the bridge can be swapped (e.g. for the
 * browser-fallback mock) without touching call sites.
 */
export interface RepositoryService {
    state(path?: string): Promise<RepositoryState>;
    open(path: string): Promise<RepositoryState>;
    refresh(path: string): Promise<RepositoryState>;
    readCommit(sha: string, path?: string): Promise<RepositoryState>;
    listCommits(path?: string, limit?: number): Promise<{ commits: CommitSummary[] }>;
    listPullRequests(
        path?: string,
        limit?: number,
    ): Promise<{ pullRequests: PullRequestSummary[] }>;
    readPullRequest(number: number, path?: string): Promise<RepositoryState>;
    readFile(root: string, path: string): Promise<RepositoryFile>;
    readFileRange(request: {
        root: string;
        path: string;
        ref?: string;
        startLine: number;
        endLine: number;
    }): Promise<RepositoryFileRange>;
    listBranches(path?: string): Promise<{ branches: string[]; records?: Branch[] }>;
    checkoutBranch(path: string, branch: string): Promise<RepositoryState>;
    createBranch(path: string, name: string): Promise<RepositoryState>;
    deleteBranch(name: string, path?: string): Promise<void>;
    lockBranch(name: string, path?: string): Promise<{ branches: Branch[] }>;
    unlockBranch(name: string, path?: string): Promise<{ branches: Branch[] }>;
    choose(defaultPath?: string): Promise<string | null>;
    list(): Promise<Repository[]>;
    searchFiles(query: string, limit?: number): Promise<FileSearchResult[]>;
    upsert(request: { path: string; name?: string }): Promise<Repository>;
    remove(path: string): Promise<void>;
    listCollaborators(path: string): Promise<RepositoryCollaborator[]>;
    addCollaborator(request: {
        path: string;
        userId: number;
        role: "write" | "read";
    }): Promise<RepositoryCollaborator>;
    removeCollaborator(request: { path: string; userId: number }): Promise<void>;
}

export function createRepositoryService(): RepositoryService {
    const api = () => window.diffApp;

    return {
        state: (path) => api().repositoryState(path),
        open: (path) => api().openRepository(path),
        refresh: (path) => api().refreshRepository(path),
        readCommit: (sha, path) => api().readCommit(sha, path),
        listCommits: (path, limit) => api().listCommits(path, limit),
        listPullRequests: (path, limit) => api().listPullRequests(path, limit),
        readPullRequest: (number, path) => api().readPullRequest(number, path),
        readFile: (root, path) => api().readRepositoryFile(root, path),
        readFileRange: (request) => api().readRepositoryFileRange(request),
        listBranches: (path) => api().listBranches(path),
        checkoutBranch: (path, branch) => api().checkoutBranch(path, branch),
        createBranch: (path, name) => api().createBranch(path, name),
        deleteBranch: (name, path) => api().deleteBranch(name, path),
        lockBranch: (name, path) => api().lockBranch(name, path),
        unlockBranch: (name, path) => api().unlockBranch(name, path),
        choose: (defaultPath) => api().chooseRepository(defaultPath),
        list: () => api().listRepositories(),
        searchFiles: (query, limit) => api().searchRepositoryFiles(query, limit),
        upsert: (request) => api().upsertRepository(request),
        remove: (path) => api().removeRepository(path),
        listCollaborators: (path) => api().listCollaborators(path),
        addCollaborator: (request) => api().addCollaborator(request),
        removeCollaborator: (request) => api().removeCollaborator(request),
    };
}
