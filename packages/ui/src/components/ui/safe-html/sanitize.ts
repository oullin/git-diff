import DOMPurify from "dompurify";

export type SanitizeProfile = "rich-text" | "minimal";

const MINIMAL_TAGS = ["strong", "em", "u", "s", "code", "a", "br", "span"];
const MINIMAL_ATTRS = ["href", "target", "rel", "class"];

const RICH_TEXT_TAGS = [
  "p",
  "br",
  "strong",
  "em",
  "u",
  "s",
  "code",
  "pre",
  "blockquote",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "ul",
  "ol",
  "li",
  "a",
  "img",
  "table",
  "thead",
  "tbody",
  "tr",
  "th",
  "td",
  "span",
  "hr",
];

const RICH_TEXT_ATTRS = [
  "href",
  "target",
  "rel",
  "src",
  "alt",
  "title",
  "class",
  "data-mention-id",
  "data-language",
  "data-lexical-text",
  "colspan",
  "rowspan",
];

export type SanitizeOptions = {
  profile?: SanitizeProfile;
  allowedTags?: string[];
  allowedAttrs?: string[];
};

let hookRegistered = false;

function ensureLinkHook(): void {
  if (hookRegistered) {
    return;
  }

  hookRegistered = true;

  DOMPurify.addHook("afterSanitizeAttributes", (node) => {
    if (!(node instanceof Element)) {
      return;
    }

    if (node.tagName === "A") {
      node.setAttribute("target", "_blank");
      node.setAttribute("rel", "noopener noreferrer");
    }
  });
}

export function sanitizeHtml(input: string, options: SanitizeOptions = {}): string {
  ensureLinkHook();

  const profile = options.profile ?? "rich-text";
  const tags = options.allowedTags ?? (profile === "minimal" ? MINIMAL_TAGS : RICH_TEXT_TAGS);
  const attrs = options.allowedAttrs ?? (profile === "minimal" ? MINIMAL_ATTRS : RICH_TEXT_ATTRS);

  return DOMPurify.sanitize(input, {
    ALLOWED_TAGS: tags,
    ALLOWED_ATTR: attrs,
    ALLOW_DATA_ATTR: false,
    FORBID_ATTR: ["style"],
  });
}
