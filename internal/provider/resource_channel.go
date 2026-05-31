package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &channelResource{}
	_ resource.ResourceWithConfigure   = &channelResource{}
	_ resource.ResourceWithImportState = &channelResource{}
)

func NewChannelResource() resource.Resource {
	return &channelResource{}
}

type channelResource struct {
	client *discordgo.Session
}

type channelResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ServerID    types.String `tfsdk:"server_id"`
	Name        types.String `tfsdk:"name"`
	Type        types.Int64  `tfsdk:"type"`
	Topic       types.String `tfsdk:"topic"`
	NSFW        types.Bool   `tfsdk:"nsfw"`
	Position    types.Int64  `tfsdk:"position"`
	ParentID    types.String `tfsdk:"parent_id"`
	Bitrate     types.Int64  `tfsdk:"bitrate"`
	UserLimit   types.Int64  `tfsdk:"user_limit"`
	RateLimit   types.Int64  `tfsdk:"rate_limit_per_user"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *channelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel"
}

func (r *channelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Discord channel within a server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the channel.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "The ID of the server (guild) the channel belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the channel (1-100 characters).",
				Required:    true,
			},
			"type": schema.Int64Attribute{
				Description: "The type of channel (0 = GUILD_TEXT, 2 = GUILD_VOICE, 4 = GUILD_CATEGORY, etc.).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"topic": schema.StringAttribute{
				Description: "The topic of the channel (0-1024 characters).",
				Optional:    true,
			},
			"nsfw": schema.BoolAttribute{
				Description: "Whether the channel is NSFW.",
				Optional:    true,
				Computed:    true,
			},
			"position": schema.Int64Attribute{
				Description: "The position of the channel in the left-hand listing.",
				Optional:    true,
				Computed:    true,
			},
			"parent_id": schema.StringAttribute{
				Description: "ID of the parent category for the channel.",
				Optional:    true,
			},
			"bitrate": schema.Int64Attribute{
				Description: "The bitrate (in bits) of the voice channel (voice only).",
				Optional:    true,
			},
			"user_limit": schema.Int64Attribute{
				Description: "The user limit of the voice channel (voice only, 0 for unlimited).",
				Optional:    true,
			},
			"rate_limit_per_user": schema.Int64Attribute{
				Description: "Amount of seconds a user has to wait before sending another message (0-21600).",
				Optional:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *channelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*discordgo.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *discordgo.Session, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *channelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan channelResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	chType := discordgo.ChannelTypeGuildText
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		chType = discordgo.ChannelType(plan.Type.ValueInt64())
	}

	data := discordgo.GuildChannelCreateData{
		Name: plan.Name.ValueString(),
		Type: chType,
		NSFW: plan.NSFW.ValueBool(),
	}

	if !plan.Topic.IsNull() && !plan.Topic.IsUnknown() {
		data.Topic = plan.Topic.ValueString()
	}
	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		data.Position = int(plan.Position.ValueInt64())
	}
	if !plan.ParentID.IsNull() && !plan.ParentID.IsUnknown() {
		data.ParentID = plan.ParentID.ValueString()
	}
	if !plan.Bitrate.IsNull() && !plan.Bitrate.IsUnknown() {
		data.Bitrate = int(plan.Bitrate.ValueInt64())
	}
	if !plan.UserLimit.IsNull() && !plan.UserLimit.IsUnknown() {
		data.UserLimit = int(plan.UserLimit.ValueInt64())
	}
	if !plan.RateLimit.IsNull() && !plan.RateLimit.IsUnknown() {
		data.RateLimitPerUser = int(plan.RateLimit.ValueInt64())
	}

	channel, err := r.client.GuildChannelCreateComplex(plan.ServerID.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Channel", err.Error())
		return
	}

	plan.ID = types.StringValue(channel.ID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	if plan.Type.IsUnknown() || plan.Type.IsNull() {
		plan.Type = types.Int64Value(int64(channel.Type))
	}
	if plan.NSFW.IsUnknown() || plan.NSFW.IsNull() {
		plan.NSFW = types.BoolValue(channel.NSFW)
	}

	tflog.Trace(ctx, "created a channel")
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *channelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state channelResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := r.client.Channel(state.ID.ValueString())
	if err != nil {
		if err.Error() == "HTTP 404 Not Found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Channel", err.Error())
		return
	}

	state.Name = types.StringValue(channel.Name)
	state.Type = types.Int64Value(int64(channel.Type))
	state.NSFW = types.BoolValue(channel.NSFW)
	state.Position = types.Int64Value(int64(channel.Position))
	if channel.Topic != "" {
		state.Topic = types.StringValue(channel.Topic)
	} else {
		state.Topic = types.StringNull()
	}
	if channel.ParentID != "" {
		state.ParentID = types.StringValue(channel.ParentID)
	} else {
		state.ParentID = types.StringNull()
	}
	if channel.Bitrate > 0 {
		state.Bitrate = types.Int64Value(int64(channel.Bitrate))
	} else {
		state.Bitrate = types.Int64Null()
	}
	if channel.UserLimit > 0 {
		state.UserLimit = types.Int64Value(int64(channel.UserLimit))
	} else {
		state.UserLimit = types.Int64Null()
	}
	if channel.RateLimitPerUser > 0 {
		state.RateLimit = types.Int64Value(int64(channel.RateLimitPerUser))
	} else {
		state.RateLimit = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *channelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan channelResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := discordgo.ChannelEdit{
		Name: plan.Name.ValueString(),
		NSFW: plan.NSFW.ValueBoolPointer(),
	}

	if !plan.Topic.IsNull() && !plan.Topic.IsUnknown() {
		data.Topic = plan.Topic.ValueString()
	}
	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		pos := int(plan.Position.ValueInt64())
		data.Position = &pos
	}
	if !plan.ParentID.IsNull() && !plan.ParentID.IsUnknown() {
		data.ParentID = plan.ParentID.ValueString()
	}
	if !plan.Bitrate.IsNull() && !plan.Bitrate.IsUnknown() {
		data.Bitrate = int(plan.Bitrate.ValueInt64())
	}
	if !plan.UserLimit.IsNull() && !plan.UserLimit.IsUnknown() {
		data.UserLimit = int(plan.UserLimit.ValueInt64())
	}
	if !plan.RateLimit.IsNull() && !plan.RateLimit.IsUnknown() {
		rl := int(plan.RateLimit.ValueInt64())
		data.RateLimitPerUser = &rl
	}

	_, err := r.client.ChannelEdit(plan.ID.ValueString(), &data)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Channel", err.Error())
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *channelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state channelResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.ChannelDelete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Channel", err.Error())
		return
	}
}

func (r *channelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
