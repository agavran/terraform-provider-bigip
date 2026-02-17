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

var TEST_CM_DEVICE_NAME_RESOURCE_NAME = "bigip_cm_device_name.test-device-name"
var TEST_CM_DEVICE_NAME_RESOURCE = `
resource "bigip_cm_device_name" "test-device-name" {
	name = "bigip-test.example.com"
}
`
var TEST_CM_DEVICE_NAME_RESOURCE_UPDATE = `
resource "bigip_cm_device_name" "test-device-name" {
	name = "bigip-updated.example.com"
}
`

func TestAccBigipCmDeviceName_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckCmDeviceNameDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_CM_DEVICE_NAME_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckCmDeviceNameExists("bigip-test.example.com"),
					resource.TestCheckResourceAttr("bigip_cm_device_name.test-device-name", "name", "bigip-test.example.com"),
				),
			},
		},
	})
}

func TestAccBigipCmDeviceName_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckCmDeviceNameDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_CM_DEVICE_NAME_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckCmDeviceNameExists("bigip-test.example.com"),
					resource.TestCheckResourceAttr("bigip_cm_device_name.test-device-name", "name", "bigip-test.example.com"),
				),
			},
			{
				Config: TEST_CM_DEVICE_NAME_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					testCheckCmDeviceNameExists("bigip-updated.example.com"),
					resource.TestCheckResourceAttr("bigip_cm_device_name.test-device-name", "name", "bigip-updated.example.com"),
				),
			},
		},
	})
}

func TestAccBigipCmDeviceName_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckCmDeviceNameDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_CM_DEVICE_NAME_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckCmDeviceNameExists("bigip-test.example.com"),
				),
			},
			{
				ResourceName:      TEST_CM_DEVICE_NAME_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckCmDeviceNameExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		name, err := client.GetSelfDeviceName()
		if err != nil {
			return err
		}
		if name != expectedName {
			return fmt.Errorf("CM Device Name mismatch: got %q, expected %q", name, expectedName)
		}
		return nil
	}
}

func testCheckCmDeviceNameDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_cm_device_name" {
			continue
		}

		name, err := client.GetSelfDeviceName()
		if err != nil {
			return err
		}
		if name != "bigip1" {
			return fmt.Errorf("CM Device Name not reset to default (bigip1), got: %s", name)
		}
	}
	return nil
}
