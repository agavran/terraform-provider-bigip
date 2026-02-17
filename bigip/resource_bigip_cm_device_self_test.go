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

var TEST_CM_DEVICE_SELF_NAME = "device_self"
var TEST_CM_DEVICE_SELF_RESOURCE_NAME = "bigip_cm_device_self.test-device-self"
var TEST_CM_DEVICE_SELF_RESOURCE = `
resource "bigip_cm_device_self" "test-device-self" {
	name                = "bigip1"
	configsync_ip       = "10.1.131.38"
	unicast_address {
		ip   = "10.1.131.38"
		port = 1026
		effective_ip   = "10.1.131.38"
		effective_port = 1026
	}
	unicast_address {
		effective_ip   = "management-ip"
		effective_port = 1026
		ip   = "management-ip"
		port = 1026
	}
}
`
var TEST_CM_DEVICE_SELF_RESOURCE_UPDATE = `
resource "bigip_cm_device_self" "test-device-self" {
	name                = "bigip1"
	mirror_ip           = "10.1.131.38"
	configsync_ip       = "10.1.131.38"
	multicast_ip        = "10.1.131.38"
	unicast_address {
		effective_ip   = "10.1.131.38"
		effective_port = 1026
		ip   = "10.1.131.38"
		port = 1026
	}
}
`

func TestAccBigipCmDeviceSelf_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_CM_DEVICE_SELF_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckCmDeviceSelfExists(true),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "name", "bigip1"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "configsync_ip", "10.1.131.38"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.#", "2"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.0.ip", "10.1.131.38"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.0.port", "1026"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.1.ip", "management-ip"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.1.port", "1026"),
				),
			},
		},
	})
}

func TestAccBigipCmDeviceSelf_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckCmDeviceSelfDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_CM_DEVICE_SELF_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "name", "bigip1"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "configsync_ip", "10.1.131.38"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.#", "2"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.0.ip", "10.1.131.38"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.0.port", "1026"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.1.ip", "management-ip"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.1.port", "1026"),
				),
			},
			{
				Config: TEST_CM_DEVICE_SELF_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "name", "bigip1"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "mirror_ip", "10.1.131.38"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "configsync_ip", "10.1.131.38"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.#", "1"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.0.ip", "10.1.131.38"),
					resource.TestCheckResourceAttr("bigip_cm_device_self.test-device-self", "unicast_address.0.port", "1026"),
				),
			},
		},
	})
}

func TestAccBigipCmDeviceSelf_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckCmDeviceSelfDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_CM_DEVICE_SELF_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckCmDeviceSelfExists(true),
				),
			},
			{
				ResourceName:      TEST_CM_DEVICE_SELF_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_CM_DEVICE_SELF_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckCmDeviceSelfExists(exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetDeviceSelf("bigip1")
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("CM Device Self configuration was not created.")
		}
		if !exists && p != nil {
			return fmt.Errorf("CM Device Self configuration still exists.")
		}
		return nil
	}
}

func testCheckCmDeviceSelfDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_cm_device_self" {
			continue
		}

		deviceSelf, err := client.GetDeviceSelf("bigip1")
		if err != nil {
			return err
		}

		var errors []string

		if deviceSelf.MirrorIp != "any6" {
			errors = append(errors, fmt.Sprintf("mirror_ip not reset to default (any6), got: %s", deviceSelf.MirrorIp))
		}
		if deviceSelf.MirrorSecondaryIp != "any6" {
			errors = append(errors, fmt.Sprintf("mirror_secondary_ip not reset to default (any6), got: %s", deviceSelf.MirrorSecondaryIp))
		}
		if deviceSelf.ConfigsyncIp != "none" {
			errors = append(errors, fmt.Sprintf("configsync_ip not reset to default (none), got: %s", deviceSelf.ConfigsyncIp))
		}
		if len(deviceSelf.UnicastAddress) != 0 {
			errors = append(errors, fmt.Sprintf("unicast_address not reset to default ([]), got: %v", deviceSelf.UnicastAddress))
		}

		if len(errors) > 0 {
			errorMsg := "CM Device Self configuration not properly reset to defaults:\n"
			for _, err := range errors {
				errorMsg += fmt.Sprintf("  - %s\n", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		fmt.Println("[INFO] CM Device Self configuration successfully reset to default values")
	}
	return nil
}
