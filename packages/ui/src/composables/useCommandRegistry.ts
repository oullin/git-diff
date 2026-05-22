import { ref, computed, type Ref, type ComputedRef, getCurrentInstance, onUnmounted } from "vue";
import type { KeymapAction } from "@git-diff/contracts";

/**
 * One palette-visible action. Pure data — no Vue refs, no DOM. The run
 * handler is invoked by the palette when the user picks the command.
 */
export interface PaletteCommand {
    id: string;
    title: string;
    /** Optional section header for grouping ("Navigation", "Diff", ...). */
    section?: string;
    /** Optional keymap action — used to display the live shortcut. */
    keymapId?: KeymapAction;
    /** Invoked when the command is selected. May be async. */
    run: () => void | Promise<void>;
}

// Module-singleton store. Every component that calls useCommandRegistry()
// reads/writes the same list — letting the palette mounted once in App.vue
// see commands registered anywhere in the tree.
const commands = ref<PaletteCommand[]>([]);

export interface UseCommandRegistry {
    /** All currently registered commands, reactive. */
    commands: Ref<PaletteCommand[]>;
    /** Snapshot for read-only consumers (palette filtering etc.). */
    list: ComputedRef<PaletteCommand[]>;
    /**
     * Register a command. Returns a disposer; auto-disposes on the
     * caller's onUnmounted when invoked from a component setup().
     */
    register(command: PaletteCommand): () => void;
    /** Remove a command by id. No-op when the id isn't registered. */
    deregister(id: string): void;
}

export function useCommandRegistry(): UseCommandRegistry {
    return {
        commands,
        list: computed(() => commands.value),
        register(command) {
            // Replace any existing command with the same id — keeps re-renders
            // (HMR, prop changes) from accumulating duplicates.
            const existing = commands.value.findIndex((c) => c.id === command.id);

            if (existing >= 0) {
                commands.value.splice(existing, 1, command);
            } else {
                commands.value.push(command);
            }

            const dispose = (): void => deregister(command.id);

            // Auto-cleanup when called inside a setup() context.
            if (getCurrentInstance()) {
                onUnmounted(dispose);
            }

            return dispose;
        },
        deregister,
    };
}

function deregister(id: string): void {
    const index = commands.value.findIndex((c) => c.id === id);

    if (index >= 0) {
        commands.value.splice(index, 1);
    }
}
