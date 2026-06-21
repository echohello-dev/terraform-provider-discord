package provider

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &messageResource{}
	_ resource.ResourceWithConfigure = &messageResource{}
)

func NewMessageResource() resource.Resource {
	return &messageResource{}
}

type messageResource struct {
	client *discordgo.Session
}

type messageResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ChannelID types.String `tfsdk:"channel_id"`
	Content   types.String `tfsdk:"content"`
	TTS       types.Bool   `tfsdk:"tts"`
	EmbedJSON types.String `tfsdk:"embed_json"`
	AuthorID  types.String `tfsdk:"author_id"`
	Timestamp types.String `tfsdk:"timestamp"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *messageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_message"
}

func (r *messageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sends a message to a Discord channel. Create-only; recreating the resource sends a new message.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Snowflake ID of the sent message.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"channel_id": schema.StringAttribute{
				Description: "ID of the channel to send the message to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Description: "Message text (up to 2000 characters).",
				Optional:    true,
			},
			"tts": schema.BoolAttribute{
				Description:   "Whether the message should be read aloud by clients (text-to-speech).",
				Optional:      true,
				PlanModifiers: []planmodifier.Bool{
					// No bool plan modifier required for RequiresReplace; bools use RequiresReplace via custom modifier if needed.
				},
				Default: nil,
			},
			"embed_json": schema.StringAttribute{
				Description: "JSON-encoded Discord embed object. Conflicts with content.",
				Optional:    true,
			},
			"author_id": schema.StringAttribute{
				Description: "ID of the user that sent the message.",
				Computed:    true,
			},
			"timestamp": schema.StringAttribute{
				Description: "ISO timestamp when the message was sent.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "ISO timestamp when the message was sent (alias of timestamp).",
				Computed:    true,
			},
		},
	}
}

func (r *messageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *messageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan messageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Content.IsNull() && plan.EmbedJSON.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Content",
			"Either 'content' or 'embed_json' must be set.",
		)
		return
	}
	if !plan.Content.IsNull() && !plan.EmbedJSON.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting Content",
			"Only one of 'content' or 'embed_json' may be set.",
		)
		return
	}

	data := &discordgo.MessageSend{
		TTS: !plan.TTS.IsNull() && plan.TTS.ValueBool(),
	}
	if !plan.Content.IsNull() {
		data.Content = plan.Content.ValueString()
	}

	msg, err := r.client.ChannelMessageSendComplex(plan.ChannelID.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("Error Sending Message", err.Error())
		return
	}

	plan.ID = types.StringValue(msg.ID)
	plan.AuthorID = types.StringValue(msg.Author.ID)
	plan.Timestamp = types.StringValue(msg.Timestamp.Format("2006-01-02T15:04:05Z07:00"))
	plan.CreatedAt = types.StringValue(msg.Timestamp.Format("2006-01-02T15:04:05Z07:00"))

	tflog.Trace(ctx, "sent message", map[string]any{"channel_id": plan.ChannelID.ValueString(), "message_id": msg.ID})
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *messageResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// Messages are immutable and create-only; Read is a no-op.
}

func (r *messageResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"discord_message is create-only; modify the message via the Discord API or recreate the resource.",
	)
}

func (r *messageResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Create-only resource; Delete is a no-op so we don't delete the message from Discord.
}
