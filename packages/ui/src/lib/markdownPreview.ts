import type { ChangedFile } from "@git-diff/domain";

import { parsePatch } from "@git-diff/domain/diff";

// Returns an empty set for fully-new files since highlighting every
// line would be noise.
export function getAddedLineNumbers(file: ChangedFile): ReadonlySet<number> {
    if (file.status === "added" || file.status === "untracked") {
        return EMPTY;
    }

    const added = new Set<number>();

    for (const section of file.sections) {
        for (const line of parsePatch(section, false)) {
            if (line.type === "add" && line.newLine != null) {
                added.add(line.newLine);
            }
        }
    }

    return added;
}

export function isMarkdownPath(path: string): boolean {
    return /\.(md|markdown)$/i.test(path);
}

// Per-line rendering trades full block fidelity for line anchoring so
// `data-line` attributes line up with `addedLines` tinting. Multi-line
// blocks render once on the first line; trailing lines become empty
// placeholders so the gutter stays aligned.
export function renderMarkdownWithLineAnchors(
    source: string,
    addedLines: ReadonlySet<number>,
    renderInline: (markdown: string) => string,
    sanitize: (html: string) => string,
): string {
    const sourceLines = source.split("\n");
    const out: string[] = [];

    for (let i = 0; i < sourceLines.length; i++) {
        const lineNumber = i + 1;
        const lineSource = sourceLines[i] ?? "";
        const rendered = lineSource.trim() === "" ? "" : sanitize(renderInline(lineSource));
        const addedClass = addedLines.has(lineNumber) ? " md-line-added" : "";

        out.push(
            `<div class="md-line${addedClass}" data-line="${lineNumber}">${rendered || "&nbsp;"}</div>`,
        );
    }

    return out.join("");
}

const EMPTY: ReadonlySet<number> = new Set<number>();
