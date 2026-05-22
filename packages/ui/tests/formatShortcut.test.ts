import { describe, expect, test } from "vitest";
import { formatShortcut } from "../src/composables/formatShortcut.js";

describe("formatShortcut", () => {
    test("mac: cmd+shift+p → ⌘⇧P", () => {
        expect(formatShortcut("cmd+shift+p", true)).toBe("⌘⇧P");
    });

    test("non-mac: cmd+shift+p → Cmd+Shift+P", () => {
        expect(formatShortcut("cmd+shift+p", false)).toBe("Cmd+Shift+P");
    });

    test("named keys map to symbols on mac", () => {
        expect(formatShortcut("cmd+enter", true)).toBe("⌘↵");
        expect(formatShortcut("escape", true)).toBe("Esc");
        expect(formatShortcut("arrowdown", true)).toBe("↓");
    });

    test("non-mac: named keys are capitalised", () => {
        expect(formatShortcut("cmd+enter", false)).toBe("Cmd+↵");
        expect(formatShortcut("escape", false)).toBe("Esc");
    });

    test("single letter", () => {
        expect(formatShortcut("j", true)).toBe("J");
        expect(formatShortcut("j", false)).toBe("J");
    });

    test("empty input returns empty string", () => {
        expect(formatShortcut("", true)).toBe("");
        expect(formatShortcut("", false)).toBe("");
    });

    test("backslash stays a backslash", () => {
        expect(formatShortcut("cmd+\\", true)).toBe("⌘\\");
    });
});
