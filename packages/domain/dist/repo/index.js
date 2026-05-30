/**
 * Lowercase file extensions (with leading dot) the renderer can display
 * via the inline image-diff component. The backend serves the raw bytes
 * regardless of extension — this list is purely a UI routing decision.
 */
export const IMAGE_EXTENSIONS = [
    ".png",
    ".jpg",
    ".jpeg",
    ".gif",
    ".webp",
    ".svg",
    ".avif",
    ".ico",
    ".bmp",
];
/** True when path's extension is in IMAGE_EXTENSIONS (case-insensitive). */
export function isImagePath(path) {
    const dot = path.lastIndexOf(".");
    if (dot < 0) {
        return false;
    }
    return IMAGE_EXTENSIONS.includes(path.slice(dot).toLowerCase());
}
