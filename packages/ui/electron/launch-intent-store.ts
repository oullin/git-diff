import type { LaunchIntent } from "#electron/launch-intent.js";

// The launch intent is read once at startup by main.ts (or on second-instance)
// and consumed by the renderer through an IPC fetch. Treat it as a one-shot
// queue: each consumer call resets the stored value so a fresh window doesn't
// reapply an old intent.

let pending: LaunchIntent | null = null;

export function setLaunchIntent(intent: LaunchIntent | null): void {
  pending = intent;
}

export function takeLaunchIntent(): LaunchIntent | null {
  const value = pending;
  pending = null;
  return value;
}
