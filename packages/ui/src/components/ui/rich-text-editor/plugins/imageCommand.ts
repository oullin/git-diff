import { createCommand, type LexicalCommand } from 'lexical';
import type { ImageNodeProps } from '@rich-text-editor/nodes/ImageNode';

export const INSERT_IMAGE_COMMAND: LexicalCommand<ImageNodeProps> = createCommand('INSERT_IMAGE_COMMAND');
