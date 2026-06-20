# terraform-provider-discord

Terraform provider for [Discord](https://discord.com) — manage servers (guilds), channels, roles, emojis, webhooks, and invites via a Discord bot token.

## Quick Reference

| | |
|---|---|
| **Registry** | `echohello-dev/discord` |
| **SDK** | `terraform-plugin-framework` (v1.19+) |
| **Discord client** | `github.com/bwmarrin/discordgo` |
| **Go** | 1.25 |
| **Tools** | mise, goreleaser, golangci-lint, tfplugindocs |
| **Release** | release-please + goreleaser + GPG signing |

## Architecture

```
main.go                              # providerserver entrypoint, sets Address
internal/provider/
  provider.go                        # schema + resource/data source registration
  provider_test.go                   # provider schema validation tests
  resource_<name>.go                 # CRUD for each resource
  resource_<name>_test.go            # acceptance test stubs (TF_ACC-gated)
  data_source_<name>.go              # Read-only lookups
  data_source_<name>_test.go
  utils.go                           # small shared helpers
.github/workflows/
  ci.yml                             # build, vet, test, lint, docs-check via mise
  release.yml                        # release-please + goreleaser
.github/scripts/
  check-docs.sh                      # assert docs are up to date
  import-gpg.sh                      # import + verify GPG signing key
examples/
  provider/provider.tf               # provider block usage
  resources/<name>/resource.tf       # Registry per-resource examples
  data-sources/<name>/data-source.tf # Registry per-data-source examples
docs/                                # committed tfplugindocs output (registry-ingestable)
  guides/                            # hand-written guides (NOT regenerated)
mise.toml                            # tool versions + tasks
.goreleaser.yaml                     # build matrix, archives, checksums, GPG signing
```

## Tooling

All tools managed by [mise](https://mise.jdx.dev). Install once, then:

```bash
mise install          # install all tools from mise.toml
mise run build        # compile provider binary to .build/
mise run vet          # go vet ./...
mise run lint         # golangci-lint run ./...
mise run test         # unit tests (skips TF_ACC when unset)
mise run testacc      # acceptance tests (requires DISCORD_TOKEN)
mise run fmt          # go fmt + terraform fmt
mise run doc          # regenerate tfplugindocs
mise run validate     # lint + test
mise run install      # build + copy to local Terraform plugins dir
mise run dev-setup    # write ~/.terraformrc dev_overrides
mise run tidy         # go mod tidy
mise run release      # goreleaser release --clean (CI only)
```

## Adding a Resource

1. Create `internal/provider/resource_<name>.go`
2. Implement `Metadata`, `Schema`, `Configure`, `Create`, `Read`, `Update`, `Delete`, and (where applicable) `ImportState` on the resource type
3. Register in `internal/provider/provider.go` `Resources()` slice
4. Add `internal/provider/resource_<name>_test.go` with `TestAccResource_<name>` (TF_ACC-gated)
5. Add `examples/resources/<name>/resource.tf`
6. Run `mise run fmt` (formats both Go and Terraform) and `mise run doc`
7. Run `mise run validate` before committing

## Adding a Data Source

1. Create `internal/provider/data_source_<name>.go` implementing `Metadata`, `Schema`, `Configure`, `Read`
2. Register in `internal/provider/provider.go` `DataSources()` slice
3. Add `examples/data-sources/<name>/data-source.tf`
4. Run `mise run fmt` and `mise run doc`
5. Run `mise run validate` before committing

## Provider Conventions

- All resource/data source logic lives in `internal/provider/`
- Framework types: prefer `types.String`, `types.Int64`, `types.Bool` from `terraform-plugin-framework/types`
- Sensitive: `Sensitive: true` for tokens, webhooks, IDs that grant access
- Computed: `Computed: true` and `PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}` for fields populated post-Create
- ForceNew: `RequiresReplace()` plan modifier for immutable fields (e.g., server_id)
- Auth: bot token read from `token` attribute or `DISCORD_TOKEN` env var in `provider.go` `Configure`
- Logging: use `tflog` with context for all diagnostic output

## Testing

- **Unit tests**: always run on PR; cover schema validation, plan modifiers, helpers
- **Acceptance tests** (`*_test.go` with `TestAccResource_*`): TF_ACC-gated, require real `DISCORD_TOKEN`
  - Test names begin with `TestAcc` and call `resource.Test(t, resource.TestCase{...})`
- **Provider tests**: `internal/provider/provider_test.go` covers provider schema validation

## Docs

- `mise run doc` regenerates everything in `docs/` **except** `docs/guides/` (hand-written)
- New resources/data sources automatically pick up `examples/resources/<name>/resource.tf` content
- Hand-written guides in `docs/guides/` are preserved; tfplugindocs will NOT delete them (only deletes directories it owns)
- CI runs `.github/scripts/check-docs.sh` to fail builds when docs drift

## Release Process

1. Push conventional commits to `main`
2. release-please opens a "chore(main): release X.Y.Z" PR
3. Merge the release PR
4. release-please creates tag + GitHub Release
5. goreleaser builds binaries for linux/darwin/windows (amd64+arm64), generates SHASUMS, signs with GPG
6. Terraform Registry ingests from GitHub Release automatically

Required repo secrets for releases:

- `GPG_PRIVATE_KEY` — ASCII-armored secret key (fingerprinted in `.github/scripts/import-gpg.sh`)
- `GPG_PASSPHRASE` — key passphrase (if any)

## Commit Convention

Conventional Commits, enforced by release-please changelog grouping:

| Type | Section |
|---|---|
| `feat:` | Features |
| `fix:` | Bug Fixes |
| `perf:` | Performance |
| `refactor:` | Refactoring |
| `docs:` | Documentation |
| `chore:` / `ci:` | Miscellaneous |

Always include `Co-authored-by: opencode-agent <noreply@opencode.ai>` when an AI agent assists.
