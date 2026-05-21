<script setup lang="ts">
import { ref, watch } from "vue";
import {
    Dialog,
    DialogBody,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { Button } from "@components/ui/button";
import { Loader2 } from "lucide-vue-next";

type Props = {
    open: boolean;
    upload?: (file: File) => Promise<string>;
};

const props = defineProps<Props>();
const emit = defineEmits<{
    "update:open": [boolean];
    insert: [{ src: string; altText: string }];
}>();

const fileInput = ref<HTMLInputElement | null>(null);
const altText = ref("");
const url = ref("");
const file = ref<File | null>(null);
const error = ref<string | null>(null);
const busy = ref(false);

watch(
    () => props.open,
    (val) => {
        if (!val) {
            altText.value = "";
            url.value = "";
            file.value = null;
            error.value = null;
            busy.value = false;
        }
    },
);

function onFileChange(event: Event): void {
    const target = event.target as HTMLInputElement;

    file.value = target.files?.[0] ?? null;
}

async function submit(): Promise<void> {
    error.value = null;

    if (file.value && props.upload) {
        busy.value = true;
        try {
            const uploadedUrl = await props.upload(file.value);

            emit("insert", { src: uploadedUrl, altText: altText.value });
            emit("update:open", false);
        } catch (err) {
            error.value = err instanceof Error ? err.message : "Upload failed";
        } finally {
            busy.value = false;
        }

        return;
    }

    if (url.value) {
        emit("insert", { src: url.value, altText: altText.value });
        emit("update:open", false);

        return;
    }

    error.value = "Choose a file or paste a URL.";
}
</script>

<template>
    <Dialog :open="open" @update:open="(val: boolean) => emit('update:open', val)">
        <DialogHeader>
            <DialogTitle>Insert image</DialogTitle>
            <DialogDescription>Upload a file or paste an image URL.</DialogDescription>
        </DialogHeader>
        <DialogBody class="space-y-4">
            <div v-if="upload">
                <label class="mb-1 block text-xs font-medium text-muted-foreground">File</label>
                <input
                    ref="fileInput"
                    type="file"
                    accept="image/*"
                    class="block w-full text-sm"
                    @change="onFileChange"
                />
            </div>
            <div>
                <label class="mb-1 block text-xs font-medium text-muted-foreground">{{
                    upload ? "or paste URL" : "Image URL"
                }}</label>
                <input
                    v-model="url"
                    type="url"
                    placeholder="https://"
                    class="block w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm focus:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
            </div>
            <div>
                <label class="mb-1 block text-xs font-medium text-muted-foreground">Alt text</label>
                <input
                    v-model="altText"
                    type="text"
                    placeholder="Describe the image"
                    class="block w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm focus:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
            </div>
            <p v-if="error" class="text-xs text-destructive">{{ error }}</p>
        </DialogBody>
        <DialogFooter>
            <Button variant="outline" :disabled="busy" @click="emit('update:open', false)"
                >Cancel</Button
            >
            <Button :disabled="busy" @click="submit">
                <Loader2 v-if="busy" class="mr-2 h-4 w-4 animate-spin" />Insert
            </Button>
        </DialogFooter>
    </Dialog>
</template>
