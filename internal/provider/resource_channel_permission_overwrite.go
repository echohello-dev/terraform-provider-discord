package provider

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &channelPermissionOverwriteResource{}
	_ resource.ResourceWithConfigure   = &channelPermissionOverwriteResource{}
	_ resource.ResourceWithImportState = &channelPermissionOverwriteResource{}
)

func NewChannelPermissionOverwriteResource() resource.Resource {
	return &channelPermissionOverwriteResource{}
}

type channelPermissionOverwriteResource struct {
	client *discordgo.Session
}

type channelPermissionOverwriteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ChannelID   types.String `tfsdk:"channel_id"`
	TargetID    types.String `tfsdk:"target_id"`
	TargetType  types.String `tfsdk:"target_type"`
	Allow       types.String `tfsdk:"allow"`
	Deny        types.String `tfsdk:"deny"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *channelPermissionOverwriteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel_permission_overwrite"
}

func (r *channelPermissionOverwriteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a permission overwrite on a Discord channel for a role or member.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID of the overwrite (channel_id:target_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"channel_id": schema.StringAttribute{
				Description: "The ID of the channel.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_id": schema.StringAttribute{
				Description: "ID of the role or member the overwrite applies to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_type": schema.StringAttribute{
				Description: "Type of target: 'role' or 'member'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allow": schema.StringAttribute{
				Description: "Bitwise permission value to allow.",
				Required:    true,
			},
			"deny": schema.StringAttribute{
				Description: "Bitwise permission value to deny.",
				Required:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *channelPermissionOverwriteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *channelPermissionOverwriteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan channelPermissionOverwriteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := discordgo.PermissionOverwriteTypeRole
	if plan.TargetType.ValueString() == "member" {
		targetType = discordgo.PermissionOverwriteTypeMember
	}

	allow, err := parsePermissions(plan.Allow.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid allow bitfield", err.Error())
		return
	}
	deny, err := parsePermissions(plan.Deny.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid deny bitfield", err.Error())
		return
	}

	if err := r.client.ChannelPermissionSet(
		plan.ChannelID.ValueString(),
		plan.TargetID.ValueString(),
		targetType,
		allow,
		deny,
	); err != nil {
		resp.Diagnostics.AddError("Error Creating Permission Overwrite", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.ChannelID.ValueString(), plan.TargetID.ValueString()))
	plan.LastUpdated = types.StringValue(nowRFC3339())

	tflog.Trace(ctx, "created channel permission overwrite")
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *channelPermissionOverwriteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state channelPermissionOverwriteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := r.client.Channel(state.ChannelID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var found *discordgo.PermissionOverwrite
	for _, ow := range channel.PermissionOverwrites {
		if ow.ID == state.TargetID.ValueString() {
			found = ow
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Allow = types.StringValue(formatPermissions(found.Allow))
	state.Deny = types.StringValue(formatPermissions(found.Deny))
	state.TargetType = types.StringValue(overwriteTypeString(found.Type))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *channelPermissionOverwriteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan channelPermissionOverwriteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := discordgo.PermissionOverwriteTypeRole
	if plan.TargetType.ValueString() == "member" {
		targetType = discordgo.PermissionOverwriteTypeMember
	}

	allow, err := parsePermissions(plan.Allow.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid allow bitfield", err.Error())
		return
	}
	deny, err := parsePermissions(plan.Deny.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid deny bitfield", err.Error())
		return
	}

	if err := r.client.ChannelPermissionSet(
		plan.ChannelID.ValueString(),
		plan.TargetID.ValueString(),
		targetType,
		allow,
		deny,
	); err != nil {
		resp.Diagnostics.AddError("Error Updating Permission Overwrite", err.Error())
		return
	}

	plan.LastUpdated = types.StringValue(nowRFC3339())
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *channelPermissionOverwriteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state channelPermissionOverwriteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ChannelPermissionDelete(
		state.ChannelID.ValueString(),
		state.TargetID.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error Deleting Permission Overwrite", err.Error())
		return
	}
}

func (r *channelPermissionOverwriteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func overwriteTypeString(t discordgo.PermissionOverwriteType) string {
	if t == discordgo.PermissionOverwriteTypeMember {
		return "member"
	}
	return "role"
}
