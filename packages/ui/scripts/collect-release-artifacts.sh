#!/usr/bin/env bash
# Collects the Electron Forge DMG/ZIP outputs, canonicalizes them to the names
# the Homebrew cask url expects (git-diff-review-<version>-arm64.{dmg,zip}),
# writes SHASUMS256.txt, and emits sha256/dmg to $GITHUB_OUTPUT.
#
# Run from the repo root. Usage: collect-release-artifacts.sh <version>
set -euo pipefail
shopt -s nullglob globstar

version="${1:?usage: collect-release-artifacts.sh <version>}"

out="packages/ui/out/make"
dmgs=("$out"/**/*.dmg)
zips=("$out"/**/*.zip)
if [ "${#dmgs[@]}" -eq 0 ] || [ "${#zips[@]}" -eq 0 ]; then
	echo "Missing Forge artifacts under $out"
	find "$out" -type f || true
	exit 1
fi

mkdir -p dist-release
dmg="git-diff-review-${version}-arm64.dmg"
zip="git-diff-review-${version}-arm64.zip"
cp "${dmgs[0]}" "dist-release/${dmg}"
cp "${zips[0]}" "dist-release/${zip}"

cd dist-release
shasum -a 256 "$dmg" "$zip" > SHASUMS256.txt
sha256="$(shasum -a 256 "$dmg" | awk '{print $1}')"
{
	echo "sha256=$sha256"
	echo "dmg=$dmg"
} >> "$GITHUB_OUTPUT"
echo "Artifacts:" && ls -la
