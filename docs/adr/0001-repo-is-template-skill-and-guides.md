# This repo bundles the Template, the Skill, and the Guides together

We need a way to scaffold new Go CLI projects from vetted best-practice defaults. We
chose to make this single repo hold three things at once: a compilable, CI-tested
**Template** Go module under `template/`; the agent **Skill** (`SKILL.md`) that
instantiates it; and a set of **Guides** under `docs/`. Co-locating them means the
Skill always instantiates from known-good, continuously-built sources, so the scaffold
output cannot bit-rot the way an LLM-emits-files-inline approach would.

## Considered Options

- **Skill emits files inline** — rejected: generated code drifts and stops compiling.
- **Plain template repo, clone-and-rename, no skill** — rejected: loses the automated,
  input-driven instantiation we want.
- **Separate repos for template vs skill** — rejected: splits the source of truth and
  makes it easy for the Skill to reference a stale Template.

## Consequences

`template/` is its own self-contained module so instantiation copies only it, never the
meta-files (Skill, Guides, ADRs, beads). CI must build/test `template/` (and
`examples/yaml-validator/`) independently from the repo root.
