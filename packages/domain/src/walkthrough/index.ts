import type { RepositoryMode } from '#domain/repo/index.js';

export type WalkthroughAction = 'review' | 'scan' | 'skim';

export type WalkthroughImpact = 'wide' | 'contained' | 'mechanical';

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
	/** @deprecated Use {@link groups}. Mirrored for one release. */
	order?: string[];
	/** @deprecated Use {@link groups}. Mirrored for one release. */
	notes?: Record<string, string>;
}
