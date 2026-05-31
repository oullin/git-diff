import { ref, type Ref } from "vue";

// save() optimistically merges locally then accepts the backend's response
// as authoritative (it normalises empty values into deletions).
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
