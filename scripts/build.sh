#!/usr/bin/env bash
# Builds the site into public/. CI and the Dockerfile run it.
#
# The Hugo version lives in .hugo-version only. When the `hugo` on PATH is another version (or missing), the script
# downloads that release from GitHub into .cache/ and checks it against the release's checksum file.
#
# The base URL: BASE_URL if set, otherwise https://shipward.dev/. Extra arguments go to hugo.
set -euo pipefail

cd "$(dirname "$0")/.."

version=$(tr -d '[:space:]' < .hugo-version)
production_url=https://shipward.dev/

hugo_bin=hugo
if ! command -v hugo > /dev/null 2>&1 || ! hugo version | grep -Eq "^hugo v${version//./\\.}([^0-9]|$)"; then
  case "$(uname -m)" in
    x86_64 | amd64) arch=amd64 ;;
    aarch64 | arm64) arch=arm64 ;;
    *) echo "build.sh: no Hugo download for $(uname -m)" >&2; exit 1 ;;
  esac
  dir=.cache/hugo-$version
  if [[ ! -x $dir/hugo ]]; then
    echo "build.sh: downloading Hugo $version ($arch)"
    tarball=hugo_${version}_linux-${arch}.tar.gz
    base=https://github.com/gohugoio/hugo/releases/download/v$version
    mkdir -p "$dir"
    curl -fsSL --retry 3 -o "$dir/$tarball" "$base/$tarball"
    curl -fsSL --retry 3 -o "$dir/checksums.txt" "$base/hugo_${version}_checksums.txt"
    (cd "$dir" && grep " $tarball\$" checksums.txt | sha256sum -c --quiet -)
    tar -xzf "$dir/$tarball" -C "$dir" hugo
    rm -f "$dir/$tarball"
  fi
  hugo_bin=$dir/hugo
fi

if [[ -n ${BASE_URL:-} ]]; then
  base_url=$BASE_URL
else
  base_url=$production_url
fi

"$hugo_bin" version
echo "build.sh: base URL $base_url"
exec "$hugo_bin" --minify --gc --cleanDestinationDir --baseURL "$base_url" "$@"
