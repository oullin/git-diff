import { ref, computed, type Ref, type ComputedRef, getCurrentInstance, onUnmounted } from "vue";
import type { KeymapAction } from "@git-diff/domain";

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

// Module-singleton so the palette mounted in App.vue sees commands
// registered anywhere in the tree.
const commands = ref<PaletteCommand[]>([]);

export interface UseCommandRegistry {
    commands: Ref<PaletteCommand[]>;
    list: ComputedRef<PaletteCommand[]>;
    /** register auto-disposes on the caller's onUnmounted when invoked
     *  from a component setup(). */
    register(command: PaletteCommand): () => void;
    deregister(id: string): void;
}

export function useCommandRegistry(): UseCommandRegistry {
    return {
        commands,
        list: computed(() => commands.value),
        register(command) {
            // Replace same-id entries so HMR/prop churn doesn't accumulate dupes.
            const existing = commands.value.findIndex((c) => c.id === command.id);

            if (existing >= 0) {
                commands.value.splice(existing, 1, command);
            } else {
                commands.value.push(command);
            }

            const dispose = (): void => deregister(command.id);

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
