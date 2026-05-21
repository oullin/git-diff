import "@rich-text-editor/style.css";

export { default as RichTextEditor } from "@rich-text-editor/RichTextEditor.vue";
export { default as LazyRichTextEditor } from "@rich-text-editor/LazyRichTextEditor.vue";
export { INSERT_IMAGE_COMMAND } from "@rich-text-editor/plugins/imageCommand";
export { ImageNode, $createImageNode, $isImageNode } from "@rich-text-editor/nodes/ImageNode";
export {
    MentionNode,
    $createMentionNode,
    $isMentionNode,
} from "@rich-text-editor/nodes/MentionNode";
export type { ImageNodeProps, SerializedImageNode } from "@rich-text-editor/nodes/ImageNode";
export type { SerializedMentionNode } from "@rich-text-editor/nodes/MentionNode";
export type { MentionItem } from "@rich-text-editor/plugins/mentions";
export type { RichTextFeatures } from "@rich-text-editor/features";
