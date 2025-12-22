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

var TEST_SYS_AUTH_SOURCE_NAME = "auth_source"
var TEST_AUTH_SOURCE_RESOURCE_NAME = "bigip_sys_auth_source.test-auth-source"
var TEST_AUTH_SOURCE_RESOURCE = `
resource "bigip_sys_auth_source" "test-auth-source" {
	type     = "ldap"
	fallback = "true"
}
`
var TEST_AUTH_SOURCE_RESOURCE_UPDATE = `
resource "bigip_sys_auth_source" "test-auth-source" {
	type     = "local"
	fallback = "true"
}
`

func TestAccBigipSysAuthSource_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_AUTH_SOURCE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckAuthSourceExists(true),
					resource.TestCheckResourceAttr("bigip_sys_auth_source.test-auth-source", "type", "ldap"),
					resource.TestCheckResourceAttr("bigip_sys_auth_source.test-auth-source", "fallback", "true"),
				),
			},
		},
	})
}

func TestAccBigipSysAuthSource_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckAuthSourceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_AUTH_SOURCE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_auth_source.test-auth-source", "type", "ldap"),
					resource.TestCheckResourceAttr("bigip_sys_auth_source.test-auth-source", "fallback", "true"),
				),
			},
			{
				Config: TEST_AUTH_SOURCE_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_auth_source.test-auth-source", "type", "local"),
					resource.TestCheckResourceAttr("bigip_sys_auth_source.test-auth-source", "fallback", "true"),
				),
			},
		},
	})
}

func TestAccBigipSysAuthSource_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckAuthSourceDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_AUTH_SOURCE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckAuthSourceExists(true),
				),
			},
			{
				ResourceName:      TEST_AUTH_SOURCE_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_SYS_AUTH_SOURCE_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckAuthSourceExists(exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetAuthSource()
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("auth source configuration was not created.")
		}
		if !exists && p != nil {
			return fmt.Errorf("auth source configuration still exists.")
		}
		return nil
	}
}

func testCheckAuthSourceDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_auth_source" {
			continue
		}

		authSource, err := client.GetAuthSource()
		if err != nil {
			return err
		}
		if authSource.Type != "local" {
			return fmt.Errorf("auth source type not reset to default (local), got: %s", authSource.Type)
		}
		if authSource.Fallback != "false" {
			return fmt.Errorf("auth source fallback not reset to default (false), got: %s", authSource.Fallback)
		}
		fmt.Println("[INFO] Auth source successfully reset to default values (type: local, fallback: false)")
	}
	return nil
}
