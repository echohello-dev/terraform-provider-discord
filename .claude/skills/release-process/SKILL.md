---
name: release-process
description: Cut a release of the Discord Terraform provider. Use when preparing to ship a new version to the Terraform Registry, when working with release-please PRs, or when troubleshooting GPG signing or goreleaser build failures in the release pipeline.
---

# Cutting a Release

The Discord provider uses **release-please + goreleaser** to publish signed binaries to the Terraform Registry. No manual tagging.

## How it works

1. Conventional commits land on `main` (`feat:`, `fix:`, `perf:`, `refactor:`, `docs:`, `chore:`, `ci:`)
2. On every push to `main`, the `release-please` job in `.github/workflows/release.yml` opens or updates a PR titled `chore(main): release X.Y.Z`
3. When that PR is merged, release-please creates the tag `vX.Y.Z` and the GitHub Release
4. The `goreleaser` job (triggered by `v*` tag push) builds binaries for `linux/darwin/windows × amd64/arm64`, generates SHASUMS, signs them with GPG, and attaches everything to the existing release
5. Terraform Registry ingests from the GitHub Release automatically (registry watches `echohello-dev/discord`)

## Prerequisites

Repo secrets (required for the goreleaser job):

- `GPG_PRIVATE_KEY` — ASCII-armored secret key matching the fingerprint hardcoded in `.github/workflows/release.yml` (`2B11E3055D7BAED4`)
- `GPG_PASSPHRASE` — passphrase for that key (if any)

Verify locally before tagging:

```bash
# Validate config
goreleaser check

# Dry run (requires --snapshot to skip publishing)
goreleaser release --clean --snapshot
```

## Conventional commit types → version bumps

release-please groups commits by type and bumps accordingly:

| Commits since last release | Version bump |
|---|---|
| `feat:` present | minor (X.Y → X.(Y+1).0) |
| only `fix:` / `perf:` | patch (X.Y.Z → X.Y.(Z+1)) |
| `feat:` with `BREAKING CHANGE:` footer | major |

`chore:`, `ci:`, `docs:`, `refactor:` do not trigger a bump unless accompanied by a `feat:` or `fix:`.

## Local release sanity checks

Before pushing to trigger release-please:

```bash
mise run validate    # lint + test must pass
mise run fmt         # code + terraform examples formatted
mise run doc         # docs regenerated and committed
goreleaser check     # .goreleaser.yaml validates
```

## Troubleshooting

**release-please didn't open a PR**

- Check commits since last release have valid conventional prefixes
- Check the `release-please` job ran on the push to `main` (Actions tab)
- Check that `CHANGELOG.md` is committed (release-please edits it; if the bot can't commit, the workflow fails silently)

**goreleaser job failed with "wrong signing key"**

- The `.github/workflows/release.yml` hardcodes a fingerprint check. If `GPG_PRIVATE_KEY` was rotated, update the `EXPECTED_KEYID` in `.github/scripts/import-gpg.sh` and the matching string in the workflow
- Re-run the failed job after fixing

**Registry didn't pick up the release**

- The Registry polls GitHub Releases for `echohello-dev/*` repos; usually takes a few minutes
- Verify the release has `.zip` archives named `terraform-provider-discord_<version>_<os>_<arch>.zip` (goreleaser config in `.goreleaser.yaml`)
- Verify `SHA256SUMS` and `SHA256SUMS.sig` are attached — Registry requires GPG-signed checksums

## Local install vs Registry install

For local development, `mise run install` copies the built binary to `~/.terraform.d/plugins/registry.terraform.io/echohello-dev/discord/<version>/<os>_<arch>/`. For Terraform to pick it up, set the version to match the directory.

For Registry consumers, they declare `version = "~> 0.1"` in their `required_providers` block and `terraform init` pulls the binary automatically.