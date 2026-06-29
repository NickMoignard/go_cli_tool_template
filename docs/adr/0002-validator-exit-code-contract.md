# Exit codes separate "invalid input" from "tool error"

CLI exit codes are a public, scripted contract that is painful to change once users
depend on it. We define: **`0` = success/valid**, **`1` = validation failure** (the tool
ran correctly but the input did not conform), **`2` = usage error** (bad flags, missing
file), and **`>2` reserved for unexpected/internal errors**. The non-obvious part is that
a *validation failure is not a tool error* — it gets its own code distinct from both
success and crashes.

## Consequences

This lets scripts compose correctly, e.g. `tool validate ... && deploy`, and lets CI tell
"the file is wrong" apart from "the tool broke." Core logic in `internal/<domain>` returns
typed results/errors rather than calling `os.Exit`; only the thin `cmd/` entrypoint maps
those to exit codes, so the contract stays in one place and is testable via `testscript`.
