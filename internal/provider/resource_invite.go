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
	_ resource.Resource                = &inviteResource{}
	_ resource.ResourceWithConfigure   = &inviteResource{}
	_ resource.ResourceWithImportState = &inviteResource{}
)

func NewInviteResource() resource.Resource {
	return &inviteResource{}
}

type inviteResource struct {
	client *discordgo.Session
}

type inviteResourceModel struct {
	Code       types.String `tfsdk:"code"`
	ChannelID  types.String `tfsdk:"channel_id"`
	ServerID   types.String `tfsdk:"server_id"`
	MaxAge     types.Int64  `tfsdk:"max_age"`
	MaxUses    types.Int64  `tfsdk:"max_uses"`
	Temporary  types.Bool   `tfsdk:"temporary"`
	Unique     types.Bool   `tfsdk:"unique"`
	TargetType types.Int64  `tfsdk:"target_type"`
}

func (r *inviteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

func (r *inviteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Discord invite.",
		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				Description: "The unique invite code.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"channel_id": schema.StringAttribute{
				Description: "The ID of the channel the invite is for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "The ID of the server the invite is for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_age": schema.Int64Attribute{
				Description: "Duration of invite in seconds (0 for never).",
				Optional:    true,
				Computed:    true,
			},
			"max_uses": schema.Int64Attribute{
				Description: "Maximum number of uses (0 for unlimited).",
				Optional:    true,
				Computed:    true,
			},
			"temporary": schema.BoolAttribute{
				Description: "Whether the invite grants temporary membership.",
				Optional:    true,
				Computed:    true,
			},
			"unique": schema.BoolAttribute{
				Description: "Whether the invite is unique (don't reuse similar invites).",
				Optional:    true,
				Computed:    true,
			},
			"target_type": schema.Int64Attribute{
				Description: "Invite target type (0 = STREAM, 1 = EMBEDDED_APPLICATION, 2 = ROLE_SUBSCRIPTION_TEMPLATE).",
				Optional:    true,
			},
		},
	}
}

func (r *inviteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*discordgo.Session)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *discordgo.Session, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *inviteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan inviteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	maxAge := 0
	if !plan.MaxAge.IsNull() && !plan.MaxAge.IsUnknown() {
		maxAge = int(plan.MaxAge.ValueInt64())
	}

	maxUses := 0
	if !plan.MaxUses.IsNull() && !plan.MaxUses.IsUnknown() {
		maxUses = int(plan.MaxUses.ValueInt64())
	}

	temporary := false
	if !plan.Temporary.IsNull() && !plan.Temporary.IsUnknown() {
		temporary = plan.Temporary.ValueBool()
	}

	unique := true
	if !plan.Unique.IsNull() && !plan.Unique.IsUnknown() {
		unique = plan.Unique.ValueBool()
	}

	invite, err := r.client.ChannelInviteCreate(plan.ChannelID.ValueString(), discordgo.Invite{
		MaxAge:    maxAge,
		MaxUses:   maxUses,
		Temporary: temporary,
		Unique:    unique,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Invite", err.Error())
		return
	}

	plan.Code = types.StringValue(invite.Code)
	tflog.Trace(ctx, "created an invite")

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *inviteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state inviteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	invite, err := r.client.Invite(state.Code.ValueString())
	if err != nil {
		if err.Error() == "HTTP 404 Not Found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Invite", err.Error())
		return
	}

	state.ChannelID = types.StringValue(invite.Channel.ID)
	state.ServerID = types.StringValue(invite.Guild.ID)
	state.MaxAge = types.Int64Value(int64(invite.MaxAge))
	state.MaxUses = types.Int64Value(int64(invite.MaxUses))
	state.Temporary = types.BoolValue(invite.Temporary)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *inviteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state inviteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.InviteDelete(state.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Invite", err.Error())
		return
	}
}

func (r *inviteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Updates not supported", "Discord invites cannot be updated after creation. The resource will be recreated.")
}

func (r *inviteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("code"), req, resp)
}
