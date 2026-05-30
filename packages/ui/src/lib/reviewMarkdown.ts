import type { ReviewComment, ReviewDetail } from "@git-diff/domain";

// Minimal HTML→Markdown converter scoped to the subset our rich text editor
// emits. Adding turndown would also work, but the tag set is small and we'd
// rather not pull in a DOM-coupled dep.
//
// Supported tags: p, br, strong/b, em/i, code, pre, ul, ol, li, a, blockquote.
// Anything else collapses to its text content.

export function htmlToMarkdown(html: string): string {
    if (!html) {
        return "";
    }

    const doc = new DOMParser().parseFromString(`<root>${html}</root>`, "text/html");
    const root = doc.querySelector("root");

    if (!root) {
        return stripTags(html);
    }

    return convertNode(root, { listDepth: 0, ordered: false }).trim();
}

interface RenderContext {
    listDepth: number;
    ordered: boolean;
    orderedIndex?: number;
}

function convertNode(node: Node, ctx: RenderContext): string {
    if (node.nodeType === Node.TEXT_NODE) {
        return (node.textContent ?? "").replace(/ /g, " ");
    }

    if (node.nodeType !== Node.ELEMENT_NODE) {
        return "";
    }

    const element = node as Element;
    const tag = element.tagName.toLowerCase();

    switch (tag) {
        case "br":
            return "\n";
        case "p":
            return `${convertChildren(element, ctx).trim()}\n\n`;
        case "strong":
        case "b":
            return `**${convertChildren(element, ctx)}**`;
        case "em":
        case "i":
            return `_${convertChildren(element, ctx)}_`;
        case "code":
            if (element.parentElement?.tagName.toLowerCase() === "pre") {
                return convertChildren(element, ctx);
            }

            return `\`${element.textContent ?? ""}\``;
        case "pre": {
            const inner = element.textContent ?? "";

            return `\n\`\`\`\n${inner.replace(/\n+$/, "")}\n\`\`\`\n\n`;
        }
        case "blockquote":
            return (
                convertChildren(element, ctx)
                    .trim()
                    .split("\n")
                    .map((line) => `> ${line}`)
                    .join("\n") + "\n\n"
            );
        case "ul":
        case "ol": {
            const ordered = tag === "ol";
            let index = 1;
            const items: string[] = [];

            for (const child of Array.from(element.children)) {
                if (child.tagName.toLowerCase() !== "li") {
                    continue;
                }

                const marker = ordered ? `${index}.` : "-";
                const inner = convertChildren(child as Element, {
                    ...ctx,
                    listDepth: ctx.listDepth + 1,
                    ordered,
                }).trim();

                items.push(
                    `${"  ".repeat(ctx.listDepth)}${marker} ${inner.replace(/\n/g, "\n   ")}`,
                );
                index += 1;
            }

            return items.join("\n") + "\n\n";
        }
        case "a": {
            const text = convertChildren(element, ctx);
            const href = element.getAttribute("href") ?? "";

            return href ? `[${text}](${href})` : text;
        }
        case "root":
        case "div":
        case "span":
        default:
            return convertChildren(element, ctx);
    }
}

function convertChildren(element: Element, ctx: RenderContext): string {
    let out = "";

    for (const child of Array.from(element.childNodes)) {
        out += convertNode(child, ctx);
    }

    return out;
}

function stripTags(html: string): string {
    return html.replace(/<[^>]+>/g, "").trim();
}

export function formatReviewAsMarkdown(review: ReviewDetail): string {
    const session = review.review;
    const lines: string[] = [];

    const headerSubject = session.title || `Review of ${session.repoRoot}`;

    lines.push(`# ${headerSubject}`);
    lines.push("");

    const meta: string[] = [];

    meta.push(`Repository: \`${session.repoRoot}\``);
    if (session.contextKind === "commit" && session.contextSha) {
        meta.push(`Commit: \`${session.contextSha}\``);
    } else {
        meta.push(`Branch: \`${session.branch || "(working tree)"}\``);
        if (session.headSha) {
            meta.push(`HEAD: \`${session.headSha}\``);
        }
    }

    meta.push(`Started: ${session.startedAt}`);
    if (session.summary) {
        meta.push("");
        meta.push(stripTags(session.summary));
    }

    lines.push(meta.join("  \n"));
    lines.push("");

    const grouped = new Map<string, ReviewComment[]>();

    for (const comment of review.comments) {
        if (comment.deletedAt) {
            continue;
        }

        const bucket = grouped.get(comment.filePath) ?? [];

        bucket.push(comment);
        grouped.set(comment.filePath, bucket);
    }

    if (grouped.size === 0) {
        lines.push("_No comments yet._");

        return lines.join("\n");
    }

    for (const [filePath, comments] of grouped) {
        lines.push(`## \`${filePath}\``);
        lines.push("");
        comments.sort(
            (a, b) => a.lineNumber - b.lineNumber || a.createdAt.localeCompare(b.createdAt),
        );
        for (const comment of comments) {
            const side =
                comment.side === "del" ? "old" : comment.side === "add" ? "new" : comment.side;
            const author = comment.authorLabel || "reviewer";

            lines.push(
                `- **L${comment.lineNumber}** (${side}) — _${author}, ${comment.createdAt}_`,
            );
            const body = htmlToMarkdown(comment.bodyHtml).trim();

            if (body) {
                for (const bodyLine of body.split("\n")) {
                    lines.push(`  ${bodyLine}`);
                }
            }

            lines.push("");
        }
    }

    return lines.join("\n").trim() + "\n";
}
