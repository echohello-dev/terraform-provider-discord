package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRoleResourceModel_Defaults(t *testing.T) {
	m := roleResourceModel{
		ID:          types.StringValue("role-1"),
		ServerID:    types.StringValue("guild-1"),
		Name:        types.StringValue("Moderator"),
		Permissions: types.StringValue("268435456"),
		Color:       types.Int64Value(3447003),
		Hoist:       types.BoolValue(true),
		Mentionable: types.BoolValue(true),
		Position:    types.Int64Null(),
	}

	if m.ID.ValueString() != "role-1" {
		t.Fatalf("expected ID role-1, got %s", m.ID.ValueString())
	}
	if m.Name.ValueString() != "Moderator" {
		t.Fatalf("expected name Moderator, got %s", m.Name.ValueString())
	}
	if m.Permissions.ValueString() != "268435456" {
		t.Fatalf("expected permissions 268435456, got %s", m.Permissions.ValueString())
	}
	if m.Color.ValueInt64() != 3447003 {
		t.Fatalf("expected color 3447003, got %d", m.Color.ValueInt64())
	}
	if !m.Hoist.ValueBool() {
		t.Fatal("expected Hoist true")
	}
	if !m.Mentionable.ValueBool() {
		t.Fatal("expected Mentionable true")
	}
}

func TestParsePermissions(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"268435456", 268435456, false},
		{"0", 0, false},
		{"-1", -1, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parsePermissions(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePermissions(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.expected {
				t.Fatalf("parsePermissions(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}
