import { parseHunkHeader } from "./patch-parser.js";
// Each expansion lands before the next hunk header whose oldStart it
// precedes, preserving line order.
export function applyExpansions(lines, expansions) {
    if (!expansions.length) {
        return [...lines];
    }
    const sorted = [...expansions].sort((a, b) => a.oldLine - b.oldLine);
    const result = [];
    let cursor = 0;
    function emit(prefix, exp) {
        result.push({
            id: `${prefix}:${exp.oldLine}`,
            type: "context",
            text: exp.text,
            oldLine: exp.oldLine,
            newLine: exp.newLine,
        });
    }
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (line.type === "meta" && line.text.startsWith("@@")) {
            const header = parseHunkHeader(line.text);
            if (header) {
                while (cursor < sorted.length && sorted[cursor].oldLine < header.oldStart) {
                    emit(`${line.id}:exp`, sorted[cursor]);
                    cursor++;
                }
            }
        }
        result.push(line);
    }
    while (cursor < sorted.length) {
        emit("exp:trail", sorted[cursor]);
        cursor++;
    }
    return result;
}
