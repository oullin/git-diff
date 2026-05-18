export function applyTheme(): void {
  const root = globalThis.document?.documentElement;
  if (!root) {
    return;
  }
  root.classList.remove("light");
  root.classList.add("dark");
  root.dataset.colorMode = "dark";
}

export function initTheme(): void {
  applyTheme();
}
