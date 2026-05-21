/**
 * Shared enums and unions used across more than one domain module.
 * Domain modules re-export the names they need so consumers can keep
 * importing them from the domain they care about.
 */
export type GitFileStatus = "added" | "deleted" | "modified" | "renamed" | "untracked";
export type DiffSectionKind = "staged" | "unstaged" | "untracked" | "commit";
export type RepositoryMode = "working" | "commit";
export type RepositoryRole = "owner" | "write" | "read";
/**
 * The context in which a review or pending comment was created.
 * The bridge converts the flat wire shape (`contextKind` + `contextSha`)
 * into this discriminated union at its boundary; consumers should always
 * read this typed form.
 */
export type ReviewContext = {
    kind: "working";
} | {
    kind: "commit";
    sha: string;
};
//# sourceMappingURL=index.d.ts.map