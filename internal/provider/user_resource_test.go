// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "graphdb_user" "test" {
  username = "TestUser"
  password = "SuperSecret"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("graphdb_user.test", "username", "TestUser")),
			},
		},
	})
}
