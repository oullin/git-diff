import type { WalkthroughRecord } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function generateWalkthrough(
  socketPath: string,
  request: {
    path?: string;
    kind?: "working" | "commit";
    sha?: string;
    refresh?: boolean;
  },
): Promise<WalkthroughRecord> {
  return requestJson<WalkthroughRecord>(socketPath, "POST", "/v1/walkthrough", {
    path: request.path ?? "",
    kind: request.kind ?? "working",
    sha: request.sha ?? "",
    refresh: request.refresh ?? false,
  });
}
