<script setup lang="ts">
import { ChevronDown, LogOut } from "lucide-vue-next";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@ui/dropdown-menu";
import type { AuthUser } from "@git-diff/domain";

defineProps<{
    currentUser: AuthUser | null;
    userInitials: string;
}>();

const emit = defineEmits<{
    "log-out": [];
}>();
</script>

<template>
    <DropdownMenu>
        <DropdownMenuTrigger as-child>
            <button
                type="button"
                class="inline-flex items-center"
                :style="{
                    gap: '8px',
                    height: '32px',
                    padding: '3px 10px 3px 3px',
                    borderRadius: '999px',
                    background: 'var(--gd-panel-2)',
                    border: '1px solid var(--gd-border)',
                    color: 'var(--gd-text)',
                    fontSize: '13.5px',
                    fontWeight: 500,
                    whiteSpace: 'nowrap',
                    flexShrink: 0,
                    cursor: 'pointer',
                }"
            >
                <span
                    :style="{
                        width: '24px',
                        height: '24px',
                        borderRadius: '50%',
                        background:
                            'linear-gradient(135deg, var(--gd-accent-strong), var(--gd-accent))',
                        color: '#fff',
                        fontSize: '11px',
                        fontWeight: 600,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                    }"
                    >{{ userInitials }}</span
                >
                <span>{{ currentUser?.osUsername || "guest" }}</span>
                <ChevronDown :size="12" :style="{ color: 'var(--gd-text-3)' }" />
            </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-[200px]">
            <DropdownMenuItem disabled>
                {{ currentUser?.displayName || currentUser?.osUsername || "Signed out" }}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem @select="emit('log-out')">
                <LogOut class="h-4 w-4" />
                Log out
            </DropdownMenuItem>
        </DropdownMenuContent>
    </DropdownMenu>
</template>
