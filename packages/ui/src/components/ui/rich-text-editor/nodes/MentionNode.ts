import {
  $applyNodeReplacement,
  type EditorConfig,
  type LexicalNode,
  type NodeKey,
  type SerializedTextNode,
  type Spread,
  TextNode,
} from "lexical";

export type SerializedMentionNode = Spread<
  {
    mentionId: string;
    mentionLabel: string;
  },
  SerializedTextNode
>;

export class MentionNode extends TextNode {
  __mentionId: string;
  __mentionLabel: string;

  static getType(): string {
    return "mention";
  }

  static clone(node: MentionNode): MentionNode {
    return new MentionNode(node.__mentionId, node.__mentionLabel, node.__text, node.__key);
  }

  constructor(mentionId: string, mentionLabel: string, text?: string, key?: NodeKey) {
    super(text ?? `@${mentionLabel}`, key);
    this.__mentionId = mentionId;
    this.__mentionLabel = mentionLabel;
  }

  createDOM(config: EditorConfig): HTMLElement {
    const dom = super.createDOM(config);
    dom.setAttribute("data-mention-id", this.__mentionId);
    dom.classList.add("rte-mention");
    return dom;
  }

  isTextEntity(): true {
    return true;
  }

  canInsertTextBefore(): boolean {
    return false;
  }

  canInsertTextAfter(): boolean {
    return false;
  }

  static importJSON(serializedNode: SerializedMentionNode): MentionNode {
    const node = $createMentionNode(serializedNode.mentionId, serializedNode.mentionLabel);
    node.setFormat(serializedNode.format);
    node.setDetail(serializedNode.detail);
    node.setMode(serializedNode.mode);
    node.setStyle(serializedNode.style);
    return node;
  }

  exportJSON(): SerializedMentionNode {
    return {
      ...super.exportJSON(),
      mentionId: this.__mentionId,
      mentionLabel: this.__mentionLabel,
      type: "mention",
      version: 1,
    };
  }
}

export function $createMentionNode(mentionId: string, mentionLabel: string): MentionNode {
  const node = new MentionNode(mentionId, mentionLabel);
  node.setMode("token").toggleDirectionless();
  return $applyNodeReplacement(node);
}

export function $isMentionNode(node: LexicalNode | null | undefined): node is MentionNode {
  return node instanceof MentionNode;
}
