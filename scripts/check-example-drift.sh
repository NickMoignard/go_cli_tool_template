#!/bin/sh
#
# check-example-drift.sh — keep the worked example honest.
#
# examples/yaml-validator/ is a hand-maintained instantiation of template/: the
# scaffold output, then extended with the YAML-validation domain. As template/
# evolves, the example's INFRASTRUCTURE files (config, logging, progress,
# version, the CLI plumbing, the global flag wiring, the shared txtar scripts)
# can silently drift from what a fresh scaffold would now produce.
#
# This script re-runs the scaffold against template/ using the example's exact
# identity parameters, then `diff -r`s the generated tree against the committed
# example — EXCLUDING the files the example deliberately owns as domain (the
# validate command, validator testdata, README, go.mod/go.sum, the dropped
# community/release files, etc.). If anything outside that exclude set differs,
# the example has drifted and the script exits non-zero, naming the files.
#
# Run from anywhere: sh scripts/check-example-drift.sh
#
set -eu

REPO_ROOT=$(cd "$(dirname "$0")/.." && pwd)
EXAMPLE="$REPO_ROOT/examples/yaml-validator"

# --- scaffold identity ------------------------------------------------------
# These reproduce how examples/yaml-validator was instantiated. Module path and
# tool name are load-bearing (they appear in every compared source file via the
# scaffold's sentinel substitution). Author/year/description only land in files
# the example dropped or rewrote (LICENSE/README/.goreleaser.yaml), so they do
# not affect the diff — they are passed for fidelity, not correctness.
MODULE="github.com/NickMoignard/yamlvalidate"
NAME="yamlvalidate"
AUTHOR="Nick Moignard"
YEAR="2025"
DESCRIPTION="validate YAML files against a JSON Schema"

# --- generate a fresh scaffold into a temp dir ------------------------------
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT INT TERM
GEN="$TMP/gen"

# -skip-git/-skip-tidy: we diff SOURCE, not a built project. Skipping `go mod
# tidy` keeps the run hermetic (no network) and avoids touching go.sum, which is
# excluded anyway. The scaffold copies only template/, so the workspace go.work
# is irrelevant to the generated output.
( cd "$REPO_ROOT" && go run ./cmd/scaffold \
    -dest "$GEN" \
    -module "$MODULE" \
    -name "$NAME" \
    -author "$AUTHOR" \
    -year "$YEAR" \
    -description "$DESCRIPTION" \
    -skip-git -skip-tidy >/dev/null )

# --- exclude set ------------------------------------------------------------
# diff -x matches BASENAMES (every basename below is unique in these trees, so
# basename matching is precise here). Three buckets:
EXCLUDES=""

# 1. Inherently varying / not first-class source.
EXCLUDES="$EXCLUDES -x .git -x go.sum"

# 2. Domain files the example legitimately owns or rewrote ------------------
# go.mod: the example adds the jsonschema + yaml dependencies its domain needs.
EXCLUDES="$EXCLUDES -x go.mod"
# README.md: rewritten as a worked-example walkthrough, not the scaffold stub.
EXCLUDES="$EXCLUDES -x README.md"
# The placeholder `check` command/domain the example REPLACED with `validate`
# (present only in the fresh scaffold).
EXCLUDES="$EXCLUDES -x check -x check.go -x check_cmd_test.go -x check_resolve_test.go -x check.txtar"
# The `validate` command/domain the example ADDED (present only in the example).
EXCLUDES="$EXCLUDES -x validate -x validate.go -x validate_cmd_test.go -x validate_resolve_test.go -x validate.txtar"
# cli.go: registers validate instead of check and carries the validate-specific
# root Short description — the command-registration seam is domain by design.
EXCLUDES="$EXCLUDES -x cli.go"
# manpages_test.go: asserts the per-command man page name (yamlvalidate-validate.1
# vs the scaffold's yamlvalidate-check.1) — follows the command name, so domain.
EXCLUDES="$EXCLUDES -x manpages_test.go"
# help.golden: the rendered --help lists validate (not check); column widths also
# shift with the command name length. Driven entirely by the domain command.
EXCLUDES="$EXCLUDES -x help.golden"
# Community/release/CI files the example intentionally OMITS (it ships only the
# Go source needed to be a runnable showcase). Present only in the fresh scaffold.
EXCLUDES="$EXCLUDES -x .github -x .golangci.yml -x .goreleaser.yaml"
EXCLUDES="$EXCLUDES -x CODE_OF_CONDUCT.md -x CONTRIBUTING.md -x LICENSE -x SECURITY.md"

# Note: cli_test.go and execute.go are deliberately NOT excluded — their infra
# comments are kept byte-identical to the scaffold output, so they are compared
# like every other infrastructure file. If you reword such a comment in the
# template, mirror it in the example (or this check will flag the divergence).

# --- diff -------------------------------------------------------------------
set +e
# shellcheck disable=SC2086  # word-splitting of $EXCLUDES into -x flags is intended.
DRIFT=$(diff -r $EXCLUDES "$GEN" "$EXAMPLE" 2>&1)
rc=$?
set -e

if [ "$rc" -eq 0 ]; then
  echo "example-drift: OK — non-domain files match a fresh scaffold of template/."
  exit 0
fi

if [ "$rc" -gt 1 ]; then
  echo "example-drift: ERROR — diff failed to run:" >&2
  echo "$DRIFT" >&2
  exit "$rc"
fi

echo "example-drift: DRIFT DETECTED" >&2
echo >&2
echo "The yaml-validator example's infrastructure files have diverged from what" >&2
echo "the scaffold now produces from template/. Differences (generated vs example):" >&2
echo >&2
echo "$DRIFT" >&2
echo >&2
echo "To fix: reconcile examples/yaml-validator/ with template/ — port the template" >&2
echo "change into the example (or, if a file became domain-specific, add it to the" >&2
echo "exclude set in scripts/check-example-drift.sh with a justifying comment)." >&2
exit 1
