import { reactive } from "vue";

// Shared registry that lets viewport-deferred diff files be force-rendered on
// demand — e.g. when in-diff search or hunk navigation needs content that has
// not yet scrolled into view. Module-level singleton so SearchBar and the
// navigation composable can reach it without prop drilling.
interface LazyRenderState {
  renderAll: boolean;
  forced: Set<string>;
}

const state = reactive<LazyRenderState>({
  renderAll: false,
  forced: new Set<string>(),
});

export function forceRenderAllDiffFiles(): void {
  state.renderAll = true;
}

export function forceRenderDiffFile(path: string): void {
  state.forced.add(path);
}

export function isDiffFileForced(path: string): boolean {
  return state.renderAll || state.forced.has(path);
}

// Called when the diff context changes (new repo, branch, or commit) so a
// previous "render all" does not permanently defeat lazy loading.
export function resetLazyRender(): void {
  state.renderAll = false;
  state.forced.clear();
}
