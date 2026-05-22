import type { UIAccent } from "@git-diff/contracts";

export interface Accent {
    key: UIAccent;
    name: string;
    hex: string;
    strong: string;
    soft: string;
}

// Pierre is monochrome — there is no colored accent system. Every key resolves
// to the same neutral preset so existing user preferences ("copper", "indigo", …)
// keep working without a contracts-package migration. Re-introduce variants here
// if a future theme reinstates colored accents.
const PIERRE_NEUTRAL: Omit<Accent, "key" | "name"> = {
    hex: "oklch(20.5% 0 0)",
    strong: "oklch(14.5% 0 0)",
    soft: "oklch(97% 0 0)",
};

export const ACCENTS: Record<UIAccent, Accent> = {
    copper: { key: "copper", name: "Pierre", ...PIERRE_NEUTRAL },
    indigo: { key: "indigo", name: "Pierre", ...PIERRE_NEUTRAL },
    emerald: { key: "emerald", name: "Pierre", ...PIERRE_NEUTRAL },
    amber: { key: "amber", name: "Pierre", ...PIERRE_NEUTRAL },
    rose: { key: "rose", name: "Pierre", ...PIERRE_NEUTRAL },
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
