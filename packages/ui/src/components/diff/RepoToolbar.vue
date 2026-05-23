<script setup lang="ts">
import type { CommitSummary, PullRequestSummary, RepositoryState } from "@git-diff/contracts";
import CommitPicker from "@entry/components/commits/CommitPicker.vue";
import PullRequestPicker from "@entry/components/commits/PullRequestPicker.vue";

defineProps<{
    state: RepositoryState;
    commits: CommitSummary[];
    commitsLoading: boolean;
    repoMode: "working" | "commit";
    pullRequests: PullRequestSummary[];
    pullRequestsLoading: boolean;
    activePullRequest: PullRequestSummary | null;
    walkthroughLoading: boolean;
    hasWalkthrough: boolean;
    hasActiveReview: boolean;
    copyReviewState: "idle" | "copied" | "error";
}>();

const emit = defineEmits<{
    "load-commits": [];
    "open-commit": [sha: string];
    "back-to-working": [];
    "load-pull-requests": [];
    "open-pull-request": [number: number];
    "generate-walkthrough": [];
    "copy-review-as-markdown": [];
}>();
</script>

<template>
    <div
        class="flex flex-wrap items-center gap-2 border-b border-border bg-background/70 px-4 py-1.5 text-xs"
    >
        <CommitPicker
            :commits="commits"
            :loading="commitsLoading"
            :active-sha="state.commitSha"
            :mode="repoMode"
            @open="emit('load-commits')"
            @select="(sha: string) => emit('open-commit', sha)"
            @back="emit('back-to-working')"
        />
        <PullRequestPicker
            :pull-requests="pullRequests"
            :loading="pullRequestsLoading"
            :active-number="activePullRequest?.number"
            @open="emit('load-pull-requests')"
            @select="(number: number) => emit('open-pull-request', number)"
        />
        <span
            v-if="activePullRequest"
            class="flex items-center gap-1 font-mono text-muted-foreground"
        >
            PR #{{ activePullRequest.number }}
            <span v-if="activePullRequest.title" class="font-sans"
                >· {{ activePullRequest.title }}</span
            >
        </span>
        <span
            v-else-if="repoMode === 'commit' && state.commitSha"
            class="font-mono text-muted-foreground"
        >
            commit {{ state.branch || state.commitSha.slice(0, 7) }}
        </span>
        <span class="flex-1" />
        <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded border border-border bg-background px-2.5 py-1 text-xs font-medium hover:bg-muted disabled:opacity-50"
            :disabled="walkthroughLoading"
            @click="emit('generate-walkthrough')"
        >
            {{
                walkthroughLoading
                    ? "Generating…"
                    : hasWalkthrough
                      ? "Refresh walkthrough"
                      : "Generate walkthrough"
            }}
        </button>
        <button
            v-if="hasActiveReview"
            type="button"
            class="inline-flex items-center gap-1.5 rounded border border-border bg-background px-2.5 py-1 text-xs font-medium hover:bg-muted disabled:opacity-50"
            :disabled="copyReviewState === 'copied'"
            @click="emit('copy-review-as-markdown')"
        >
            {{ copyReviewState === "copied" ? "Copied!" : "Copy review as Markdown" }}
        </button>
    </div>
</template>
