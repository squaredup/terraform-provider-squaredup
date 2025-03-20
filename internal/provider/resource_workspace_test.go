package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/pborman/uuid"
)

func TestAccResourceWorkSpace(t *testing.T) {
	uuid := uuid.NewRandom().String()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create Test
			{
				Config: providerConfig + `
				resource "squaredup_workspace" "test" {
					display_name = "Workspace Test ` + uuid + `"
					description = "Workspace Used for Testing"
					type = "application"
					tags = ["test", "test2"]
					allow_dashboard_sharing = true
					}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_workspace.test", "display_name", `Workspace Test `+uuid),
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
						display_name = "Workspace Test ` + uuid + `- Updated"
					}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_workspace.test", "display_name", `Workspace Test `+uuid+`- Updated`),
				),
			},
		},
	})
}
