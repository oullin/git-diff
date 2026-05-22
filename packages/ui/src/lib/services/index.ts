import { type AuthService, createAuthService } from "@lib/services/auth";
import { type PreferenceService, createPreferenceService } from "@lib/services/preferences";
import { type RepositoryService, createRepositoryService } from "@lib/services/repository";
import { type ReviewService, createReviewService } from "@lib/services/review";
import { type WalkthroughService, createWalkthroughService } from "@lib/services/walkthrough";

export type {
    AuthService,
    PreferenceService,
    RepositoryService,
    ReviewService,
    WalkthroughService,
};

/**
 * The renderer's typed bridge facade. Each field hides
 * `window.diffApp` behind a domain-shaped interface so call sites read
 * `services.auth.login(...)` instead of `window.diffApp.authLogin(...)`,
 * and the bridge can be swapped (e.g. for browser-fallback) without
 * touching consumers.
 */
export interface Container {
    readonly auth: AuthService;
    readonly preferences: PreferenceService;
    readonly repository: RepositoryService;
    readonly review: ReviewService;
    readonly walkthrough: WalkthroughService;
}

let registry: Container | null = null;

export function services(): Container {
    if (!registry) {
        registry = {
            auth: createAuthService(),
            preferences: createPreferenceService(),
            repository: createRepositoryService(),
            review: createReviewService(),
            walkthrough: createWalkthroughService(),
        };
    }

    return registry;
}

/** Swap the active service container (tests). Returns a restore func. */
export function setServices(next: Container): () => void {
    const prev = registry;

    registry = next;

    return () => {
        registry = prev;
    };
}
