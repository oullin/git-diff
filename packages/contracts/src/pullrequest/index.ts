export interface PullRequestSummary {
    number: number;
    title: string;
    author: string;
    state: string;
    baseRef: string;
    headRef: string;
    url: string;
}
