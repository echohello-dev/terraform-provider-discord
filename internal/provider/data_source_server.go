package provider

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &serverDataSource{}
	_ datasource.DataSourceWithConfigure = &serverDataSource{}
)

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

type serverDataSource struct {
	client *discordgo.Session
}

type serverDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Region          types.String `tfsdk:"region"`
	VerificationLvl types.Int64  `tfsdk:"verification_level"`
	DefaultMsgNotif types.Int64  `tfsdk:"default_message_notifications"`
	ExplicitContent types.Int64  `tfsdk:"explicit_content_filter"`
	OwnerID         types.String `tfsdk:"owner_id"`
	MemberCount     types.Int64  `tfsdk:"member_count"`
}

func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a Discord server (guild).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the server.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the server.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Voice region ID for the server.",
				Computed:    true,
			},
			"verification_level": schema.Int64Attribute{
				Description: "Verification level required for the server.",
				Computed:    true,
			},
			"default_message_notifications": schema.Int64Attribute{
				Description: "Default message notification level.",
				Computed:    true,
			},
			"explicit_content_filter": schema.Int64Attribute{
				Description: "Explicit content filter level.",
				Computed:    true,
			},
			"owner_id": schema.StringAttribute{
				Description: "ID of the owner of the server.",
				Computed:    true,
			},
			"member_count": schema.Int64Attribute{
				Description: "Approximate number of members in the server.",
				Computed:    true,
			},
		},
	}
}

func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*discordgo.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *discordgo.Session, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state serverDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	guild, err := d.client.Guild(state.ID.ValueString(), discordgo.WithRetryOnRatelimit(true))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Server",
			"Could not read server ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(guild.Name)
	state.Region = types.StringValue(guild.Region)
	state.VerificationLvl = types.Int64Value(int64(guild.VerificationLevel))
	state.DefaultMsgNotif = types.Int64Value(int64(guild.DefaultMessageNotifications))
	state.ExplicitContent = types.Int64Value(int64(guild.ExplicitContentFilter))
	state.OwnerID = types.StringValue(guild.OwnerID)
	state.MemberCount = types.Int64Value(int64(guild.MemberCount))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
