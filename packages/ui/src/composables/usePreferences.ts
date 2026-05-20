import { ref, type Ref } from "vue";

// usePreferences mirrors the renderer-side UI preference map. save() does an
// optimistic local merge then trusts the backend's response as the source of
// truth (it normalises empty values into deletions). load() initialises the
// store from the backend; reset() drops cached values on logout.
export interface UsePreferences {
  values: Ref<Record<string, string>>;
  load: () => Promise<void>;
  save: (patch: Record<string, string>) => Promise<void>;
  reset: () => void;
}

export interface UsePreferencesOptions {
  onSaveError?: (cause: unknown) => void;
}

export function usePreferences(opts: UsePreferencesOptions = {}): UsePreferences {
  const values = ref<Record<string, string>>({});

  async function load(): Promise<void> {
    const response = await window.diffApp.getUIPreferences();
    values.value = response.values;
  }

  async function save(patch: Record<string, string>): Promise<void> {
    const next = { ...values.value };

    for (const [key, value] of Object.entries(patch)) {
      if (value === "") {
        delete next[key];
      } else {
        next[key] = value;
      }
    }

    values.value = next;

    try {
      const response = await window.diffApp.saveUIPreferences(patch);
      values.value = response.values;
    } catch (cause) {
      opts.onSaveError?.(cause);
    }
  }

  function reset(): void {
    values.value = {};
  }

  return { values, load, save, reset };
}
