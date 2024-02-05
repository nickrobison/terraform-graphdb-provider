package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Setup users
			{
				Config: providerConfig + `
resource "graphdb_user" "user1" {
  username = "DSUser1"
  role = "user"
}

resource "graphdb_user" "manager1" {
  username = "DSManager1"
  password = "Hello1"
  role = "repo-manager"
}

resource "graphdb_user" "admin1" {
  username = "DSadmin1"
  role = "admin"
}
 `,
			},
			{
				Config: providerConfig + `

data "graphdb_users" "users" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// The admin user already exists
					resource.TestCheckResourceAttr("data.graphdb_users.users", "users.#", "4"),
				),
			},
		},
	})
}
