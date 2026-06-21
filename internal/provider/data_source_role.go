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
	_ datasource.DataSource              = &roleDataSource{}
	_ datasource.DataSourceWithConfigure = &roleDataSource{}
)

func NewRoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

type roleDataSource struct {
	client *discordgo.Session
}

type roleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	ServerID    types.String `tfsdk:"server_id"`
	Name        types.String `tfsdk:"name"`
	Permissions types.String `tfsdk:"permissions"`
	Color       types.Int64  `tfsdk:"color"`
	Hoist       types.Bool   `tfsdk:"hoist"`
	Mentionable types.Bool   `tfsdk:"mentionable"`
	Managed     types.Bool   `tfsdk:"managed"`
	Position    types.Int64  `tfsdk:"position"`
}

func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a Discord role by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the role.",
				Required:    true,
			},
			"server_id": schema.StringAttribute{
				Description: "ID of the server (guild) the role belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Role name.",
				Computed:    true,
			},
			"permissions": schema.StringAttribute{
				Description: "Bitwise permission integer.",
				Computed:    true,
			},
			"color": schema.Int64Attribute{
				Description: "RGB color value.",
				Computed:    true,
			},
			"hoist": schema.BoolAttribute{
				Description: "Whether members are displayed separately from online members.",
				Computed:    true,
			},
			"mentionable": schema.BoolAttribute{
				Description: "Whether the role can be mentioned by members.",
				Computed:    true,
			},
			"managed": schema.BoolAttribute{
				Description: "Whether the role is managed by an integration (bot role, etc.).",
				Computed:    true,
			},
			"position": schema.Int64Attribute{
				Description: "Position of the role in the role list (higher = higher in the list).",
				Computed:    true,
			},
		},
	}
}

func (d *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state roleDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := d.client.State.Role(state.ServerID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Role",
			fmt.Sprintf("Could not read role %s in server %s: %s",
				state.ID.ValueString(), state.ServerID.ValueString(), err.Error()),
		)
		return
	}

	state.Name = types.StringValue(role.Name)
	state.Permissions = types.StringValue(formatPermissions(role.Permissions))
	state.Color = types.Int64Value(int64(role.Color))
	state.Hoist = types.BoolValue(role.Hoist)
	state.Mentionable = types.BoolValue(role.Mentionable)
	state.Managed = types.BoolValue(role.Managed)
	state.Position = types.Int64Value(int64(role.Position))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
