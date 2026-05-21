export type RichTextFeatures = {
    tables?: boolean;
    images?: boolean;
    mentions?: boolean;
    checklist?: boolean;
    slashMenu?: boolean;
    markdownShortcuts?: boolean;
};

export type ResolvedRichTextFeatures = Required<RichTextFeatures>;

export function resolveFeatures(features?: RichTextFeatures): ResolvedRichTextFeatures {
    return {
        tables: features?.tables ?? true,
        images: features?.images ?? true,
        mentions: features?.mentions ?? true,
        checklist: features?.checklist ?? true,
        slashMenu: features?.slashMenu ?? true,
        markdownShortcuts: features?.markdownShortcuts ?? true,
    };
}
