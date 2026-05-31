package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerDataSourceModel_Defaults(t *testing.T) {
	m := serverDataSourceModel{
		ID:              types.StringValue("guild-1"),
		Name:            types.StringNull(),
		Region:          types.StringNull(),
		VerificationLvl: types.Int64Null(),
		DefaultMsgNotif: types.Int64Null(),
		ExplicitContent: types.Int64Null(),
		OwnerID:         types.StringNull(),
		MemberCount:     types.Int64Null(),
	}

	if m.ID.ValueString() != "guild-1" {
		t.Fatalf("expected ID guild-1, got %s", m.ID.ValueString())
	}
	if !m.Name.IsNull() {
		t.Fatal("expected Name to be null before read")
	}
	if !m.MemberCount.IsNull() {
		t.Fatal("expected MemberCount to be null before read")
	}
}
