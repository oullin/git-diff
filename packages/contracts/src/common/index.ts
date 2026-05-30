export type GitFileStatus = "added" | "deleted" | "modified" | "renamed" | "untracked";

export type DiffSectionKind = "staged" | "unstaged" | "untracked" | "commit";

export type RepositoryMode = "working" | "commit";

export type RepositoryRole = "owner" | "write" | "read";

/** The bridge converts the flat wire shape (`contextKind` + `contextSha`)
 *  into this discriminated union; consumers should read this form. */
export type ReviewContext = { kind: "working" } | { kind: "commit"; sha: string };
