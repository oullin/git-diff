import { computed, type ComputedRef } from "vue";
import type { ThemeChoice } from "@composables/useTheme";

import {
  PREF_KEYS,
  type DiffDensity,
  type DiffHunkStyle,
  type DiffViewMode,
  type UIAccent,
} from "@git-diff/domain";

export interface Tweaks {
  accent: UIAccent;
  viewMode: DiffViewMode;
  diffStyle: DiffHunkStyle;
  density: DiffDensity;
  showMinimap: boolean;
  showStatusBar: boolean;
  wordHighlight: boolean;
  hideResolved: boolean;
  theme: ThemeChoice;
}

export const TWEAK_DEFAULTS: Tweaks = {
  accent: "copper",
  viewMode: "split",
  diffStyle: "soft",
  density: "comfortable",
  showMinimap: true,
  showStatusBar: true,
  wordHighlight: true,
  hideResolved: false,
  theme: "system",
};

function bool(value: string | undefined, fallback: boolean): boolean {
  if (value === "1") {
    return true;
  }

  if (value === "") {
    return false;
  }

  if (value === "0") {
    return false;
  }

  return fallback;
}

function oneOf<T extends string>(value: string | undefined, options: readonly T[], fallback: T): T {
  return options.includes(value as T) ? (value as T) : fallback;
}

export function useTweaks(prefs: ComputedRef<Record<string, string>>): ComputedRef<Tweaks> {
  return computed<Tweaks>(() => {
    const p = prefs.value;

    return {
      accent: oneOf(
        p[PREF_KEYS.uiAccent],
        ["copper", "indigo", "emerald", "amber", "rose"] as const,
        TWEAK_DEFAULTS.accent,
      ),
      viewMode: oneOf(
        p[PREF_KEYS.diffViewMode],
        ["split", "unified"] as const,
        TWEAK_DEFAULTS.viewMode,
      ),
      diffStyle: oneOf(
        p[PREF_KEYS.diffStyle],
        ["soft", "punchy", "bar"] as const,
        TWEAK_DEFAULTS.diffStyle,
      ),
      density: oneOf(
        p[PREF_KEYS.diffDensity],
        ["comfortable", "compact"] as const,
        TWEAK_DEFAULTS.density,
      ),
      showMinimap: bool(p[PREF_KEYS.uiShowMinimap], TWEAK_DEFAULTS.showMinimap),
      showStatusBar: bool(p[PREF_KEYS.uiShowStatusBar], TWEAK_DEFAULTS.showStatusBar),
      wordHighlight: bool(p[PREF_KEYS.diffWordHi], TWEAK_DEFAULTS.wordHighlight),
      hideResolved: bool(p[PREF_KEYS.diffHideResolved], TWEAK_DEFAULTS.hideResolved),
      theme: oneOf(p[PREF_KEYS.theme], ["light", "dark", "system"] as const, TWEAK_DEFAULTS.theme),
    };
  });
}

export function tweakPrefPatch(
  key: keyof Tweaks,
  value: Tweaks[keyof Tweaks],
): Record<string, string> {
  switch (key) {
    case "accent":
      return { [PREF_KEYS.uiAccent]: String(value) };

    case "viewMode":
      return { [PREF_KEYS.diffViewMode]: String(value) };

    case "diffStyle":
      return { [PREF_KEYS.diffStyle]: String(value) };

    case "density":
      return { [PREF_KEYS.diffDensity]: String(value) };

    case "showMinimap":
      return { [PREF_KEYS.uiShowMinimap]: value ? "1" : "0" };

    case "showStatusBar":
      return { [PREF_KEYS.uiShowStatusBar]: value ? "1" : "0" };

    case "wordHighlight":
      return { [PREF_KEYS.diffWordHi]: value ? "1" : "0" };

    case "hideResolved":
      return { [PREF_KEYS.diffHideResolved]: value ? "1" : "0" };

    case "theme":
      return { [PREF_KEYS.theme]: String(value) };
  }
}
