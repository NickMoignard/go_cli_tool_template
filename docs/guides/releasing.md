# Guide: releasing & distribution

Every tool scaffolded from the Template ships with a release pipeline:
[`.goreleaser.yaml`](../../template/.goreleaser.yaml) plus a
[`release.yml`](../../template/.github/workflows/release.yml) workflow. This guide
explains how it works, what a maintainer must set up once, and how to validate it
without publishing.

## How a release happens

Releases are tag-driven. Push a semver tag and the workflow runs
[GoReleaser](https://goreleaser.com):

```console
$ git tag v0.1.0
$ git push origin v0.1.0
```

GoReleaser then, in one run:

- **cross-compiles** the binary for linux/darwin/windows × amd64/arm64;
- packages **archives** (`.tar.gz`, `.zip` on Windows) and a `checksums.txt`;
- publishes a **GitHub Release** with those assets;
- generates an **SBOM** per archive (via syft) and a keyless **cosign** signature
  over the checksums (via the workflow's OIDC token — no private key to manage);
- builds **Linux packages** (`.deb` and `.rpm`, via nFPM) and attaches them to the
  release;
- updates a **Homebrew tap** with a cask and a **Scoop bucket** with a manifest for
  the new version.

`go install` needs nothing extra — it works the moment the tag exists, because it
builds from source at the tag.

## Version metadata

The build injects version info into `internal/version` with `-ldflags -X`, so a
released binary reports a real version, commit, and date:

```console
$ mytool --version
mytool v0.1.0
```

Plain `go install`/`go build` (no ldflags) still works — `internal/version` falls
back to the VCS stamps Go records in the build info.

## One-time setup for a maintainer

| What | Why |
|------|-----|
| Push the repo to GitHub | the workflow and release live there |
| Create a `homebrew-tap` repo under your account | GoReleaser pushes the cask there |
| Add a `HOMEBREW_TAP_GITHUB_TOKEN` repo secret | a PAT with `contents:write` on the tap (the default `GITHUB_TOKEN` cannot push to another repo) |
| Create a `scoop-bucket` repo under your account | GoReleaser pushes the Scoop manifest there |
| Add a `SCOOP_BUCKET_GITHUB_TOKEN` repo secret | a PAT with `contents:write` on the bucket (same reason as the tap) |

The release owner (`{{ .Env.GITHUB_REPOSITORY_OWNER }}` in the config) is filled
in automatically from the Actions context. Not publishing a tap? Delete the
`homebrew_casks:` block in `.goreleaser.yaml` and the secret reference in the
workflow — everything else still works.

### Linux packages (`.deb` / `.rpm`) and Scoop

Two extra channels ride along on the same tag-driven run:

- **Linux packages** — the `nfpms:` block builds a `.deb` and a `.rpm` from the
  already-compiled binary and attaches them to the GitHub Release. This needs **no
  extra repo, token, or secret** — the only field to fill in is a real `maintainer`
  email in the block (the `.deb` format requires one; the template ships a
  placeholder). The packages are not signed; install with your distro's tooling
  (`dpkg -i` / `rpm -i`) or point a repo at the release assets.
- **Scoop (Windows)** — the `scoops:` block pushes a manifest to a `scoop-bucket`
  repo, exactly mirroring the Homebrew tap: create the repo once and add the
  `SCOOP_BUCKET_GITHUB_TOKEN` PAT secret (rows above). Users then
  `scoop bucket add <owner> https://github.com/<owner>/scoop-bucket` and
  `scoop install REPLACE_TOOL`.

Disabling either is the same one-block delete as the tap: drop `nfpms:` (no secret
to remove) or `scoops:` (and its `SCOOP_BUCKET_GITHUB_TOKEN` reference) from
`.goreleaser.yaml`. The rest of the pipeline is unaffected.

## Validate locally before tagging

You do not need to publish to check the config:

```console
$ goreleaser check                                  # config is valid
$ goreleaser release --snapshot --clean --skip=sign,sbom
```

The snapshot build produces the archives and a cask under `dist/` without
touching GitHub, so you can confirm cross-compilation and packaging work.

## Channels & scope

Shipped: **GitHub Releases**, **`go install`**, **Homebrew tap**, **Scoop bucket**,
**nFPM (`.deb`/`.rpm`) packages**, **cosign signatures**, and **SBOMs**.
Deliberately out of scope for now: container images (dropped). See
[ADR-0004](../adr/0004-cli-framework-cobra-fang.md) for the framework/versioning
decisions the release config builds on.
