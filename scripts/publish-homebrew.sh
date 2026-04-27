#!/usr/bin/env bash
set -euo pipefail

VERSION="$1"

if [ -z "${HOMEBREW_TAP_TOKEN:-}" ]; then
	echo "HOMEBREW_TAP_TOKEN is not set, skipping Homebrew formula update"
	exit 0
fi

SHA256_ARM64=$(sha256sum "newsgoat-${VERSION}-darwin-arm64.tar.gz" | awk '{print $1}')
SHA256_AMD64=$(sha256sum "newsgoat-${VERSION}-darwin-amd64.tar.gz" | awk '{print $1}')

sed -e "s/VERSION/${VERSION}/g" \
	-e "s/SHA256_ARM64/${SHA256_ARM64}/g" \
	-e "s/SHA256_AMD64/${SHA256_AMD64}/g" \
	scripts/newsgoat.rb.tmpl >newsgoat.rb

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

git clone "https://x-access-token:${HOMEBREW_TAP_TOKEN}@github.com/jarv/homebrew-newsgoat.git" "$TMPDIR"
cp newsgoat.rb "$TMPDIR/newsgoat.rb"

cd "$TMPDIR"
git add newsgoat.rb
git commit -m "newsgoat ${VERSION}"
git push
