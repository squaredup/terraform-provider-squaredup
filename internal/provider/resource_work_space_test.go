package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceWorkSpace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create Test
			{
				Config: providerConfig + `
				resource "squaredup_workspace" "test" {
					display_name = "Workspace Test"
					description = "Workspace Used for Testing"
					type = "application"
					tags = ["test", "test2"]
					open_access_enabled = true
					}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_workspace.test", "display_name", "Workspace Test"),
					//Check Dynamic Values
					resource.TestCheckResourceAttrSet("squaredup_workspace.test", "id"),
					resource.TestCheckResourceAttrSet("squaredup_workspace.test", "last_updated"),
				),
			},
			// Import Test
			{
				ResourceName:            "squaredup_workspace.test",
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update Test
			{
				Config: providerConfig + `
					resource "squaredup_workspace" "test" {
						display_name = "Workspace Test - Updated"
					}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_workspace.test", "display_name", "Workspace Test - Updated"),
				),
			},
		},
	})
}
