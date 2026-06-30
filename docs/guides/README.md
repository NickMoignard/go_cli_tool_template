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

New to the project? Start with the [root README](../../README.md), then
[CONTEXT.md](../../CONTEXT.md) for vocabulary.

> Guides for distribution/releases and CI/CD will be added once those parts of the
> Template are built (tracked as issues — run `bd ready`).
