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

var TEST_SNMP_COMMUNITY_NAME = "/Common/test-community"
var TEST_SNMP_COMMUNITY_RESOURCE_NAME = "bigip_sys_snmp_community.test-community"
var TEST_SNMP_COMMUNITY_RESOURCE = `
resource "bigip_sys_snmp_community" "test-community" {
	name           = "` + TEST_SNMP_COMMUNITY_NAME + `"
	community_name = "public123"
	access         = "ro"
	source         = "all"
	description    = "Test community"
}
`
var TEST_SNMP_COMMUNITY_RESOURCE_UPDATE = `
resource "bigip_sys_snmp_community" "test-community" {
	name           = "` + TEST_SNMP_COMMUNITY_NAME + `"
	community_name = "private456"
	access         = "rw"
	source         = "10.0.0.0/8"
	ipv6           = "enabled"
	description    = "Updated test community"
}
`

func TestAccBigipSysSnmpCommunity_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_SNMP_COMMUNITY_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSnmpCommunityExists(TEST_SNMP_COMMUNITY_NAME, true),
					testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_snmp_community.test-community", "name", "/Common/test-community"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_community.test-community", "access", "ro"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_community.test-community", "community_name", "public123"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_community.test-community", "source", "all"),
				),
			},
		},
	})
}

func TestAccBigipSysSnmpCommunity_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSnmpCommunityDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SNMP_COMMUNITY_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_snmp_community.test-community", "access", "ro"),
				),
			},
			{
				Config: TEST_SNMP_COMMUNITY_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_snmp_community.test-community", "access", "rw"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_community.test-community", "source", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("bigip_sys_snmp_community.test-community", "ipv6", "enabled"),
				),
			},
		},
	})
}

func TestAccBigipSysSnmpCommunity_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckSnmpCommunityDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_SNMP_COMMUNITY_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSnmpCommunityExists(TEST_SNMP_COMMUNITY_NAME, true),
				),
			},
			{
				ResourceName:      TEST_SNMP_COMMUNITY_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_SNMP_COMMUNITY_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckSnmpCommunityExists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetSnmpCommunity(name)
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("SNMP community %s was not created.", name)
		}
		if !exists && p != nil {
			return fmt.Errorf("SNMP community %s still exists.", name)
		}
		return nil
	}
}

func testCheckSnmpCommunityDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_snmp_community" {
			continue
		}

		name := rs.Primary.ID
		community, err := client.GetSnmpCommunity(name)
		if err != nil {
			return err
		}
		if community != nil {
			return fmt.Errorf("SNMP community %s not destroyed.", name)
		}
		fmt.Printf("[INFO] SNMP Community '%s' has been successfully destroyed\n", name)
	}
	return nil
}
