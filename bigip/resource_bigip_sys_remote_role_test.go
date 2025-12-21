/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2025 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/

package bigip

import (
	"fmt"
	"testing"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TEST_REMOTE_ROLE_NAME = "test-remote-role"
var TEST_REMOTE_ROLE_FULL_NAME = "/Common/" + TEST_REMOTE_ROLE_NAME
var TEST_REMOTE_ROLE_RESOURCE_NAME = "bigip_sys_remote_role." + TEST_REMOTE_ROLE_NAME
var TEST_REMOTE_ROLE_RESOURCE = `

resource "bigip_sys_remote_role" "test-remote-role" {
	name           = "` + TEST_REMOTE_ROLE_FULL_NAME + `"
	attribute      = "memberof=cn=testgroup,ou=groups,dc=example,dc=com"
	line_order     = 1000
	role           = "operator"
	user_partition = "all"
	console        = "disabled"
	description    = "Test remote role mapping"
}
`

var TEST_REMOTE_ROLE_RESOURCE_UPDATE = `

resource "bigip_sys_remote_role" "test-remote-role" {
	name           = "` + TEST_REMOTE_ROLE_FULL_NAME + `"
	attribute      = "memberof=cn=updatedgroup,ou=groups,dc=example,dc=com"
	line_order     = 1001
	role           = "admin"
	user_partition = "Common"
	console        = "tmsh"
	description    = "Updated test remote role mapping"
}
`

func TestAccBigipSysRemoteRole_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckRemoteRoleDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_REMOTE_ROLE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckRemoteRoleExists(TEST_REMOTE_ROLE_FULL_NAME, true),
					//testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "name", "/Common/test-remote-role"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "attribute", "memberof=cn=testgroup,ou=groups,dc=example,dc=com"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "line_order", "1000"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "role", "operator"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "user_partition", "all"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "console", "disabled"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "description", "Test remote role mapping"),
				),
			},
		},
	})
}

func TestAccBigipSysRemoteRole_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckRemoteRoleDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_REMOTE_ROLE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "attribute", "memberof=cn=testgroup,ou=groups,dc=example,dc=com"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "line_order", "1000"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "role", "operator"),
				),
			},
			{
				Config: TEST_REMOTE_ROLE_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "attribute", "memberof=cn=updatedgroup,ou=groups,dc=example,dc=com"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "line_order", "1001"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "role", "admin"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "user_partition", "Common"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "console", "tmsh"),
					resource.TestCheckResourceAttr("bigip_sys_remote_role.test-remote-role", "description", "Updated test remote role mapping"),
				),
			},
		},
	})
}

func TestAccBigipSysRemoteRole_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckRemoteRoleDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_REMOTE_ROLE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckRemoteRoleExists(TEST_REMOTE_ROLE_FULL_NAME, true),
				),
			},
			{
				ResourceName:      TEST_REMOTE_ROLE_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_REMOTE_ROLE_FULL_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckRemoteRoleExists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetRemoteRole(name)
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("remote role %s was not created.", name)
		}
		if !exists && p != nil {
			return fmt.Errorf("remote role %s still exists.", name)
		}
		return nil
	}
}

func testCheckRemoteRoleDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_remote_role" {
			continue
		}

		name := rs.Primary.ID
		remoteRole, err := client.GetRemoteRole(name)
		if err != nil {
			return err
		}
		if remoteRole != nil {
			return fmt.Errorf("remote role %s not destroyed.", name)
		}
		fmt.Printf("[INFO] Remote role '%s' has been successfully destroyed\n", name)
	}
	return nil
}
