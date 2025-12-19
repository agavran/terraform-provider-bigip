/*
Copyright 2019 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

package bigip

import (
	"fmt"
	"testing"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TEST_MGMT_ROUTE_RESOURCE_NAME = "bigip_sys_mgmt_route.test-route"
var TEST_MGMT_ROUTE_NAME = "/Common/test-route"
var TEST_MGMT_ROUTE_RESOURCE = `
resource "bigip_sys_mgmt_route" "test-route" {
	  name = "` + TEST_MGMT_ROUTE_NAME + `"
	  network = "33.30.31.0/24"
	  gateway = "10.171.125.61"
}
`
var TEST_MGMT_ROUTE_RESOURCE_UPDATE = `
resource "bigip_sys_mgmt_route" "test-route" {
          name = "` + TEST_MGMT_ROUTE_NAME + `"
          network = "33.30.31.0/24"
          gateway = "10.171.125.62"
}
`

func TestAccBigipSysMgmtRoute_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckMgmtRouteDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_MGMT_ROUTE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckMgmtRouteExists(TEST_MGMT_ROUTE_NAME, true),
					testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_route.test-route", "name", "/Common/test-route"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_route.test-route", "network", "33.30.31.0/24"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_route.test-route", "gateway", "10.171.125.61"),
				),
			},
		},
	})
}
func TestAccBigipSysMgmtRoute_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckMgmtRouteDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_MGMT_ROUTE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_route.test-route", "gateway", "10.171.125.61"),
				),
			},
			{
				Config: TEST_MGMT_ROUTE_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSleep(30),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_route.test-route", "gateway", "10.171.125.62"),
				),
			},
		},
	})
}
func TestAccBigipSysMgmtRoute_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckMgmtRouteDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_MGMT_ROUTE_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckSleep(30),
					testCheckMgmtRouteExists(TEST_MGMT_ROUTE_NAME, true),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_route.test-route", "name", "/Common/test-route"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_route.test-route", "network", "33.30.31.0/24"),
					resource.TestCheckResourceAttr("bigip_sys_mgmt_route.test-route", "gateway", "10.171.125.61"),
				),
			},
			{
				ResourceName:      TEST_MGMT_ROUTE_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_MGMT_ROUTE_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckMgmtRouteExists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetManagementRoute(name)
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("route %s was not created.", name)
		}
		if !exists && p != nil {
			return fmt.Errorf("route %s still exists.", name)
		}
		return nil
	}
}

func testCheckMgmtRouteDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_sys_mgmt_route" {
			continue
		}

		name := rs.Primary.ID
		route, err := client.GetManagementRoute(name)
		if err != nil {
			return err
		}
		if route != nil {
			return fmt.Errorf("route %s not destroyed.", name)
		}
		fmt.Printf("[INFO] Management route '%s' has been successfully destroyed\n", name)
	}
	return nil
}
