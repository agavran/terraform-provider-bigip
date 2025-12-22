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

var TEST_REMOTE_USER_NAME = "remote_user"
var TEST_REMOTE_USER_RESOURCE_NAME = "bigip_sys_remote_user.test-remote-user"
var TEST_REMOTE_USER_RESOURCE = `
resource "bigip_sys_remote_user" "test-remote-user" {
	default_partition       = "Common"
	default_role            = "guest"
	remote_console_access   = "disabled"
}
`
var TEST_REMOTE_USER_RESOURCE_UPDATE = `
resource "bigip_sys_remote_user" "test-remote-user" {
	default_partition       = "all"
	default_role            = "operator"
	remote_console_access   = "tmsh"
}
`

func TestAccBigipSysRemoteUser_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckRemoteUserDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_REMOTE_USER_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckRemoteUserExists(true),
					resource.TestCheckResourceAttr("bigip_sys_remote_user.test-remote-user", "default_partition", "Common"),
					resource.TestCheckResourceAttr("bigip_sys_remote_user.test-remote-user", "default_role", "guest"),
					resource.TestCheckResourceAttr("bigip_sys_remote_user.test-remote-user", "remote_console_access", "disabled"),
				),
			},
		},
	})
}

func TestAccBigipSysRemoteUser_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckRemoteUserDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_REMOTE_USER_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_remote_user.test-remote-user", "default_partition", "Common"),
					resource.TestCheckResourceAttr("bigip_sys_remote_user.test-remote-user", "default_role", "guest"),
					resource.TestCheckResourceAttr("bigip_sys_remote_user.test-remote-user", "remote_console_access", "disabled"),
				),
			},
			{
				Config: TEST_REMOTE_USER_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_remote_user.test-remote-user", "default_partition", "all"),
					resource.TestCheckResourceAttr("bigip_sys_remote_user.test-remote-user", "default_role", "operator"),
					resource.TestCheckResourceAttr("bigip_sys_remote_user.test-remote-user", "remote_console_access", "tmsh"),
				),
			},
		},
	})
}

func TestAccBigipSysRemoteUser_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckRemoteUserDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_REMOTE_USER_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckRemoteUserExists(true),
				),
			},
			{
				ResourceName:      TEST_REMOTE_USER_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_REMOTE_USER_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckRemoteUserExists(exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetRemoteUser()
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("remote user configuration was not created.")
		}
		if !exists && p != nil {
			return fmt.Errorf("remote user configuration still exists.")
		}
		return nil
	}
}

func testCheckRemoteUserDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_remote_user" {
			continue
		}

		remoteUser, err := client.GetRemoteUser()
		if err != nil {
			return err
		}
		if remoteUser.DefaultPartition != "all" {
			return fmt.Errorf("remote user default partition not reset to default (all), got: %s", remoteUser.DefaultPartition)
		}
		if remoteUser.DefaultRole != "no-access" {
			return fmt.Errorf("remote user default role not reset to default (no-access), got: %s", remoteUser.DefaultRole)
		}
		if remoteUser.RemoteConsoleAccess != "disabled" {
			return fmt.Errorf("remote user console access not reset to default (disabled), got: %s", remoteUser.RemoteConsoleAccess)
		}
		fmt.Println("[INFO] Remote user successfully reset to default values (default partition: all, default role: no-access, console access: disabled)")
	}
	return nil
}
