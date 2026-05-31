package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"discord": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add validation here (e.g., ensure DISCORD_TOKEN is set for real API tests).
	// For mocked tests this is a no-op.
}

// TestProviderMeta is a basic unit test that validates the provider schema.
func TestProviderMeta(t *testing.T) {
	// Ensure New() returns a function that produces a valid provider.
	p := New("test")()
	if p == nil {
		t.Fatal("provider factory returned nil")
	}
}

// TestUnitParsePermissions ensures the parsePermissions helper works.
func TestUnitParsePermissions(t *testing.T) {
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

// TestAccExampleServer is an example acceptance test.
// Run with: TF_ACC=1 mise run test -- TestAccExampleServer
func TestAccExampleServer(t *testing.T) {
	// Skip if TF_ACC is not set — keeps this from touching real infrastructure
	// outside of explicit acceptance runs.
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccExampleServerConfig("Test Server"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_server.test", "name", "Test Server"),
				),
			},
			// Update and Read testing
			{
				Config: testAccExampleServerConfig("Test Server Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("discord_server.test", "name", "Test Server Updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccExampleServerConfig(name string) string {
	return `resource "discord_server" "test" {
	name = "` + name + `"
}`
}
