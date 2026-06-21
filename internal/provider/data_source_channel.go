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
	_ datasource.DataSource              = &channelDataSource{}
	_ datasource.DataSourceWithConfigure = &channelDataSource{}
)

func NewChannelDataSource() datasource.DataSource {
	return &channelDataSource{}
}

type channelDataSource struct {
	client *discordgo.Session
}

type channelDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ServerID  types.String `tfsdk:"server_id"`
	Name      types.String `tfsdk:"name"`
	Type      types.Int64  `tfsdk:"type"`
	Topic     types.String `tfsdk:"topic"`
	NSFW      types.Bool   `tfsdk:"nsfw"`
	Position  types.Int64  `tfsdk:"position"`
	ParentID  types.String `tfsdk:"parent_id"`
	Bitrate   types.Int64  `tfsdk:"bitrate"`
	UserLimit types.Int64  `tfsdk:"user_limit"`
	RateLimit types.Int64  `tfsdk:"rate_limit_per_user"`
}

func (d *channelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel"
}

func (d *channelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a Discord channel by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the channel.",
				Required:    true,
			},
			"server_id": schema.StringAttribute{
				Description: "ID of the server (guild) the channel belongs to.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the channel.",
				Computed:    true,
			},
			"type": schema.Int64Attribute{
				Description: "Channel type (0 = GUILD_TEXT, 2 = GUILD_VOICE, etc.).",
				Computed:    true,
			},
			"topic": schema.StringAttribute{
				Description: "Channel topic.",
				Computed:    true,
			},
			"nsfw": schema.BoolAttribute{
				Description: "Whether the channel is NSFW.",
				Computed:    true,
			},
			"position": schema.Int64Attribute{
				Description: "Channel position in the left-hand listing.",
				Computed:    true,
			},
			"parent_id": schema.StringAttribute{
				Description: "ID of the parent category.",
				Computed:    true,
			},
			"bitrate": schema.Int64Attribute{
				Description: "Bitrate of voice channels.",
				Computed:    true,
			},
			"user_limit": schema.Int64Attribute{
				Description: "User limit of voice channels (0 = unlimited).",
				Computed:    true,
			},
			"rate_limit_per_user": schema.Int64Attribute{
				Description: "Per-user rate limit in seconds (0-21600).",
				Computed:    true,
			},
		},
	}
}

func (d *channelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *channelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state channelDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ch, err := d.client.Channel(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Channel",
			"Could not read channel ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(ch.Name)
	state.ServerID = types.StringValue(ch.GuildID)
	state.Type = types.Int64Value(int64(ch.Type))
	state.NSFW = types.BoolValue(ch.NSFW)
	state.Position = types.Int64Value(int64(ch.Position))

	if ch.Topic != "" {
		state.Topic = types.StringValue(ch.Topic)
	} else {
		state.Topic = types.StringNull()
	}
	if ch.ParentID != "" {
		state.ParentID = types.StringValue(ch.ParentID)
	} else {
		state.ParentID = types.StringNull()
	}
	if ch.Bitrate > 0 {
		state.Bitrate = types.Int64Value(int64(ch.Bitrate))
	} else {
		state.Bitrate = types.Int64Null()
	}
	if ch.UserLimit > 0 {
		state.UserLimit = types.Int64Value(int64(ch.UserLimit))
	} else {
		state.UserLimit = types.Int64Null()
	}
	if ch.RateLimitPerUser > 0 {
		state.RateLimit = types.Int64Value(int64(ch.RateLimitPerUser))
	} else {
		state.RateLimit = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
