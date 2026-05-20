import type { OpItem, OpVault } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function listOpVaults(socketPath: string): Promise<{ vaults: OpVault[] }> {
  return requestJson<{ vaults: OpVault[] }>(socketPath, "GET", "/v1/onepassword/vaults");
}

export function listOpItems(
  socketPath: string,
  request: { vault: string },
): Promise<{ items: OpItem[] }> {
  return requestJson<{ items: OpItem[] }>(
    socketPath,
    "GET",
    `/v1/onepassword/items?vault=${encodeURIComponent(request.vault)}`,
  );
}
