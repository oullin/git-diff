import type { RepositoryMode, WalkthroughRecord } from "@git-diff/domain";

export interface WalkthroughService {
  generate(request: {
    path?: string;
    kind?: RepositoryMode;
    sha?: string;
    refresh?: boolean;
  }): Promise<WalkthroughRecord>;
}

export function createWalkthroughService(): WalkthroughService {
  return {
    generate: (request) => window.diffApp.generateWalkthrough(request),
  };
}
