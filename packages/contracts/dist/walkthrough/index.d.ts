import type { RepositoryMode } from "../repo/index.js";
/**
 * What the user is expected to do with a file in a walkthrough group.
 * Mirrors codiff's vocabulary so the UI can colour-code consistently.
 */
export type WalkthroughAction = "review" | "scan" | "skim";
/** How broadly a change reaches across the codebase. */
export type WalkthroughImpact = "wide" | "contained" | "mechanical";
export interface WalkthroughFileEntry {
    path: string;
    note: string;
    action: WalkthroughAction;
    impact: WalkthroughImpact;
}
export interface WalkthroughGroup {
    id: string;
    title: string;
    rationale: string;
    files: WalkthroughFileEntry[];
}
export interface WalkthroughRecord {
    repoRoot: string;
    contextKind: RepositoryMode;
    contextSha?: string;
    fingerprint: string;
    providerId: string;
    modelId: string;
    groups: WalkthroughGroup[];
    summary: string;
    generatedAt: string;
    stale?: boolean;
    /**
     * @deprecated Use {@link groups}. Mirrored for one release so renderer
     * code that hasn't migrated still has data to display.
     */
    order?: string[];
    /** @deprecated Use {@link groups}. Mirrored for one release. */
    notes?: Record<string, string>;
}
//# sourceMappingURL=index.d.ts.map