import type { UIAccent } from "@git-diff/domain";

export interface Accent {
  key: UIAccent;
  name: string;
  hex: string;
  strong: string;
  soft: string;
}

// GitHub Primer has a single colored accent (blue). Every key resolves to the same
// preset so existing user preferences ("copper", "indigo", …) keep working without a
// contracts-package migration. The preset points at the per-theme `--primer-accent`
// token (defined in style.css) so the inline override applyAccent() writes resolves to
// Primer light blue (#0969da) or Primer dark blue (#2f81f7) automatically.
const PRIMER_BLUE: Omit<Accent, "key" | "name"> = {
  hex: "var(--primer-accent)",
  strong: "var(--primer-accent-strong)",
  soft: "var(--primer-accent-soft)",
};

export const ACCENTS: Record<UIAccent, Accent> = {
  copper: { key: "copper", name: "GitHub", ...PRIMER_BLUE },
  indigo: { key: "indigo", name: "GitHub", ...PRIMER_BLUE },
  emerald: { key: "emerald", name: "GitHub", ...PRIMER_BLUE },
  amber: { key: "amber", name: "GitHub", ...PRIMER_BLUE },
  rose: { key: "rose", name: "GitHub", ...PRIMER_BLUE },
};

export function applyAccent(accent: Accent): void {
  const root = document.documentElement.style;

  root.setProperty("--gd-accent", accent.hex);
  root.setProperty("--gd-accent-strong", accent.strong);
  root.setProperty("--gd-accent-soft", accent.soft);
}

export function resolveAccent(value: string | undefined): Accent {
  if (value && value in ACCENTS) {
    return ACCENTS[value as UIAccent];
  }

  return ACCENTS.copper;
}

export type DiffStyleColors = {
  addBg: string;
  addStrong: string;
  remBg: string;
  remStrong: string;
  addNum: string;
  remNum: string;
  addBar: string;
  remBar: string;
};

// Pierre uses solid fills with a saturated 4px edge stripe. The three styles
// keep the same shape so the existing toggle in TweaksPanel still works:
//   soft   — the default Pierre look (filled row, soft tint)
//   punchy — same fill but a stronger gutter strip for high-contrast scanning
//   bar    — transparent rows with only the edge stripe (minimal mode)
export function diffBgs(
  style: "soft" | "punchy" | "bar",
  added = "var(--gd-add-border)",
  removed = "var(--gd-del-border)",
): DiffStyleColors {
  if (style === "punchy") {
    return {
      addBg: "var(--diff-added-bg)",
      addStrong: "var(--diff-added-bg-strong)",
      remBg: "var(--diff-removed-bg)",
      remStrong: "var(--diff-removed-bg-strong)",
      addNum: "var(--diff-added-bg-strong)",
      remNum: "var(--diff-removed-bg-strong)",
      addBar: added,
      remBar: removed,
    };
  }

  if (style === "bar") {
    return {
      addBg: "transparent",
      addStrong: "var(--diff-added-bg)",
      remBg: "transparent",
      remStrong: "var(--diff-removed-bg)",
      addNum: "transparent",
      remNum: "transparent",
      addBar: added,
      remBar: removed,
    };
  }

  return {
    addBg: "var(--diff-added-bg)",
    addStrong: "var(--diff-added-bg-strong)",
    remBg: "var(--diff-removed-bg)",
    remStrong: "var(--diff-removed-bg-strong)",
    addNum: "var(--diff-added-bg-strong)",
    remNum: "var(--diff-removed-bg-strong)",
    addBar: added,
    remBar: removed,
  };
}
