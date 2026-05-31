package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &serverResource{}
	_ resource.ResourceWithConfigure   = &serverResource{}
	_ resource.ResourceWithImportState = &serverResource{}
)

// NewServerResource is a helper function to simplify the provider implementation.
func NewServerResource() resource.Resource {
	return &serverResource{}
}

// serverResource is the resource implementation.
type serverResource struct {
	client *discordgo.Session
}

// serverResourceModel maps the resource schema data.
type serverResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Region          types.String `tfsdk:"region"`
	VerificationLvl types.Int64  `tfsdk:"verification_level"`
	DefaultMsgNotif types.Int64  `tfsdk:"default_message_notifications"`
	ExplicitContent types.Int64  `tfsdk:"explicit_content_filter"`
	AfkChannelID    types.String `tfsdk:"afk_channel_id"`
	AfkTimeout      types.Int64  `tfsdk:"afk_timeout"`
	Icon            types.String `tfsdk:"icon"`
	Splash          types.String `tfsdk:"splash"`
	OwnerID         types.String `tfsdk:"owner_id"`
	LastUpdated     types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

// Schema defines the schema for the resource.
func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Discord server (guild).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the server.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the server (2-100 characters).",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "Voice region ID for the server.",
				Optional:    true,
				Computed:    true,
			},
			"verification_level": schema.Int64Attribute{
				Description: "Verification level required for the server (0-4).",
				Optional:    true,
				Computed:    true,
			},
			"default_message_notifications": schema.Int64Attribute{
				Description: "Default message notification level (0 = ALL_MESSAGES, 1 = ONLY_MENTIONS).",
				Optional:    true,
				Computed:    true,
			},
			"explicit_content_filter": schema.Int64Attribute{
				Description: "Explicit content filter level (0 = DISABLED, 1 = MEMBERS_WITHOUT_ROLES, 2 = ALL_MEMBERS).",
				Optional:    true,
				Computed:    true,
			},
			"afk_channel_id": schema.StringAttribute{
				Description: "ID of the AFK voice channel.",
				Optional:    true,
			},
			"afk_timeout": schema.Int64Attribute{
				Description: "AFK timeout in seconds (60, 300, 900, 1800, 3600).",
				Optional:    true,
				Computed:    true,
			},
			"icon": schema.StringAttribute{
				Description: "Base64 encoded icon image for the server.",
				Optional:    true,
			},
			"splash": schema.StringAttribute{
				Description: "Base64 encoded splash image for the server (VIP only).",
				Optional:    true,
			},
			"owner_id": schema.StringAttribute{
				Description: "ID of the owner of the server. If transferred, this will update.",
				Optional:    true,
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*discordgo.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *discordgo.Session, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// GuildCreate only accepts a name, so we create first then edit
	guild, err := r.client.GuildCreate(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Server",
			"Could not create server, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(guild.ID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Apply additional settings via edit
	params := discordgo.GuildParams{
		Name: plan.Name.ValueString(),
	}

	if !plan.Region.IsNull() && !plan.Region.IsUnknown() {
		params.Region = plan.Region.ValueString()
	}

	if !plan.VerificationLvl.IsNull() && !plan.VerificationLvl.IsUnknown() {
		vl := discordgo.VerificationLevel(plan.VerificationLvl.ValueInt64())
		params.VerificationLevel = &vl
	}

	if !plan.DefaultMsgNotif.IsNull() && !plan.DefaultMsgNotif.IsUnknown() {
		params.DefaultMessageNotifications = int(plan.DefaultMsgNotif.ValueInt64())
	}

	if !plan.ExplicitContent.IsNull() && !plan.ExplicitContent.IsUnknown() {
		params.ExplicitContentFilter = int(plan.ExplicitContent.ValueInt64())
	}

	if !plan.AfkTimeout.IsNull() && !plan.AfkTimeout.IsUnknown() {
		params.AfkTimeout = int(plan.AfkTimeout.ValueInt64())
	}

	if !plan.Icon.IsNull() && !plan.Icon.IsUnknown() {
		params.Icon = plan.Icon.ValueString()
	}

	if !plan.Splash.IsNull() && !plan.Splash.IsUnknown() {
		params.Splash = plan.Splash.ValueString()
	}

	if !plan.OwnerID.IsNull() && !plan.OwnerID.IsUnknown() {
		params.OwnerID = plan.OwnerID.ValueString()
	}

	// Only edit if there are params beyond name
	hasEdit := params.Region != "" || params.VerificationLevel != nil ||
		params.DefaultMessageNotifications != 0 || params.ExplicitContentFilter != 0 ||
		params.AfkTimeout != 0 || params.Icon != "" || params.Splash != "" || params.OwnerID != ""

	if hasEdit {
		guild, err = r.client.GuildEdit(guild.ID, &params)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error Updating Server After Create",
				"Server was created but could not apply all settings: "+err.Error(),
			)
		}
	}

	// Set state to fully populated data
	if plan.Region.IsUnknown() || plan.Region.IsNull() {
		plan.Region = types.StringValue(guild.Region)
	}
	if plan.VerificationLvl.IsUnknown() || plan.VerificationLvl.IsNull() {
		plan.VerificationLvl = types.Int64Value(int64(guild.VerificationLevel))
	}
	if plan.DefaultMsgNotif.IsUnknown() || plan.DefaultMsgNotif.IsNull() {
		plan.DefaultMsgNotif = types.Int64Value(int64(guild.DefaultMessageNotifications))
	}
	if plan.ExplicitContent.IsUnknown() || plan.ExplicitContent.IsNull() {
		plan.ExplicitContent = types.Int64Value(int64(guild.ExplicitContentFilter))
	}
	if plan.AfkTimeout.IsUnknown() || plan.AfkTimeout.IsNull() {
		plan.AfkTimeout = types.Int64Value(int64(guild.AfkTimeout))
	}
	if plan.OwnerID.IsUnknown() || plan.OwnerID.IsNull() {
		plan.OwnerID = types.StringValue(guild.OwnerID)
	}

	tflog.Trace(ctx, "created a server")

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	guild, err := r.client.Guild(state.ID.ValueString(), discordgo.WithRetryOnRatelimit(true))
	if err != nil {
		if err.Error() == "HTTP 404 Not Found" {
			resp.State.RemoveResource(ctx)
			return
		}
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
	state.AfkTimeout = types.Int64Value(int64(guild.AfkTimeout))
	state.OwnerID = types.StringValue(guild.OwnerID)

	if guild.AfkChannelID != "" {
		state.AfkChannelID = types.StringValue(guild.AfkChannelID)
	} else {
		state.AfkChannelID = types.StringNull()
	}

	if guild.Icon != "" {
		state.Icon = types.StringValue(guild.Icon)
	} else {
		state.Icon = types.StringNull()
	}

	if guild.Splash != "" {
		state.Splash = types.StringValue(guild.Splash)
	} else {
		state.Splash = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := discordgo.GuildParams{
		Name: plan.Name.ValueString(),
	}

	if !plan.Region.IsNull() && !plan.Region.IsUnknown() {
		params.Region = plan.Region.ValueString()
	}

	if !plan.VerificationLvl.IsNull() && !plan.VerificationLvl.IsUnknown() {
		vl := discordgo.VerificationLevel(plan.VerificationLvl.ValueInt64())
		params.VerificationLevel = &vl
	}

	if !plan.DefaultMsgNotif.IsNull() && !plan.DefaultMsgNotif.IsUnknown() {
		params.DefaultMessageNotifications = int(plan.DefaultMsgNotif.ValueInt64())
	}

	if !plan.ExplicitContent.IsNull() && !plan.ExplicitContent.IsUnknown() {
		params.ExplicitContentFilter = int(plan.ExplicitContent.ValueInt64())
	}

	if !plan.AfkChannelID.IsNull() && !plan.AfkChannelID.IsUnknown() {
		params.AfkChannelID = plan.AfkChannelID.ValueString()
	}

	if !plan.AfkTimeout.IsNull() && !plan.AfkTimeout.IsUnknown() {
		params.AfkTimeout = int(plan.AfkTimeout.ValueInt64())
	}

	if !plan.Icon.IsNull() && !plan.Icon.IsUnknown() {
		params.Icon = plan.Icon.ValueString()
	}

	if !plan.Splash.IsNull() && !plan.Splash.IsUnknown() {
		params.Splash = plan.Splash.ValueString()
	}

	if !plan.OwnerID.IsNull() && !plan.OwnerID.IsUnknown() {
		params.OwnerID = plan.OwnerID.ValueString()
	}

	_, err := r.client.GuildEdit(plan.ID.ValueString(), &params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Server",
			"Could not update server, unexpected error: "+err.Error(),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.GuildDelete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Server",
			"Could not delete server, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
