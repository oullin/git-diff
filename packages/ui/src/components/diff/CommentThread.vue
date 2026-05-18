<script setup lang="ts">
import { ref } from "vue";
import { MoreHorizontal } from "lucide-vue-next";
import { LazyRichTextEditor, type RichTextFeatures } from "@ui/rich-text-editor";
import { SafeHtml } from "@ui/safe-html";
import Kbd from "./Kbd.vue";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@ui/dropdown-menu";
import type { ReviewComment } from "@api";

defineProps<{
  comment: ReviewComment;
  replyFeatures: RichTextFeatures;
}>();

const emit = defineEmits<{ delete: []; reply: [bodyHtml: string] }>();

const draft = ref("");

function submit() {
  const value = draft.value.trim();
  if (!value) return;
  emit("reply", value);
  draft.value = "";
}

function initials(name: string): string {
  return (
    name
      .split(/[\s_-]+/)
      .map((p) => p[0]?.toUpperCase() ?? "")
      .slice(0, 2)
      .join("") || "?"
  );
}

function relativeTime(iso: string): string {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "";
  const diff = Date.now() - then;
  const minutes = Math.round(diff / 60000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.round(hours / 24);
  return `${days}d ago`;
}
</script>

<template>
  <div
    :style="{
      padding: '8px 16px 12px',
      background: 'var(--gd-bg)',
      borderTop: '1px dashed var(--gd-border)',
      borderBottom: '1px dashed var(--gd-border)',
    }"
  >
    <div
      :style="{
        background: 'var(--gd-panel)',
        border: '1px solid var(--gd-border)',
        borderRadius: '10px',
        padding: '12px',
        fontFamily: 'var(--font-sans)',
        maxWidth: '720px',
      }"
    >
      <div class="flex items-center" :style="{ gap: '10px' }">
        <span
          :style="{
            width: '26px',
            height: '26px',
            borderRadius: '50%',
            background: 'linear-gradient(135deg, var(--gd-accent-strong), var(--gd-accent))',
            color: '#fff',
            fontSize: '11px',
            fontWeight: 600,
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
          }"
          >{{ initials(comment.authorLabel) }}</span
        >
        <span :style="{ fontSize: '13.5px', fontWeight: 600, color: 'var(--gd-text)' }">{{
          comment.authorLabel
        }}</span>
        <span :style="{ fontSize: '11.5px', color: 'var(--gd-text-3)' }">{{
          relativeTime(comment.createdAt)
        }}</span>
        <div class="flex-1" />
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <button
              type="button"
              :style="{
                width: '26px',
                height: '26px',
                borderRadius: '6px',
                background: 'transparent',
                border: 0,
                color: 'var(--gd-text-3)',
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
              }"
            >
              <MoreHorizontal :size="14" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem class="text-destructive" @select="emit('delete')">
              Delete comment
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <div
        :style="{
          marginTop: '8px',
          fontSize: '13.5px',
          color: 'var(--gd-text)',
          lineHeight: 1.55,
        }"
      >
        <SafeHtml :html="comment.bodyHtml" />
      </div>
      <div
        :style="{
          marginTop: '12px',
          padding: '10px',
          background: 'var(--gd-bg)',
          border: '1px solid var(--gd-border)',
          borderRadius: '8px',
          display: 'flex',
          flexDirection: 'column',
          gap: '8px',
        }"
      >
        <LazyRichTextEditor
          v-model="draft"
          min-height="3rem"
          placeholder="Write a reply…"
          aria-label="Reply"
          :features="replyFeatures"
        />
        <div class="flex justify-end">
          <button
            type="button"
            :disabled="!draft.trim()"
            :style="{
              height: '26px',
              padding: '0 10px',
              borderRadius: '6px',
              border: 0,
              background: 'var(--gd-accent-strong)',
              color: '#fff',
              fontSize: '12px',
              fontWeight: 600,
              display: 'inline-flex',
              alignItems: 'center',
              gap: '4px',
              cursor: 'pointer',
              opacity: draft.trim() ? 1 : 0.5,
            }"
            @click="submit"
          >
            Reply <Kbd>↵</Kbd>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
