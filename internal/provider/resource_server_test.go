package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerResourceModel_Defaults(t *testing.T) {
	m := serverResourceModel{
		ID:              types.StringValue("12345"),
		Name:            types.StringValue("Test Server"),
		Region:          types.StringNull(),
		VerificationLvl: types.Int64Null(),
		DefaultMsgNotif: types.Int64Null(),
		ExplicitContent: types.Int64Null(),
		AfkChannelID:    types.StringNull(),
		AfkTimeout:      types.Int64Null(),
		Icon:            types.StringNull(),
		Splash:          types.StringNull(),
		OwnerID:         types.StringNull(),
		LastUpdated:     types.StringNull(),
	}

	if m.ID.ValueString() != "12345" {
		t.Fatalf("expected ID 12345, got %s", m.ID.ValueString())
	}
	if m.Name.ValueString() != "Test Server" {
		t.Fatalf("expected name Test Server, got %s", m.Name.ValueString())
	}
	if !m.Region.IsNull() {
		t.Fatal("expected Region to be null")
	}
	if !m.VerificationLvl.IsNull() {
		t.Fatal("expected VerificationLvl to be null")
	}
}
