package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestChannelResourceModel_Defaults(t *testing.T) {
	m := channelResourceModel{
		ID:        types.StringValue("chan-1"),
		ServerID:  types.StringValue("guild-1"),
		Name:      types.StringValue("general"),
		Type:      types.Int64Value(0),
		NSFW:      types.BoolNull(),
		Topic:     types.StringNull(),
		Position:  types.Int64Null(),
		ParentID:  types.StringNull(),
		Bitrate:   types.Int64Null(),
		UserLimit: types.Int64Null(),
		RateLimit: types.Int64Null(),
	}

	if m.ID.ValueString() != "chan-1" {
		t.Fatalf("expected ID chan-1, got %s", m.ID.ValueString())
	}
	if m.Name.ValueString() != "general" {
		t.Fatalf("expected name general, got %s", m.Name.ValueString())
	}
	if m.Type.ValueInt64() != 0 {
		t.Fatalf("expected type 0, got %d", m.Type.ValueInt64())
	}
	if !m.NSFW.IsNull() {
		t.Fatal("expected NSFW to be null")
	}
}
