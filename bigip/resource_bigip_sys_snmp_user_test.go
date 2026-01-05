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

var TEST_SNMP_USER_NAME = "/Common/test-snmpv3-user"
var TEST_SNMP_USER_RESOURCE_NAME = "bigip_sys_snmp_user.test-user"
var TEST_SNMP_USER_RESOURCE = `
resource "bigip_sys_snmp_user" "test-user" {
	name              = "` + TEST_SNMP_USER_NAME + `"
	username          = "testuser"
	auth_protocol     = "sha"
	auth_password     = "authpass123"
	privacy_protocol  = "aes"
	privacy_password  = "privpass456"
	access            = "ro"
	security_level    = "auth-privacy"
	description       = "Test SNMP user"
}
`

var TEST_SNMP_USER_RESOURCE_UPDATE = `
resource "bigip_sys_snmp_user" "test-user" {
	name              = "` + TEST_SNMP_USER_NAME + `"
	username          = "testuserupdated"
	auth_protocol     = "sha256"
	auth_password     = "newauthpass"
	privacy_protocol  = "aes256"
	privacy_password  = "newprivpass"
	access            = "rw"
	security_level    = "auth-privacy"
	description       = "Updated SNMP user"
}
`

func TestAccBigipSysSnmpUser_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_SNMP_USER_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSnmpUserExists(TEST_SNMP_USER_NAME, true),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "name", "/Common/test-snmpv3-user"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "username", "testuser"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "auth_protocol", "sha"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "privacy_protocol", "aes"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "access", "ro"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "security_level", "auth-privacy"),
				),
			},
		},
	})
}

func TestAccBigipSysSnmpUser_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSnmpUserDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SNMP_USER_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "username", "testuser"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "access", "ro"),
				),
			},
			{
				Config: TEST_SNMP_USER_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "username", "testuserupdated"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "auth_protocol", "sha256"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "privacy_protocol", "aes256"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_user.test-user", "access", "rw"),
				),
			},
		},
	})
}

func TestAccBigipSysSnmpUser_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSnmpUserDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SNMP_USER_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSnmpUserExists(TEST_SNMP_USER_NAME, true),
				),
			},
			{
				ResourceName:            TEST_SNMP_USER_RESOURCE_NAME,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_password", "privacy_password"},
				ImportStateId:           TEST_SNMP_USER_NAME,
				ImportStateCheck:        testPrintImportedState,
			},
		},
	})
}

func testCheckSnmpUserExists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetSnmpUser(name)
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("SNMP user %s was not created.", name)
		}
		if !exists && p != nil {
			return fmt.Errorf("SNMP user %s still exists.", name)
		}
		return nil
	}
}

func testCheckSnmpUserDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_snmp_user" {
			continue
		}

		name := rs.Primary.ID
		user, err := client.GetSnmpUser(name)
		if err != nil {
			return err
		}
		if user != nil {
			return fmt.Errorf("SNMP user %s not destroyed.", name)
		}
		fmt.Printf("[INFO] SNMP User '%s' has been successfully destroyed\n", name)
	}
	return nil
}
