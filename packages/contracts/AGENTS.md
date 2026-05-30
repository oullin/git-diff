# Contracts

Shared DTOs and enums consumed by `@git-diff/bridge`, the Electron host, and the renderer.

## Layout

- `src/common/` — enums and unions referenced from more than one domain (`RepositoryMode`, `GitFileStatus`, `DiffSectionKind`, `RepositoryRole`, `ReviewContext`). Domain modules re-export the names they need; downstream code can keep importing from the domain folder it cares about.
- `src/<domain>/` — one folder per domain (`auth`, `repo`, `review`, `branch`, ...). Each folder owns request DTOs, response DTOs, and any domain-specific enums.
- `src/index.ts` — barrel export of every domain module.

## Conventions

### `type` vs `interface`

- Use `type` for unions, discriminated unions, and aliases of primitives or literal sets:
    ```ts
    export type RepositoryMode = "working" | "commit";
    export type ReviewContext = { kind: "working" } | { kind: "commit"; sha: string };
    ```
- Use `interface` for record-like shapes (DTOs):
    ```ts
    export interface AuthUser {
        id: number;
        osUsername: string;
        displayName: string;
    }
    ```
- Do not introduce `type Foo = Bar` aliases purely to rename an existing shape. If a name is wrong, change it everywhere.

### Request and response DTOs

- Co-locate request and response types in the same domain module.
- Request types end in `Request`; response types end in `Response` (or read as a noun when the response is the entity itself, e.g. `AuthUser`).
- Optional fields use `?:`; do not model optionality with `Foo | undefined` in the request shape.

### No back-compat shims

The contracts package never carries deprecated aliases. If a name or shape changes, all consumers update in the same commit.

### Discriminated unions over flag pairs

When a value depends on a mode field (e.g. a `kind`/`sha` pair), prefer a discriminated union over two co-dependent optional fields:

```ts
// Yes
export type ReviewContext = { kind: "working" } | { kind: "commit"; sha: string };

// No
export interface ReviewContext {
    kind: "working" | "commit";
    sha?: string; // only meaningful when kind === "commit"
}
```

Wire-format normalization (converting flat JSON to the union and back) lives at the bridge boundary, not in the contracts module.
