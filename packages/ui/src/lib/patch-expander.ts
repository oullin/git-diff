import { type PatchLine, parseHunkHeader } from "@lib/patch-parser";

/**
 * ExpandedContext is one extra context line fetched on demand from the
 * on-disk file when the user clicks the "show more context" affordance.
 * Both `oldLine` and `newLine` are 1-indexed; renderers use them to bind
 * the line to its row in either side of the split view.
 */
export interface ExpandedContext {
    oldLine: number;
    newLine: number;
    text: string;
}

/**
 * Splice expanded context rows into a parsed patch. Each expansion has
 * an old line number from the on-disk file; rows land before the next
 * hunk header whose oldStart they precede, preserving 1-indexed line
 * order.
 */
export function applyExpansions(
    lines: ReadonlyArray<PatchLine>,
    expansions: ReadonlyArray<ExpandedContext>,
): PatchLine[] {
    if (!expansions.length) {
        return [...lines];
    }

    const sorted = [...expansions].sort((a, b) => a.oldLine - b.oldLine);
    const result: PatchLine[] = [];
    let cursor = 0;

    function emit(prefix: string, exp: ExpandedContext): void {
        result.push({
            id: `${prefix}:${exp.oldLine}`,
            type: "context",
            text: exp.text,
            oldLine: exp.oldLine,
            newLine: exp.newLine,
        });
    }

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i]!;

        if (line.type === "meta" && line.text.startsWith("@@")) {
            const header = parseHunkHeader(line.text);

            if (header) {
                while (cursor < sorted.length && sorted[cursor]!.oldLine < header.oldStart) {
                    emit(`${line.id}:exp`, sorted[cursor]!);
                    cursor++;
                }
            }
        }

        result.push(line);
    }

    while (cursor < sorted.length) {
        emit("exp:trail", sorted[cursor]!);
        cursor++;
    }

    return result;
}
