// Bridge errors thrown by the electron preload arrive as plain Error instances
// whose .message embeds a `__BRIDGE_ERROR__<json>` payload (see
// electron/ipc/branches.ipc.ts's checkout handler). parseBridgeError pulls that
// payload out so call sites can branch on .code without parsing JSON every
// time. Returns null when the error is not a structured bridge error.

export interface BridgeErrorPayload {
  code: string;
  files?: string[];
  message?: string;
}

const BRIDGE_ERROR_PREFIX = "__BRIDGE_ERROR__";

export function parseBridgeError(cause: unknown): BridgeErrorPayload | null {
  if (!(cause instanceof Error)) return null;

  const idx = cause.message.indexOf(BRIDGE_ERROR_PREFIX);

  if (idx < 0) return null;

  try {
    const parsed = JSON.parse(cause.message.slice(idx + BRIDGE_ERROR_PREFIX.length));

    if (parsed && typeof parsed.code === "string") {
      return {
        code: parsed.code,
        files: Array.isArray(parsed.files)
          ? parsed.files.filter((f: unknown) => typeof f === "string")
          : undefined,
        message: typeof parsed.message === "string" ? parsed.message : undefined,
      };
    }
  } catch {
    // not a structured bridge error
  }

  return null;
}
