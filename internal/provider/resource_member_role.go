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
	_ resource.Resource                = &memberRoleResource{}
	_ resource.ResourceWithConfigure   = &memberRoleResource{}
	_ resource.ResourceWithImportState = &memberRoleResource{}
)

func NewMemberRoleResource() resource.Resource {
	return &memberRoleResource{}
}

type memberRoleResource struct {
	client *discordgo.Session
}

type memberRoleResourceModel struct {
	ID       types.String `tfsdk:"id"`
	ServerID types.String `tfsdk:"server_id"`
	UserID   types.String `tfsdk:"user_id"`
	RoleID   types.String `tfsdk:"role_id"`
}

func (r *memberRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_member_role"
}

func (r *memberRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Assigns a role to a guild member. Removing the resource removes the role from the member.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID of server_id:user_id:role_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "ID of the server (guild).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "ID of the member (user).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_id": schema.StringAttribute{
				Description: "ID of the role to assign.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *memberRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *memberRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan memberRoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.GuildMemberRoleAdd(
		plan.ServerID.ValueString(),
		plan.UserID.ValueString(),
		plan.RoleID.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error Assigning Role", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s",
		plan.ServerID.ValueString(),
		plan.UserID.ValueString(),
		plan.RoleID.ValueString(),
	))

	tflog.Trace(ctx, "assigned role to member")
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *memberRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state memberRoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.GuildMember(state.ServerID.ValueString(), state.UserID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	hasRole := false
	for _, rid := range member.Roles {
		if rid == state.RoleID.ValueString() {
			hasRole = true
			break
		}
	}
	if !hasRole {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *memberRoleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields are RequiresReplace, so Update is never called.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"All fields in discord_member_role force resource replacement; update should never be called.",
	)
}

func (r *memberRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state memberRoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.GuildMemberRoleRemove(
		state.ServerID.ValueString(),
		state.UserID.ValueString(),
		state.RoleID.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error Removing Role", err.Error())
		return
	}
}

func (r *memberRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
