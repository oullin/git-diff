import type { UIAccent } from "@api";

export interface Accent {
  key: UIAccent;
  name: string;
  hex: string;
  strong: string;
  soft: string;
}

export const ACCENTS: Record<UIAccent, Accent> = {
  copper: {
    key: "copper",
    name: "Copper",
    hex: "#d97757",
    strong: "#b85a3f",
    soft: "rgb(217 119 87 / 0.14)",
  },
  indigo: {
    key: "indigo",
    name: "Indigo",
    hex: "#818cf8",
    strong: "#6366f1",
    soft: "rgb(129 140 248 / 0.14)",
  },
  emerald: {
    key: "emerald",
    name: "Emerald",
    hex: "#34d399",
    strong: "#10b981",
    soft: "rgb(52 211 153 / 0.14)",
  },
  amber: {
    key: "amber",
    name: "Amber",
    hex: "#fbbf24",
    strong: "#f59e0b",
    soft: "rgb(251 191 36 / 0.14)",
  },
  rose: {
    key: "rose",
    name: "Rose",
    hex: "#fb7185",
    strong: "#f43f5e",
    soft: "rgb(251 113 133 / 0.14)",
  },
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

export function diffBgs(
  style: "soft" | "punchy" | "bar",
  added = "var(--gd-added)",
  removed = "var(--gd-removed)",
): DiffStyleColors {
  if (style === "punchy") {
    return {
      addBg: "rgb(132 204 22 / 0.14)",
      addStrong: "rgb(132 204 22 / 0.32)",
      remBg: "rgb(231 111 81 / 0.14)",
      remStrong: "rgb(231 111 81 / 0.34)",
      addNum: "rgb(132 204 22 / 0.20)",
      remNum: "rgb(231 111 81 / 0.22)",
      addBar: added,
      remBar: removed,
    };
  }
  if (style === "bar") {
    return {
      addBg: "transparent",
      addStrong: "rgb(132 204 22 / 0.18)",
      remBg: "transparent",
      remStrong: "rgb(231 111 81 / 0.20)",
      addNum: "transparent",
      remNum: "transparent",
      addBar: added,
      remBar: removed,
    };
  }
  return {
    addBg: "rgb(132 204 22 / 0.07)",
    addStrong: "rgb(132 204 22 / 0.22)",
    remBg: "rgb(231 111 81 / 0.07)",
    remStrong: "rgb(231 111 81 / 0.24)",
    addNum: "rgb(132 204 22 / 0.14)",
    remNum: "rgb(231 111 81 / 0.14)",
    addBar: "rgb(132 204 22 / 0.6)",
    remBar: "rgb(231 111 81 / 0.6)",
  };
}
