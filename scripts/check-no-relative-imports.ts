import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

const roots = process.argv.slice(2);
const sourceFile = /\.(?:ts|tsx|js|jsx|mjs|cjs|vue)$/u;
const relativeImport = /from\s+['"]\.{1,2}\//u;
const dynamicRelativeImport = /import\(\s*['"]\.{1,2}\//u;
const relativeRequire = /require\(\s*['"]\.{1,2}\//u;
const matches: string[] = [];

function walkFiles(dir: string): string[] {
  const files: string[] = [];

  if (!existsSync(dir)) {
    return files;
  }

  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);

    if (entry.isDirectory()) {
      files.push(...walkFiles(path));
      continue;
    }

    files.push(path);
  }

  return files;
}

for (const root of roots) {
  for (const path of walkFiles(root)) {
    if (!sourceFile.test(path)) {
      continue;
    }

    const content = readFileSync(path, "utf8");

    if (
      relativeImport.test(content) ||
      dynamicRelativeImport.test(content) ||
      relativeRequire.test(content)
    ) {
      matches.push(path);
    }
  }
}

if (matches.length > 0) {
  console.error(matches.join("\n"));
  process.exit(1);
}
