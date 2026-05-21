import { parsePatchFiles } from "@pierre/diffs";
import type { DiffSection } from "@git-diff/contracts";

export type PatchLineType = "context" | "add" | "del" | "meta";

export interface PatchLine {
    id: string;
    type: PatchLineType;
    text: string;
    oldLine?: number;
    newLine?: number;
}

export type SplitRow =
    | { id: string; kind: "meta"; line: PatchLine }
    | { id: string; kind: "context"; line: PatchLine }
    | { id: string; kind: "pair"; left?: PatchLine; right?: PatchLine };

export function parsePatch(section: DiffSection, hideWhitespace: boolean): PatchLine[] {
    if (section.binary) {
        return [{ id: `${section.id}:binary`, type: "meta", text: "Binary file changed" }];
    }

    const lines = parseViaPierre(section) ?? parseLineByLine(section);

    return hideWhitespace
        ? lines.filter((line) => line.type === "meta" || line.text.trim() !== "")
        : lines;
}

function parseViaPierre(section: DiffSection): PatchLine[] | null {
    let parsed;

    try {
        parsed = parsePatchFiles(section.patch, section.id);
    } catch {
        return null;
    }

    const file = parsed[0]?.files[0];

    if (!file || file.hunks.length === 0) {
        return null;
    }

    const lines: PatchLine[] = [];
    let counter = 0;
    const idFor = (suffix: string) => `${section.id}:p${counter++}:${suffix}`;

    for (const hunk of file.hunks) {
        const header = hunk.hunkSpecs
            ? hunk.hunkSpecs + (hunk.hunkContext ? ` ${hunk.hunkContext}` : "")
            : `@@ -${hunk.deletionStart},${hunk.deletionCount} +${hunk.additionStart},${hunk.additionCount} @@${hunk.hunkContext ? ` ${hunk.hunkContext}` : ""}`;

        lines.push({ id: idFor("hunk"), type: "meta", text: header });

        let oldLine = hunk.deletionStart;
        let newLine = hunk.additionStart;

        for (const content of hunk.hunkContent) {
            if (content.type === "context") {
                for (let i = 0; i < content.lines; i++) {
                    const text =
                        file.additionLines[content.additionLineIndex + i] ??
                        file.deletionLines[content.deletionLineIndex + i] ??
                        "";

                    lines.push({ id: idFor("ctx"), type: "context", text, oldLine, newLine });
                    oldLine++;
                    newLine++;
                }
            } else {
                for (let i = 0; i < content.deletions; i++) {
                    const text = file.deletionLines[content.deletionLineIndex + i] ?? "";

                    lines.push({ id: idFor("del"), type: "del", text, oldLine });
                    oldLine++;
                }

                for (let i = 0; i < content.additions; i++) {
                    const text = file.additionLines[content.additionLineIndex + i] ?? "";

                    lines.push({ id: idFor("add"), type: "add", text, newLine });
                    newLine++;
                }
            }
        }
    }

    return lines;
}

function parseLineByLine(section: DiffSection): PatchLine[] {
    const lines: PatchLine[] = [];
    let oldLine = 0;
    let newLine = 0;

    for (const [index, raw] of section.patch.split("\n").entries()) {
        const hunk = /^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/.exec(raw);

        if (hunk) {
            oldLine = Number(hunk[1]);
            newLine = Number(hunk[2]);
            lines.push({ id: `${section.id}:${index}`, type: "meta", text: raw });
            continue;
        }

        if (
            raw.startsWith("diff --git") ||
            raw.startsWith("index ") ||
            raw.startsWith("---") ||
            raw.startsWith("+++")
        ) {
            lines.push({ id: `${section.id}:${index}`, type: "meta", text: raw });
            continue;
        }

        if (raw.startsWith("+")) {
            lines.push({ id: `${section.id}:${index}`, type: "add", text: raw.slice(1), newLine });
            newLine++;
            continue;
        }

        if (raw.startsWith("-")) {
            lines.push({ id: `${section.id}:${index}`, type: "del", text: raw.slice(1), oldLine });
            oldLine++;
            continue;
        }

        lines.push({
            id: `${section.id}:${index}`,
            type: "context",
            text: raw.startsWith(" ") ? raw.slice(1) : raw,
            oldLine,
            newLine,
        });
        oldLine++;
        newLine++;
    }

    return lines;
}

export function splitPatch(section: DiffSection, hideWhitespace: boolean): SplitRow[] {
    const lines = parsePatch(section, hideWhitespace);
    const rows: SplitRow[] = [];
    let i = 0;

    while (i < lines.length) {
        const line = lines[i]!;

        if (line.type === "meta") {
            rows.push({ id: `${line.id}:split-meta`, kind: "meta", line });
            i++;
            continue;
        }

        if (line.type === "context") {
            rows.push({ id: `${line.id}:split-ctx`, kind: "context", line });
            i++;
            continue;
        }

        const dels: PatchLine[] = [];

        while (i < lines.length && lines[i]!.type === "del") {
            dels.push(lines[i]!);
            i++;
        }

        const adds: PatchLine[] = [];

        while (i < lines.length && lines[i]!.type === "add") {
            adds.push(lines[i]!);
            i++;
        }

        const pairCount = Math.max(dels.length, adds.length);

        for (let j = 0; j < pairCount; j++) {
            const left = dels[j];
            const right = adds[j];
            const id = `${(left ?? right)!.id}:split-pair:${j}`;

            rows.push({ id, kind: "pair", left, right });
        }
    }

    return rows;
}
