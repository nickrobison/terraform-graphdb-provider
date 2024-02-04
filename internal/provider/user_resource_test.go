// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test config validation
			{
				Config: providerConfig + `
			resource "graphdb_user" "test1" {
			  username = "TestUser1"
			  password = "SuperSecret"
			  role = "invalid"
			}
			`,
				ExpectError: regexp.MustCompile("Attribute role value must be one of"),
			},
			// Test create and read
			{
				Config: providerConfig + `
			resource "graphdb_user" "test" {
			  username = "TestUser"
			  password = "SuperSecret"
                          role = "user"
			}
			`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graphdb_user.test", "username", "TestUser"),
					resource.TestCheckResourceAttr("graphdb_user.test", "role", "user")),
			},
			// Test import
			{
				ResourceName:            "graphdb_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Test update and read
			{
				Config: providerConfig + `
			resource "graphdb_user" "test" {
			  username = "TestUser"
			  password = "SuperSecret"
                          role = "repo-manager"
			}
			`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graphdb_user.test", "username", "TestUser"),
					resource.TestCheckResourceAttr("graphdb_user.test", "role", "repo-manager")),
			},
		},
	})
}
