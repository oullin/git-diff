import { readdirSync } from "node:fs";
import { join } from "node:path";
import { failWithMatches } from "#scripts/lib/cli.js";
import { walkFiles } from "#scripts/lib/walk.js";

const roots = ["src", "tests", "electron", "scripts", "../../scripts", "../bridge/src"];
// Package/workspace roots whose top-level files must be TS (config files like
// forge.config, vite.config, etc.). Scanned non-recursively so node_modules and
// build output are not traversed.
const flatRoots = [".", "../.."];
const forbidden = /\.(?:js|jsx|mjs|cjs|tsx)$/u;
const matches: string[] = [];

for (const root of roots) {
    for (const path of walkFiles(root)) {
        if (forbidden.test(path)) {
            matches.push(path);
        }
    }
}

for (const root of flatRoots) {
    for (const entry of readdirSync(root, { withFileTypes: true })) {
        if (entry.isFile() && forbidden.test(entry.name)) {
            matches.push(join(root, entry.name));
        }
    }
}

failWithMatches(matches);
