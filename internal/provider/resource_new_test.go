package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEmojiResourceModel_Defaults(t *testing.T) {
	m := emojiResourceModel{
		ID:       types.StringValue("emoji123"),
		ServerID: types.StringValue("server456"),
		Name:     types.StringValue("test_emoji"),
		Image:    types.StringNull(),
		Roles:    types.ListNull(types.StringType),
	}

	if m.ID.ValueString() != "emoji123" {
		t.Fatalf("expected ID emoji123, got %s", m.ID.ValueString())
	}
	if m.ServerID.ValueString() != "server456" {
		t.Fatalf("expected ServerID server456, got %s", m.ServerID.ValueString())
	}
	if m.Name.ValueString() != "test_emoji" {
		t.Fatalf("expected Name test_emoji, got %s", m.Name.ValueString())
	}
	if !m.Image.IsNull() {
		t.Fatal("expected Image to be null")
	}
}

func TestEmojiResourceModel_WithRoles(t *testing.T) {
	roles, _ := types.ListValueFrom(context.Background(), types.StringType, []string{"role1", "role2"})
	m := emojiResourceModel{
		ID:       types.StringValue("emoji123"),
		ServerID: types.StringValue("server456"),
		Name:     types.StringValue("test_emoji"),
		Image:    types.StringNull(),
		Roles:    roles,
	}

	if m.Roles.IsNull() {
		t.Fatal("expected Roles to not be null")
	}
}

func TestWebhookResourceModel_Defaults(t *testing.T) {
	m := webhookResourceModel{
		ID:        types.StringValue("webhook123"),
		ChannelID: types.StringValue("channel456"),
		ServerID:  types.StringValue("server789"),
		Name:      types.StringValue("test_webhook"),
		Avatar:    types.StringNull(),
		Token:     types.StringNull(),
	}

	if m.ID.ValueString() != "webhook123" {
		t.Fatalf("expected ID webhook123, got %s", m.ID.ValueString())
	}
	if m.Name.ValueString() != "test_webhook" {
		t.Fatalf("expected Name test_webhook, got %s", m.Name.ValueString())
	}
	if !m.Avatar.IsNull() {
		t.Fatal("expected Avatar to be null")
	}
}

func TestInviteResourceModel_Defaults(t *testing.T) {
	m := inviteResourceModel{
		Code:       types.StringValue("abc123"),
		ChannelID:  types.StringValue("channel456"),
		ServerID:   types.StringValue("server789"),
		MaxAge:     types.Int64Value(0),
		MaxUses:    types.Int64Value(0),
		Temporary:  types.BoolValue(false),
		Unique:     types.BoolValue(true),
		TargetType: types.Int64Null(),
	}

	if m.Code.ValueString() != "abc123" {
		t.Fatalf("expected Code abc123, got %s", m.Code.ValueString())
	}
	if m.MaxAge.ValueInt64() != 0 {
		t.Fatalf("expected MaxAge 0, got %d", m.MaxAge.ValueInt64())
	}
	if !m.Unique.ValueBool() {
		t.Fatal("expected Unique to be true")
	}
}

func TestInviteResourceModel_WithMaxAge(t *testing.T) {
	m := inviteResourceModel{
		Code:      types.StringValue("abc123"),
		ChannelID: types.StringValue("channel456"),
		ServerID:  types.StringValue("server789"),
		MaxAge:    types.Int64Value(3600),
		MaxUses:   types.Int64Value(10),
		Temporary: types.BoolValue(true),
		Unique:    types.BoolValue(false),
	}

	if m.MaxAge.ValueInt64() != 3600 {
		t.Fatalf("expected MaxAge 3600, got %d", m.MaxAge.ValueInt64())
	}
	if m.MaxUses.ValueInt64() != 10 {
		t.Fatalf("expected MaxUses 10, got %d", m.MaxUses.ValueInt64())
	}
	if !m.Temporary.ValueBool() {
		t.Fatal("expected Temporary to be true")
	}
}

func TestParsePermissions_Valid(t *testing.T) {
	result, err := parsePermissions("1024")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 1024 {
		t.Fatalf("expected 1024, got %d", result)
	}
}

func TestParsePermissions_Empty(t *testing.T) {
	_, err := parsePermissions("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestParsePermissions_Invalid(t *testing.T) {
	_, err := parsePermissions("not_a_number")
	if err == nil {
		t.Fatal("expected error for invalid string")
	}
}

func TestFormatPermissions(t *testing.T) {
	result := formatPermissions(1024)
	if result != "1024" {
		t.Fatalf("expected 1024, got %s", result)
	}
}

func TestInt64Ptr(t *testing.T) {
	val := int64Ptr(42)
	if *val != 42 {
		t.Fatalf("expected 42, got %d", *val)
	}
}

func TestBoolPtr(t *testing.T) {
	val := boolPtr(true)
	if *val != true {
		t.Fatalf("expected true, got %v", *val)
	}
}

func TestStringPtr(t *testing.T) {
	val := stringPtr("test")
	if *val != "test" {
		t.Fatalf("expected test, got %s", *val)
	}
}
