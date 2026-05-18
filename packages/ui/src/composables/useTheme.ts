export function applyTheme(): void {
  // Theme follows the OS via `@media (prefers-color-scheme: dark)` in style.css.
  // No class is applied to the document root.
}

export function initTheme(): void {
  applyTheme();
}
