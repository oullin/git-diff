// Pure key-matching utility for binding strings like "cmd+shift+p".

const MODIFIERS = new Set(["cmd", "ctrl", "alt", "shift", "meta", "mod"]);

export interface ParsedBinding {
    cmd: boolean;
    ctrl: boolean;
    alt: boolean;
    shift: boolean;
    /**
     * The non-modifier key. Normalised to lowercase. Single-char keys
     * (a-z, 0-9, `/`) match `event.key`; multi-char keys (`enter`,
     * `escape`, `arrowup`...) match the lower-cased `event.key` too.
     */
    key: string;
}

/** Returns null for malformed input. `mod` is cmd on macOS, ctrl elsewhere. */
export function parseBinding(
    binding: string,
    isMac: boolean = isMacPlatform(),
): ParsedBinding | null {
    if (!binding) {
        return null;
    }

    const parts = binding
        .toLowerCase()
        .split("+")
        .map((p) => p.trim())
        .filter(Boolean);

    if (parts.length === 0) {
        return null;
    }

    let cmd = false;
    let ctrl = false;
    let alt = false;
    let shift = false;
    let key: string | null = null;

    for (const part of parts) {
        if (MODIFIERS.has(part)) {
            switch (part) {
                case "cmd":
                case "meta":
                    cmd = true;
                    break;
                case "ctrl":
                    ctrl = true;
                    break;
                case "alt":
                    alt = true;
                    break;
                case "shift":
                    shift = true;
                    break;
                case "mod":
                    if (isMac) {
                        cmd = true;
                    } else {
                        ctrl = true;
                    }

                    break;
            }

            continue;
        }

        if (key !== null) {
            return null;
        }

        key = part;
    }

    if (!key) {
        return null;
    }

    return { cmd, ctrl, alt, shift, key };
}

/** Modifier requirements are strict — a binding without `shift` won't
 *  match a Shift-held keystroke. */
export function matchesBinding(event: KeyboardEvent, binding: ParsedBinding): boolean {
    if (binding.cmd !== event.metaKey) {
        return false;
    }

    if (binding.ctrl !== event.ctrlKey) {
        return false;
    }

    if (binding.alt !== event.altKey) {
        return false;
    }

    if (binding.shift !== event.shiftKey) {
        return false;
    }

    return normaliseKey(event.key) === binding.key;
}

function normaliseKey(key: string): string {
    return key.toLowerCase();
}

function isMacPlatform(): boolean {
    if (typeof navigator === "undefined") {
        return false;
    }

    return /mac|iphone|ipad|ipod/i.test(navigator.platform || "");
}
