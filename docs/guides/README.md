# Guides

Background reading on *why* the Template is shaped the way it is — distinct from
the step-by-step Skill runbook (a pending task) and from the
[Architecture Decision Records](../adr/), which capture individual decisions.

- **[CLI conventions & the runtime contract](cli-conventions.md)** — streams
  (stdout = data, stderr = chatter), the `0/1/2/>2` exit-code contract, the global
  flag set, configuration precedence (`flags > env > file > defaults`), logging,
  and progress.
- **[Instantiating a new tool](instantiating-a-new-tool.md)** — what the scaffold
  does, the sentinel tokens it substitutes, its flags, and how the worked example
  turns scaffold output into a real tool.
- **[Releasing & distribution](releasing.md)** — the tag-driven GoReleaser
  pipeline: cross-compiled archives, checksums, GitHub Releases, `go install`, a
  Homebrew tap, cosign signatures, and SBOMs; plus one-time maintainer setup and
  local validation.
- **[The test suite](testing.md)** — the three test styles every scaffolded tool
  inherits (table-driven, golden-file with `-update`, testscript E2E), the
  `Run`/`runCmd` seams, and how the exit-code contract is tested on both sides.
- **[CI/CD for this repository](ci-cd.md)** — the multi-module workflow that
  builds and tests the Template, scaffold, and example standalone (`GOWORK=off`),
  the asymmetric matrix, Dependabot, and the shared lint config.

New to the project? Start with the [root README](../../README.md), then
[CONTEXT.md](../../CONTEXT.md) for vocabulary.
