<script setup lang="ts">
import { X } from "lucide-vue-next";
import { LazyRichTextEditor, type RichTextFeatures } from "@ui/rich-text-editor";
import { SafeHtml } from "@ui/safe-html";
import type { ReviewDetail, ReviewSession } from "@api";

defineProps<{
  reviews: ReviewSession[];
  activeReview: ReviewDetail | null;
  summaryDraft: string;
  commentDraftPath?: string | undefined;
  commentDraftLine?: number | undefined;
  commentDraft: string;
  features: RichTextFeatures;
}>();

defineEmits<{
  close: [];
  "update:summaryDraft": [value: string];
  "update:commentDraft": [value: string];
  "select-review": [review: ReviewSession];
  "start-review": [];
  "save-comment": [];
  "cancel-comment": [];
}>();

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
</script>

<template>
  <div
    class="flex flex-col h-full"
    :style="{
      background: 'var(--gd-bg)',
      borderLeft: '1px solid var(--gd-border)',
      minWidth: 0,
    }"
  >
    <div
      class="flex items-center justify-between"
      :style="{
        padding: '12px 16px',
        borderBottom: '1px solid var(--gd-border)',
      }"
    >
      <div :style="{ fontSize: '13.5px', fontWeight: 600, color: 'var(--gd-text)' }">
        Conversation
      </div>
      <button
        type="button"
        title="Close"
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
        @click="$emit('close')"
      >
        <X :size="14" />
      </button>
    </div>
    <div class="min-h-0 flex-1 overflow-auto">
      <div :style="{ padding: '16px', borderBottom: '1px solid var(--gd-border)' }">
        <div :style="{ fontSize: '13px', fontWeight: 600, color: 'var(--gd-text)' }">
          Review notes
        </div>
        <p :style="{ marginTop: '4px', fontSize: '12px', color: 'var(--gd-text-3)' }">
          Local-only review timeline and rich comments.
        </p>
        <LazyRichTextEditor
          class="mt-3"
          :model-value="summaryDraft"
          min-height="8rem"
          placeholder="Write a review summary…"
          aria-label="Review summary"
          :features="features"
          @update:model-value="(value: string) => $emit('update:summaryDraft', value)"
        />
        <button
          type="button"
          class="mt-3 w-full"
          :style="{
            height: '32px',
            borderRadius: '8px',
            border: 0,
            background: 'var(--gd-accent-strong)',
            color: '#fff',
            fontSize: '13px',
            fontWeight: 600,
            cursor: 'pointer',
          }"
          @click="$emit('start-review')"
        >
          Start review
        </button>
      </div>

      <div
        v-if="commentDraftPath !== undefined"
        :style="{ padding: '16px', borderBottom: '1px solid var(--gd-border)' }"
      >
        <div :style="{ fontSize: '13px', fontWeight: 600, color: 'var(--gd-text)' }">
          Comment on {{ commentDraftPath }}:{{ commentDraftLine }}
        </div>
        <LazyRichTextEditor
          class="mt-3"
          :model-value="commentDraft"
          min-height="8rem"
          placeholder="Leave a comment…"
          aria-label="Line comment"
          autofocus
          :features="features"
          @update:model-value="(value: string) => $emit('update:commentDraft', value)"
        />
        <div :style="{ marginTop: '12px', display: 'flex', gap: '8px' }">
          <button
            type="button"
            :style="{
              height: '30px',
              padding: '0 12px',
              borderRadius: '8px',
              border: '1px solid var(--gd-border)',
              background: 'var(--gd-accent-strong)',
              color: '#fff',
              fontSize: '12.5px',
              fontWeight: 600,
              cursor: 'pointer',
            }"
            @click="$emit('save-comment')"
          >
            Save comment
          </button>
          <button
            type="button"
            :style="{
              height: '30px',
              padding: '0 12px',
              borderRadius: '8px',
              border: '1px solid var(--gd-border)',
              background: 'var(--gd-panel-2)',
              color: 'var(--gd-text-2)',
              fontSize: '12.5px',
              fontWeight: 500,
              cursor: 'pointer',
            }"
            @click="$emit('cancel-comment')"
          >
            Cancel
          </button>
        </div>
      </div>

      <div :style="{ padding: '16px' }">
        <div
          :style="{
            marginBottom: '8px',
            fontSize: '13px',
            fontWeight: 600,
            color: 'var(--gd-text)',
          }"
        >
          Activity
        </div>
        <div v-if="reviews.length === 0" :style="{ fontSize: '12.5px', color: 'var(--gd-text-3)' }">
          No reviews yet.
        </div>
        <button
          v-for="review in reviews"
          :key="review.id"
          type="button"
          class="mb-2 w-full text-left"
          :style="{
            border: '1px solid var(--gd-border)',
            borderRadius: '8px',
            padding: '10px 12px',
            background: 'var(--gd-panel-2)',
            color: 'var(--gd-text)',
            fontSize: '12.5px',
            cursor: 'pointer',
          }"
          @click="$emit('select-review', review)"
        >
          <div :style="{ fontWeight: 500 }">{{ review.title }}</div>
          <div :style="{ fontSize: '11px', color: 'var(--gd-text-3)', marginTop: '2px' }">
            {{ formatDate(review.startedAt) }}
          </div>
        </button>

        <div
          v-if="activeReview"
          class="mt-4"
          :style="{ display: 'flex', flexDirection: 'column', gap: '10px' }"
        >
          <div
            :style="{
              border: '1px solid var(--gd-border)',
              borderRadius: '8px',
              padding: '10px 12px',
              background: 'var(--gd-panel)',
            }"
          >
            <div :style="{ fontSize: '13px', fontWeight: 600, color: 'var(--gd-text)' }">
              {{ activeReview.review.title }}
            </div>
            <div
              v-if="activeReview.review.summary"
              :style="{ marginTop: '6px', fontSize: '12.5px', color: 'var(--gd-text-2)' }"
            >
              <SafeHtml :html="activeReview.review.summary" />
            </div>
          </div>
          <div
            v-for="event in activeReview.events"
            :key="event.id"
            :style="{
              borderLeft: '2px solid var(--gd-border)',
              paddingLeft: '10px',
              fontSize: '12.5px',
              color: 'var(--gd-text)',
            }"
          >
            <div :style="{ fontWeight: 500 }">{{ event.message || event.type }}</div>
            <div :style="{ fontSize: '11px', color: 'var(--gd-text-3)' }">
              {{ formatDate(event.createdAt) }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
