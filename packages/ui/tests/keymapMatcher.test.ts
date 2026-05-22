import { describe, expect, test } from "vitest";
import { matchesBinding, parseBinding } from "../src/composables/keymapMatcher.js";

describe("parseBinding", () => {
    test("single key", () => {
        expect(parseBinding("j")).toEqual({
            cmd: false,
            ctrl: false,
            alt: false,
            shift: false,
            key: "j",
        });
    });

    test("cmd+shift+p", () => {
        expect(parseBinding("cmd+shift+p")).toEqual({
            cmd: true,
            ctrl: false,
            alt: false,
            shift: true,
            key: "p",
        });
    });

    test("order is irrelevant", () => {
        expect(parseBinding("p+shift+cmd")).toEqual(parseBinding("cmd+shift+p"));
    });

    test("meta is an alias for cmd", () => {
        expect(parseBinding("meta+enter")).toEqual(parseBinding("cmd+enter"));
    });

    test("mod resolves to cmd on mac", () => {
        const got = parseBinding("mod+s", true);

        expect(got).toEqual({ cmd: true, ctrl: false, alt: false, shift: false, key: "s" });
    });

    test("mod resolves to ctrl off mac", () => {
        const got = parseBinding("mod+s", false);

        expect(got).toEqual({ cmd: false, ctrl: true, alt: false, shift: false, key: "s" });
    });

    test("empty string returns null", () => {
        expect(parseBinding("")).toBeNull();
    });

    test("only modifiers returns null", () => {
        expect(parseBinding("cmd+shift")).toBeNull();
    });

    test("multiple non-modifier keys returns null", () => {
        expect(parseBinding("a+b")).toBeNull();
    });
});

describe("matchesBinding", () => {
    function evt(init: Partial<KeyboardEvent>): KeyboardEvent {
        return {
            key: init.key ?? "",
            metaKey: init.metaKey ?? false,
            ctrlKey: init.ctrlKey ?? false,
            altKey: init.altKey ?? false,
            shiftKey: init.shiftKey ?? false,
        } as KeyboardEvent;
    }

    test("exact match", () => {
        const binding = parseBinding("cmd+shift+p")!;

        expect(matchesBinding(evt({ key: "p", metaKey: true, shiftKey: true }), binding)).toBe(
            true,
        );
    });

    test("missing modifier fails", () => {
        const binding = parseBinding("cmd+shift+p")!;

        expect(matchesBinding(evt({ key: "p", metaKey: true }), binding)).toBe(false);
    });

    test("extra modifier fails (strict matching)", () => {
        const binding = parseBinding("p")!;

        expect(matchesBinding(evt({ key: "p", metaKey: true }), binding)).toBe(false);
    });

    test("case-insensitive on the key", () => {
        const binding = parseBinding("a")!;

        expect(matchesBinding(evt({ key: "A" }), binding)).toBe(true);
    });

    test("named keys (enter, escape, arrowdown)", () => {
        expect(
            matchesBinding(evt({ key: "Enter", metaKey: true }), parseBinding("cmd+enter")!),
        ).toBe(true);
        expect(matchesBinding(evt({ key: "Escape" }), parseBinding("escape")!)).toBe(true);
    });

    test("slash key", () => {
        expect(matchesBinding(evt({ key: "/" }), parseBinding("/")!)).toBe(true);
    });
});
