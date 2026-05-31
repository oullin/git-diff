import { h } from "vue";

import {
  $applyNodeReplacement,
  type DOMConversionMap,
  type DOMConversionOutput,
  type DOMExportOutput,
  DecoratorNode,
  type EditorConfig,
  type LexicalEditor,
  type LexicalNode,
  type NodeKey,
  type SerializedLexicalNode,
  type Spread,
} from "lexical";

export type SerializedImageNode = Spread<
  {
    src: string;
    altText: string;
    width?: number;
    height?: number;
  },
  SerializedLexicalNode
>;

export class ImageNode extends DecoratorNode<ReturnType<typeof h>> {
  __src: string;
  __altText: string;
  __width: number | undefined;
  __height: number | undefined;

  constructor(src: string, altText: string, width?: number, height?: number, key?: NodeKey) {
    super(key);
    this.__src = src;
    this.__altText = altText;
    this.__width = width;
    this.__height = height;
  }

  static getType(): string {
    return "image";
  }

  static clone(node: ImageNode): ImageNode {
    return new ImageNode(node.__src, node.__altText, node.__width, node.__height, node.__key);
  }

  static importJSON(serialized: SerializedImageNode): ImageNode {
    return $createImageNode({
      src: serialized.src,
      altText: serialized.altText,
      width: serialized.width,
      height: serialized.height,
    });
  }

  exportJSON(): SerializedImageNode {
    return {
      type: "image",
      version: 1,
      src: this.__src,
      altText: this.__altText,
      width: this.__width,
      height: this.__height,
    };
  }

  static importDOM(): DOMConversionMap | null {
    return {
      img: () => ({
        conversion: convertImageElement,
        priority: 0,
      }),
    };
  }

  exportDOM(): DOMExportOutput {
    const element = document.createElement("img");

    element.setAttribute("src", this.__src);
    element.setAttribute("alt", this.__altText);
    if (this.__width) {
      element.setAttribute("width", String(this.__width));
    }

    if (this.__height) {
      element.setAttribute("height", String(this.__height));
    }

    return { element };
  }

  createDOM(config: EditorConfig): HTMLElement {
    const span = document.createElement("span");
    const className = config.theme.image;

    if (className !== undefined) {
      span.className = className;
    }

    return span;
  }

  updateDOM(): false {
    return false;
  }

  getSrc(): string {
    return this.__src;
  }

  getAltText(): string {
    return this.__altText;
  }

  decorate(_editor: LexicalEditor): ReturnType<typeof h> {
    return h("img", {
      src: this.__src,
      alt: this.__altText,
      class: "rte-image-img",
      draggable: "false",
      ...(this.__width ? { width: this.__width } : {}),
      ...(this.__height ? { height: this.__height } : {}),
    });
  }
}

function convertImageElement(domNode: Node): DOMConversionOutput {
  if (!(domNode instanceof HTMLImageElement)) {
    return { node: null };
  }

  const node = $createImageNode({
    src: domNode.src,
    altText: domNode.alt,
    width: domNode.width || undefined,
    height: domNode.height || undefined,
  });

  return { node };
}

export type ImageNodeProps = {
  src: string;
  altText: string;
  width?: number;
  height?: number;
};

export function $createImageNode({ src, altText, width, height }: ImageNodeProps): ImageNode {
  return $applyNodeReplacement(new ImageNode(src, altText, width, height));
}

export function $isImageNode(node: LexicalNode | null | undefined): node is ImageNode {
  return node instanceof ImageNode;
}
