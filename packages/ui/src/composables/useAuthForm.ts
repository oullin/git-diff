import { ref } from "vue";

export function useAuthForm() {
    const submitting = ref(false);
    const error = ref("");

    async function submit<T>(action: () => Promise<T>): Promise<T | null> {
        error.value = "";
        submitting.value = true;

        try {
            return await action();
        } catch (cause) {
            error.value = cause instanceof Error ? cause.message : String(cause);

            return null;
        } finally {
            submitting.value = false;
        }
    }

    function fail(message: string): void {
        error.value = message;
    }

    function clearError(): void {
        error.value = "";
    }

    return {
        submitting,
        error,
        submit,
        fail,
        clearError,
    };
}
