import type { ChangedFile } from "@git-diff/contracts";

import { parsePatch } from "@lib/patch";

/**
 * Walks every section's patch in `file` and returns the set of new-file line
 * numbers that were added by the change. Used by the markdown preview to
 * highlight which rendered lines correspond to additions in the diff.
 *
 * For fully-new files (`added` / `untracked`) the set is empty — every line
 * in the rendered output is implicitly an addition, so highlighting them all
 * would be noise.
 */
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

/**
 * Checks whether a path is a markdown file (`.md` / `.markdown`, case
 * insensitive). Used to decide whether the preview toggle should appear in
 * the file header.
 */
export function isMarkdownPath(path: string): boolean {
    return /\.(md|markdown)$/i.test(path);
}

/**
 * Splits markdown source into per-line segments and renders each on its own
 * `<div data-line="N">`. Returns sanitized HTML safe to inject via v-html.
 *
 * This is intentionally less faithful than a full block-aware renderer — we
 * want the rendered output anchored to source lines so each `data-line` can
 * be tinted when it appears in `addedLines`. Block elements that span
 * multiple source lines (lists, fenced code blocks) render once on the line
 * where they begin and absorb their trailing lines as empty placeholders so
 * the line-number column still lines up.
 */
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
