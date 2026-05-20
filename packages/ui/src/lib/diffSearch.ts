// In-diff text search across all visible hunk lines for the currently shown
// files. Operates entirely client-side against the rendered DOM, so no patch
// parsing reruns when the query changes.
//
// A match is any contiguous occurrence of `query` (case-insensitive by
// default) inside the text content of a `[data-diff-line-text]` element. We
// rely on DiffBody.vue to tag those elements; everything else is renderer-
// agnostic.

export interface DiffSearchMatch {
  /** The DOM node carrying the line text. */
  node: HTMLElement;
  /** Character offset of the match start inside `node.textContent`. */
  start: number;
  /** Character offset of the match end (exclusive). */
  end: number;
}

export interface DiffSearchResult {
  matches: DiffSearchMatch[];
  /** Number of distinct line nodes that have at least one match. */
  matchedLines: number;
}

export interface DiffSearchOptions {
  caseSensitive?: boolean;
}

export function searchDiff(
  query: string,
  root: ParentNode = document,
  options: DiffSearchOptions = {},
): DiffSearchResult {
  const trimmed = query.trim();

  if (!trimmed) {
    return { matches: [], matchedLines: 0 };
  }

  const needle = options.caseSensitive ? trimmed : trimmed.toLowerCase();
  const nodes = Array.from(root.querySelectorAll<HTMLElement>("[data-diff-line-text]"));
  const matches: DiffSearchMatch[] = [];
  const linesWithMatch = new Set<HTMLElement>();

  for (const node of nodes) {
    const text = node.textContent ?? "";
    const haystack = options.caseSensitive ? text : text.toLowerCase();
    let from = 0;
    while (from <= haystack.length) {
      const found = haystack.indexOf(needle, from);
      if (found < 0) break;
      matches.push({ node, start: found, end: found + needle.length });
      linesWithMatch.add(node);
      from = found + Math.max(1, needle.length);
    }
  }

  return { matches, matchedLines: linesWithMatch.size };
}

const HIGHLIGHT_CLASS = "gd-search-hit";
const ACTIVE_CLASS = "gd-search-hit-active";

/** Wrap every match in a span; returns the wrapper elements in order so the
 * caller can highlight an active match without rerunning the search. */
export function applyHighlights(matches: DiffSearchMatch[]): HTMLSpanElement[] {
  const grouped = new Map<HTMLElement, DiffSearchMatch[]>();
  for (const match of matches) {
    const bucket = grouped.get(match.node) ?? [];
    bucket.push(match);
    grouped.set(match.node, bucket);
  }

  const wrappers: HTMLSpanElement[] = [];

  for (const [node, group] of grouped) {
    group.sort((a, b) => a.start - b.start);
    const original = node.textContent ?? "";
    node.textContent = "";

    let cursor = 0;
    for (const match of group) {
      if (match.start > cursor) {
        node.appendChild(document.createTextNode(original.slice(cursor, match.start)));
      }
      const span = document.createElement("span");
      span.className = HIGHLIGHT_CLASS;
      span.textContent = original.slice(match.start, match.end);
      node.appendChild(span);
      wrappers.push(span);
      cursor = match.end;
    }
    if (cursor < original.length) {
      node.appendChild(document.createTextNode(original.slice(cursor)));
    }
  }

  return wrappers;
}

export function clearHighlights(root: ParentNode = document): void {
  const wrappers = root.querySelectorAll<HTMLSpanElement>(`.${HIGHLIGHT_CLASS}`);
  for (const wrapper of wrappers) {
    const parent = wrapper.parentNode;
    if (!parent) continue;
    parent.replaceChild(document.createTextNode(wrapper.textContent ?? ""), wrapper);
    parent.normalize();
  }
}

export function setActiveHighlight(wrappers: HTMLSpanElement[], index: number): void {
  for (const wrapper of wrappers) {
    wrapper.classList.remove(ACTIVE_CLASS);
  }
  const active = wrappers[index];
  if (active) {
    active.classList.add(ACTIVE_CLASS);
    active.scrollIntoView({ block: "center", behavior: "smooth" });
  }
}
