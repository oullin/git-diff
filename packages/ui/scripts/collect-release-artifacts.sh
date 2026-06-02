#!/usr/bin/env bash
# Collects the Electron Forge DMG/ZIP outputs, canonicalizes them to the names
# the Homebrew cask url expects (git-diff-review-<version>-arm64.{dmg,zip}),
# writes SHASUMS256.txt, and emits sha256/dmg to $GITHUB_OUTPUT.
#
# Run from the repo root. Usage: collect-release-artifacts.sh <version>
set -euo pipefail

version="${1:?usage: collect-release-artifacts.sh <version>}"

# macOS GitHub runners ship bash 3.2, which lacks `globstar`. Use `find` instead
# of `**` so artifact discovery works regardless of the bash version.
out="packages/ui/out/make"
src_dmg="$(set +o pipefail; find "$out" -type f -name '*.dmg' | head -n1)"
src_zip="$(set +o pipefail; find "$out" -type f -name '*.zip' | head -n1)"
if [ -z "$src_dmg" ] || [ -z "$src_zip" ]; then
	echo "Missing Forge artifacts under $out"
	find "$out" -type f || true
	exit 1
fi

mkdir -p dist-release
dmg="git-diff-review-${version}-arm64.dmg"
zip="git-diff-review-${version}-arm64.zip"
cp "$src_dmg" "dist-release/${dmg}"
cp "$src_zip" "dist-release/${zip}"

cd dist-release
shasum -a 256 "$dmg" "$zip" > SHASUMS256.txt
sha256="$(shasum -a 256 "$dmg" | awk '{print $1}')"
{
	echo "sha256=$sha256"
	echo "dmg=$dmg"
} >> "$GITHUB_OUTPUT"
echo "Artifacts:" && ls -la
