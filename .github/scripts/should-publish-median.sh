#!/usr/bin/env bash
# Decide whether @b4moss/median (packages/js) should be published from the current HEAD.
# Usage: should-publish-median.sh [gate|compare|all]
#   gate    — tag / ancestry checks only (no dist, no npm pack). Run before npm ci.
#   compare — content diff vs npm (requires packages/js/dist). Run after build.
#   all     — gate then compare (default; for local use).
#
# Outputs GitHub Actions-style keys to GITHUB_OUTPUT when set:
#   skip=true|false
#   tag=vX.Y.Z (when gate passes)
#
# JS-only: root tags vX.Y.Z tied to packages/js/package.json.
# Go tags (packages/go/v*) must never drive this script.
set -euo pipefail

MODE="${1:-all}"
ROOT="$(git rev-parse --show-toplevel)"
PKG_DIR="$ROOT/packages/js"

emit() {
  local key="$1"
  local value="$2"
  if [[ "${GITHUB_OUTPUT:-}" ]]; then
    echo "${key}=${value}" >>"$GITHUB_OUTPUT"
  else
    echo "${key}=${value}"
  fi
}

skip() {
  local reason="$1"
  echo "$reason"
  emit "skip" "true"
  exit 0
}

run_gate() {
  local pkg_ver tag tag_commit head_commit
  pkg_ver="$(node -p "require('${PKG_DIR}/package.json').version")"
  tag="v${pkg_ver}"

  # Root Node tags only (reject nested module tags if misused as package version).
  if [[ "$tag" == */* ]] || [[ ! "$tag" =~ ^v[0-9] ]]; then
    skip "Refusing non-JS tag form ${tag}; npm publish uses root vX.Y.Z only."
  fi

  if ! git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
    skip "No git tag ${tag} for packages/js version ${pkg_ver}; skip npm publish."
  fi

  # Accept a v* tag on HEAD or on an ancestor (merge commits onto release
  # usually do not carry the tag themselves).
  tag_commit="$(git rev-list -n 1 "${tag}")"
  head_commit="$(git rev-parse HEAD)"
  if [[ "$tag_commit" != "$head_commit" ]] &&
    ! git merge-base --is-ancestor "$tag_commit" "$head_commit"; then
    skip "Tag ${tag} (${tag_commit}) is not an ancestor of HEAD; skip npm publish."
  fi

  echo "Using tag ${tag} at ${tag_commit} (HEAD=${head_commit})."
  emit "tag" "$tag"
}

run_compare() {
  local pkg_ver published local_tgz
  # Keep compare staging dir in a global so the EXIT trap can see it under `set -u`.
  pkg_ver="$(node -p "require('${PKG_DIR}/package.json').version")"
  published="$(npm view @b4moss/median version 2>/dev/null || true)"

  if [[ -z "$published" ]]; then
    echo "@b4moss/median is not on npm yet; will publish ${pkg_ver}."
    emit "skip" "false"
    exit 0
  fi

  if [[ ! -d "$PKG_DIR/dist" ]]; then
    echo "packages/js/dist is missing; build before comparing to npm."
    exit 1
  fi

  SHARDIAN_COMPARE_TMP="$(mktemp -d)"
  cleanup_compare_tmp() { rm -rf "${SHARDIAN_COMPARE_TMP:-}"; }
  trap cleanup_compare_tmp EXIT

  mkdir -p "$SHARDIAN_COMPARE_TMP/pub" "$SHARDIAN_COMPARE_TMP/local"

  (
    cd "$SHARDIAN_COMPARE_TMP"
    npm pack "@b4moss/median@${published}" --silent >/dev/null
    tar -xzf "b4moss-shardian-${published}.tgz" -C "$SHARDIAN_COMPARE_TMP/pub"
  )

  local_tgz="$(
    cd "$PKG_DIR"
    npm pack --silent --pack-destination "$SHARDIAN_COMPARE_TMP"
  )"
  tar -xzf "$SHARDIAN_COMPARE_TMP/$local_tgz" -C "$SHARDIAN_COMPARE_TMP/local"

  node <<EOF
const fs = require("node:fs");
const path = require("node:path");

function normalize(pkgDir) {
  const file = path.join(pkgDir, "package.json");
  const json = JSON.parse(fs.readFileSync(file, "utf8"));
  json.version = "0.0.0";
  delete json.gitHead;
  fs.writeFileSync(file, JSON.stringify(json, null, 2) + "\n");
}

normalize("$SHARDIAN_COMPARE_TMP/pub/package");
normalize("$SHARDIAN_COMPARE_TMP/local/package");
EOF

  if diff -rq "$SHARDIAN_COMPARE_TMP/pub/package" "$SHARDIAN_COMPARE_TMP/local/package" >/dev/null; then
    skip "No @b4moss/median package content change vs npm@${published} (version-normalized); skip publish."
  fi

  echo "Package content differs from npm@${published}; will publish ${pkg_ver}."
  emit "skip" "false"
}

case "$MODE" in
  gate)
    run_gate
    emit "skip" "false"
    ;;
  compare)
    run_compare
    ;;
  all)
    run_gate
    run_compare
    ;;
  *)
    echo "Unknown mode: ${MODE} (expected gate|compare|all)" >&2
    exit 1
    ;;
esac
