// Pretty-print a keymap binding string for display in the command palette.
// Pure function — keeps the rendering logic out of the .vue file so it can
// be unit-tested independently.

const MAC_SYMBOL: Record<string, string> = {
    cmd: "⌘",
    meta: "⌘",
    ctrl: "⌃",
    alt: "⌥",
    shift: "⇧",
};

const KEY_LABEL: Record<string, string> = {
    enter: "↵",
    escape: "Esc",
    arrowup: "↑",
    arrowdown: "↓",
    arrowleft: "←",
    arrowright: "→",
    backspace: "⌫",
    delete: "Del",
    tab: "Tab",
    space: "␣",
    "\\": "\\",
};

/**
 * Turn "cmd+shift+p" into "⌘⇧P" on macOS or "Ctrl+Shift+P" elsewhere.
 * Returns the empty string for falsy input — callers can render
 * conditionally without guarding.
 */
export function formatShortcut(binding: string, isMac: boolean = isMacPlatform()): string {
    if (!binding) {
        return "";
    }

    const parts = binding
        .toLowerCase()
        .split("+")
        .map((p) => p.trim())
        .filter(Boolean);

    if (isMac) {
        return parts.map((p) => MAC_SYMBOL[p] ?? formatKey(p, true)).join("");
    }

    return parts.map((p) => (MAC_SYMBOL[p] ? capitalise(p) : formatKey(p, false))).join("+");
}

function formatKey(part: string, mac: boolean): string {
    const label = KEY_LABEL[part];

    if (label) {
        return label;
    }

    if (part.length === 1) {
        return mac ? part.toUpperCase() : part.toUpperCase();
    }

    return capitalise(part);
}

function capitalise(s: string): string {
    return s.length === 0 ? s : s.charAt(0).toUpperCase() + s.slice(1);
}

function isMacPlatform(): boolean {
    if (typeof navigator === "undefined") {
        return false;
    }

    return /mac|iphone|ipad|ipod/i.test(navigator.platform || "");
}
