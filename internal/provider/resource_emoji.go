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
	_ resource.Resource                = &emojiResource{}
	_ resource.ResourceWithConfigure   = &emojiResource{}
	_ resource.ResourceWithImportState = &emojiResource{}
)

func NewEmojiResource() resource.Resource {
	return &emojiResource{}
}

type emojiResource struct {
	client *discordgo.Session
}

type emojiResourceModel struct {
	ID       types.String `tfsdk:"id"`
	ServerID types.String `tfsdk:"server_id"`
	Name     types.String `tfsdk:"name"`
	Image    types.String `tfsdk:"image"`
	Roles    types.List   `tfsdk:"roles"`
}

func (r *emojiResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_emoji"
}

func (r *emojiResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom emoji within a Discord server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the emoji.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "The ID of the server the emoji belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the emoji (2-32 characters).",
				Required:    true,
			},
			"image": schema.StringAttribute{
				Description: "The base64-encoded image data (must be PNG, JPG, GIF, or WEBP, max 256KB).",
				Required:    true,
				Sensitive:   true,
			},
			"roles": schema.ListAttribute{
				Description: "Roles that can use this emoji.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *emojiResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *emojiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emojiResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles := make([]string, 0)
	if !plan.Roles.IsNull() && !plan.Roles.IsUnknown() {
		_ = plan.Roles.ElementsAs(ctx, &roles, false)
	}

	emoji, err := r.client.GuildEmojiCreate(plan.ServerID.ValueString(), &discordgo.EmojiParams{
		Name:  plan.Name.ValueString(),
		Image: plan.Image.ValueString(),
		Roles: roles,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Emoji", err.Error())
		return
	}

	plan.ID = types.StringValue(emoji.ID)
	tflog.Trace(ctx, "created an emoji")

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *emojiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state emojiResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	emoji, err := r.client.GuildEmoji(state.ServerID.ValueString(), state.ID.ValueString())
	if err != nil {
		if err.Error() == "HTTP 404 Not Found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Emoji", err.Error())
		return
	}

	state.Name = types.StringValue(emoji.Name)

	if len(emoji.Roles) > 0 {
		roles, _ := types.ListValueFrom(ctx, types.StringType, emoji.Roles)
		state.Roles = roles
	} else {
		state.Roles = types.ListNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *emojiResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan emojiResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles := make([]string, 0)
	if !plan.Roles.IsNull() && !plan.Roles.IsUnknown() {
		_ = plan.Roles.ElementsAs(ctx, &roles, false)
	}

	_, err := r.client.GuildEmojiEdit(plan.ServerID.ValueString(), plan.ID.ValueString(), &discordgo.EmojiParams{
		Name:  plan.Name.ValueString(),
		Roles: roles,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Emoji", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *emojiResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state emojiResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.GuildEmojiDelete(state.ServerID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Emoji", err.Error())
		return
	}
}

func (r *emojiResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
