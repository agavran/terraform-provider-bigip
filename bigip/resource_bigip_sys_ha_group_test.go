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

var TEST_HA_GROUP_NAME = "ha_group_test"
var TEST_HA_GROUP_RESOURCE_NAME = "bigip_sys_ha_group.test-ha-group"
var TEST_HA_GROUP_RESOURCE = `
resource "bigip_sys_ha_group" "test-ha-group" {
	name         = "` + TEST_HA_GROUP_NAME + `"
	description  = "Test HA Group"
	active_bonus = 10
	enabled      = true

	pools {
		name              = "/Common/test_pool"
		weight            = 15
		attribute         = "percent-up-members"
		minimum_threshold = 1
	}

	trunks {
		name              = "/Common/test_trunk"
		weight            = 20
		attribute         = "percent-up-members"
		minimum_threshold = 1
	}
}
`

var TEST_HA_GROUP_RESOURCE_UPDATE = `
resource "bigip_sys_ha_group" "test-ha-group" {
	name         = "` + TEST_HA_GROUP_NAME + `"
	description  = "Updated HA Group"
	active_bonus = 20
	enabled      = false

	pools {
		name              = "/Common/test_pool"
		weight            = 25
		attribute         = "percent-up-members"
		minimum_threshold = 2
	}

	pools {
		name              = "/Common/test_pool2"
		weight            = 10
		attribute         = "percent-up-members"
		minimum_threshold = 1
	}
}
`

func TestAccBigipSysHaGroup_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_HA_GROUP_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckHaGroupExists(TEST_HA_GROUP_NAME, true),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "name", TEST_HA_GROUP_NAME),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "description", "Test HA Group"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "active_bonus", "10"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "enabled", "true"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "pools.#", "1"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "pools.0.name", "/Common/test_pool"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "pools.0.weight", "15"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "pools.0.minimum_threshold", "1"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "trunks.#", "1"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "trunks.0.name", "/Common/test_trunk"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "trunks.0.weight", "20"),
				),
			},
		},
	})
}

func TestAccBigipSysHaGroup_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckHaGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_HA_GROUP_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "description", "Test HA Group"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "active_bonus", "10"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "pools.#", "1"),
				),
			},
			{
				Config: TEST_HA_GROUP_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "description", "Updated HA Group"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "active_bonus", "20"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "enabled", "false"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "pools.#", "2"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "pools.0.weight", "25"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "pools.1.name", "/Common/test_pool2"),
					resource.TestCheckResourceAttr("bigip_sys_ha_group.test-ha", "trunks.#", "0"),
				),
			},
		},
	})
}

func TestAccBigipSysHaGroup_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckHaGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_HA_GROUP_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckHaGroupExists(TEST_HA_GROUP_NAME, true),
				),
			},
			{
				ResourceName:      TEST_HA_GROUP_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_HA_GROUP_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckHaGroupExists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetHaGroup(name)
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("HA Group %s was not created.", name)
		}
		if !exists && p != nil {
			return fmt.Errorf("HA Group %s still exists.", name)
		}
		return nil
	}
}

func testCheckHaGroupDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_ha_group" {
			continue
		}

		name := rs.Primary.ID
		haGroup, err := client.GetHaGroup(name)
		if err != nil {
			return err
		}
		if haGroup != nil {
			return fmt.Errorf("HA Group %s not destroyed.", name)
		}
		fmt.Printf("[INFO] HA Group '%s' has been successfully destroyed\n", name)
	}
	return nil
}
