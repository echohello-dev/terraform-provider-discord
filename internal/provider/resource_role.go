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

var (
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithConfigure   = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
)

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	client *discordgo.Session
}

type roleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ServerID    types.String `tfsdk:"server_id"`
	Name        types.String `tfsdk:"name"`
	Permissions types.String `tfsdk:"permissions"`
	Color       types.Int64  `tfsdk:"color"`
	Hoist       types.Bool   `tfsdk:"hoist"`
	Mentionable types.Bool   `tfsdk:"mentionable"`
	Position    types.Int64  `tfsdk:"position"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Discord role within a server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "The ID of the server (guild) the role belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the role (1-100 characters).",
				Required:    true,
			},
			"permissions": schema.StringAttribute{
				Description: "The permission bitset string for the role.",
				Optional:    true,
				Computed:    true,
			},
			"color": schema.Int64Attribute{
				Description: "The color of the role (RGB integer value, 0 for no color).",
				Optional:    true,
				Computed:    true,
			},
			"hoist": schema.BoolAttribute{
				Description: "Whether the role is displayed separately in the sidebar.",
				Optional:    true,
				Computed:    true,
			},
			"mentionable": schema.BoolAttribute{
				Description: "Whether the role is mentionable by everyone.",
				Optional:    true,
				Computed:    true,
			},
			"position": schema.Int64Attribute{
				Description: "The position of the role in the role list (0 is lowest).",
				Optional:    true,
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	perms := int64(discordgo.PermissionViewChannel)
	if !plan.Permissions.IsNull() && !plan.Permissions.IsUnknown() {
		var err error
		perms, err = parsePermissions(plan.Permissions.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid Permissions", err.Error())
			return
		}
	}

	color := 0
	if !plan.Color.IsNull() && !plan.Color.IsUnknown() {
		color = int(plan.Color.ValueInt64())
	}

	hoist := false
	if !plan.Hoist.IsNull() && !plan.Hoist.IsUnknown() {
		hoist = plan.Hoist.ValueBool()
	}

	mentionable := false
	if !plan.Mentionable.IsNull() && !plan.Mentionable.IsUnknown() {
		mentionable = plan.Mentionable.ValueBool()
	}

	role, err := r.client.GuildRoleCreate(plan.ServerID.ValueString(), &discordgo.RoleParams{
		Name:        plan.Name.ValueString(),
		Permissions: &perms,
		Color:       &color,
		Hoist:       &hoist,
		Mentionable: &mentionable,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Role", err.Error())
		return
	}

	plan.ID = types.StringValue(role.ID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	if plan.Permissions.IsUnknown() || plan.Permissions.IsNull() {
		plan.Permissions = types.StringValue(fmt.Sprintf("%d", role.Permissions))
	}
	if plan.Color.IsUnknown() || plan.Color.IsNull() {
		plan.Color = types.Int64Value(int64(role.Color))
	}
	if plan.Hoist.IsUnknown() || plan.Hoist.IsNull() {
		plan.Hoist = types.BoolValue(role.Hoist)
	}
	if plan.Mentionable.IsUnknown() || plan.Mentionable.IsNull() {
		plan.Mentionable = types.BoolValue(role.Mentionable)
	}

	// Handle position if set
	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		_, err = r.client.GuildRoleReorder(plan.ServerID.ValueString(), []*discordgo.Role{
			{ID: role.ID, Position: int(plan.Position.ValueInt64())},
		})
		if err != nil {
			resp.Diagnostics.AddWarning("Error Setting Role Position", err.Error())
		}
	}

	tflog.Trace(ctx, "created a role")
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles, err := r.client.GuildRoles(state.ServerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Roles", err.Error())
		return
	}

	var found *discordgo.Role
	for _, role := range roles {
		if role.ID == state.ID.ValueString() {
			found = role
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(found.Name)
	state.Permissions = types.StringValue(fmt.Sprintf("%d", found.Permissions))
	state.Color = types.Int64Value(int64(found.Color))
	state.Hoist = types.BoolValue(found.Hoist)
	state.Mentionable = types.BoolValue(found.Mentionable)
	state.Position = types.Int64Value(int64(found.Position))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	perms, err := parsePermissions(plan.Permissions.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Permissions", err.Error())
		return
	}

	color := int(plan.Color.ValueInt64())
	hoist := plan.Hoist.ValueBool()
	mentionable := plan.Mentionable.ValueBool()

	_, err = r.client.GuildRoleEdit(plan.ServerID.ValueString(), plan.ID.ValueString(), &discordgo.RoleParams{
		Name:        plan.Name.ValueString(),
		Permissions: &perms,
		Color:       &color,
		Hoist:       &hoist,
		Mentionable: &mentionable,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Role", err.Error())
		return
	}

	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		_, err = r.client.GuildRoleReorder(plan.ServerID.ValueString(), []*discordgo.Role{
			{ID: plan.ID.ValueString(), Position: int(plan.Position.ValueInt64())},
		})
		if err != nil {
			resp.Diagnostics.AddWarning("Error Updating Role Position", err.Error())
		}
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.GuildRoleDelete(state.ServerID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Role", err.Error())
		return
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
