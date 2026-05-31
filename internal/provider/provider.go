package provider

import (
	"context"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &discordProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &discordProvider{
			version: version,
		}
	}
}

// discordProvider is the provider implementation.
type discordProvider struct {
	version string
}

// discordProviderModel maps provider schema data to a Go type.
type discordProviderModel struct {
	Token types.String `tfsdk:"token"`
}

// Metadata returns the provider type name.
func (p *discordProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "discord"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *discordProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Discord. The provider requires a Discord bot token to authenticate.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "Discord bot token. May also be provided via DISCORD_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a Discord API client for data sources and resources.
func (p *discordProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Discord client")

	var config discordProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Discord Token",
			"The provider cannot create the Discord API client as there is an unknown configuration value for the Discord token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DISCORD_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	token := os.Getenv("DISCORD_TOKEN")

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Discord Token",
			"The provider cannot create the Discord API client as there is a missing or empty value for the Discord token. "+
				"Set the token value in the configuration or use the DISCORD_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "discord_token", token)
	tflog.Debug(ctx, "Creating Discord client")

	// Create a new Discord client using the configuration values
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Discord Client",
			"An unexpected error occurred when creating the Discord client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Discord Client Error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Configured Discord client", map[string]any{"success": true})

	// Make the Discord client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = dg
	resp.ResourceData = dg
}

// DataSources defines the data sources implemented in the provider.
func (p *discordProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServerDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *discordProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServerResource,
		NewChannelResource,
		NewRoleResource,
		NewEmojiResource,
		NewWebhookResource,
		NewInviteResource,
	}
}
