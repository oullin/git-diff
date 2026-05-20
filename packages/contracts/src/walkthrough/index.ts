import type { RepositoryMode } from "../repo/index.js";

export interface WalkthroughRecord {
  repoRoot: string;
  contextKind: RepositoryMode;
  contextSha?: string;
  fingerprint: string;
  modelId: string;
  order: string[];
  notes: Record<string, string>;
  summary: string;
  generatedAt: string;
  stale?: boolean;
}
