<script setup lang="ts">
import { computed, type HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";
import { sanitizeHtml, type SanitizeProfile } from "@ui/safe-html/sanitize";

type Props = {
  html: string;
  profile?: SanitizeProfile;
  allowedTags?: string[];
  allowedAttrs?: string[];
  as?: keyof HTMLElementTagNameMap;
  class?: HTMLAttributes["class"];
};

const props = withDefaults(defineProps<Props>(), {
  profile: "rich-text",
  as: "div",
});

const sanitized = computed(() =>
  sanitizeHtml(props.html, {
    profile: props.profile,
    allowedTags: props.allowedTags,
    allowedAttrs: props.allowedAttrs,
  }),
);
</script>

<template>
  <component
    :is="as"
    :class="cn('prose prose-sm max-w-none dark:prose-invert', props.class)"
    v-html="sanitized"
  />
</template>
