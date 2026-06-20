---
name: terraform-provider-resource
description: Add a new resource or data source to the Discord Terraform provider. Use when extending the provider with new Discord API functionality (guilds, channels, roles, emojis, webhooks, invites) or when scaffolding CRUD logic following the terraform-plugin-framework patterns used in this repo.
---

# Adding a Terraform Provider Resource or Data Source

Follow these steps when adding new CRUD functionality to `terraform-provider-discord`.

## File layout

All provider code lives in `internal/provider/`:

- `resource_<name>.go` — resource implementation
- `resource_<name>_test.go` — acceptance test stubs (TF_ACC-gated)
- `data_source_<name>.go` — data source implementation
- `provider.go` — central registration point

Examples follow the Registry layout:

- `examples/resources/<name>/resource.tf`
- `examples/data-sources/<name>/data-source.tf`

## Resource pattern

Each resource implements the terraform-plugin-framework interfaces:

```go
type <name>Resource struct {
    client *discordgo.Session
}

func New<Name>Resource() resource.Resource {
    return &<name>Resource{}
}

func (r *<name>Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_<name>"
}

func (r *<name>Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "server_id": schema.StringAttribute{
                Required:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
            },
            // ...
            "id": schema.StringAttribute{
                Computed: true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
            },
        },
    }
}

func (r *<name>Resource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    r.client = req.ProviderData.(*discordgo.Session)
}

func (r *<name>Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) { ... }
func (r *<name>Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) { ... }
func (r *<name>Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { ... }
func (r *<name>Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { ... }
```

## Data source pattern

Same as resource but only implement `Metadata`, `Schema`, `Configure`, `Read`. All attributes are either `Required` (inputs) or `Computed` (outputs).

## Registration

Add the constructor to the relevant slice in `internal/provider/provider.go`:

```go
func (p *discordProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        // ...existing
        New<Name>Resource,
    }
}
```

For data sources:

```go
func (p *discordProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        // ...existing
        New<Name>DataSource,
    }
}
```

## Examples

Create a focused example in `examples/resources/<name>/resource.tf`. Keep it minimal — one resource with the required attributes set. Reference other resources in the same example where it makes sense (e.g. `server_id = discord_server.community.id`).

## Conventions

- **Sensitive fields**: `Sensitive: true` for tokens, webhook URLs, anything that grants access
- **Computed IDs**: `Computed: true` + `stringplanmodifier.UseStateForUnknown()` for fields populated by the Create response
- **Immutable fields**: `stringplanmodifier.RequiresReplace()` for fields like `server_id` that can't be changed without recreating
- **Required strings**: `Required: true` for IDs and names the user must provide
- **Optional integers**: `Optional: true` + `OptionalInt64Default` only when Discord has a sensible API default; otherwise use `types.Int64Null()` and check `IsNull()` before sending
- **Logging**: use `tflog.Debug(ctx, "...", map[string]any{...})` in Create/Read/Update/Delete with key IDs for traceability

## Acceptance tests

Add a `TestAccResource_<name>` (or `TestAccDataSource_<name>`) in `resource_<name>_test.go`:

```go
func TestAccResource<Name>(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read
            {Config: testAcc<Name>Config_basic(), Check: resource.ComposeAggregateTestCheckFunc(
                resource.TestCheckResourceAttrSet("discord_<name>.test", "id"),
            )},
            // ImportState
            {ResourceName:      "discord_<name>.test",
             ImportState:       true,
             ImportStateVerify: true},
            // Update
            {Config: testAcc<Name>Config_updated()},
            // Delete (empty plan after destroy)
        },
    })
}
```

Use `resource.Test` (framework-style, NOT `resource.TestMain`). Tests are gated by `TF_ACC=1` and require `DISCORD_TOKEN` env var. Use `testAccPreCheck` from `provider_test.go` for shared setup.

## After adding a resource

Run these in order before committing:

```bash
mise run fmt       # go fmt + terraform fmt -recursive examples/
mise run doc       # regenerate tfplugindocs
mise run validate  # lint + test
```

The CI `docs` job will fail if docs drift, so always regenerate after adding examples.