import { computed, shallowRef, type ComputedRef } from "vue";

export type ThemeChoice = "light" | "dark" | "system";

const THEME_CHOICES: readonly ThemeChoice[] = ["light", "dark", "system"];

const choice = shallowRef<ThemeChoice>("system");

const systemPrefersDark = shallowRef(false);

export const resolvedTheme: ComputedRef<"light" | "dark"> = computed(() =>
  choice.value === "system" ? (systemPrefersDark.value ? "dark" : "light") : choice.value,
);

export function isThemeChoice(value: unknown): value is ThemeChoice {
  return typeof value === "string" && (THEME_CHOICES as readonly string[]).includes(value);
}

export function setTheme(next: ThemeChoice): void {
  if (choice.value === next) {
    apply();

    return;
  }

  choice.value = next;
  apply();
}

export function getTheme(): ThemeChoice {
  return choice.value;
}

function apply(): void {
  if (typeof document === "undefined") {
    return;
  }

  const root = document.documentElement;

  root.classList.toggle("dark", choice.value === "dark");
  root.classList.toggle("theme-system", choice.value === "system");
  root.toggleAttribute("data-prefers-dark", systemPrefersDark.value);
}

export function initTheme(initial: ThemeChoice = "system"): void {
  choice.value = initial;
  if (typeof window !== "undefined" && window.matchMedia) {
    const mql = window.matchMedia("(prefers-color-scheme: dark)");

    systemPrefersDark.value = mql.matches;
    mql.addEventListener("change", (event) => {
      systemPrefersDark.value = event.matches;
      apply();
    });
  }

  apply();
}

export function applyTheme(): void {
  apply();
}
