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
					sharing_authorized_email_domains = ["test.com"]
					}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_workspace.test", "display_name", `Workspace Test `+uuid),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "description", "Workspace Used for Testing"),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "type", "application"),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "tags.0", "test"),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "tags.1", "test2"),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "allow_dashboard_sharing", "true"),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "sharing_authorized_email_domains.#", "1"),
					resource.TestCheckTypeSetElemAttr("squaredup_workspace.test", "sharing_authorized_email_domains.*", "test.com"),
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
						allow_dashboard_sharing = false
						sharing_authorized_email_domains = []
						type = "other"
					}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_workspace.test", "display_name", `Workspace Test `+uuid+`- Updated`),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "allow_dashboard_sharing", "false"),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "type", "other"),
					resource.TestCheckResourceAttr("squaredup_workspace.test", "sharing_authorized_email_domains.#", "0"),
				),
			},
		},
	})
}
