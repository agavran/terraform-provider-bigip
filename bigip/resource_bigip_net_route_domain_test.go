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

var TEST_ROUTE_DOMAIN_NAME = "/Common/101"
var TEST_ROUTE_DOMAIN_RESOURCE_NAME = "bigip_net_route_domain.test-rd"
var TEST_ROUTE_DOMAIN_RESOURCE = `
resource "bigip_net_route_domain" "test-rd" {
	name              = "` + TEST_ROUTE_DOMAIN_NAME + `"
	rd_id             = 101
	description       = "Test route domain"
	strict            = "enabled"
	parent            = "/Common/0"
	connection_limit  = 1000
}
`

var TEST_ROUTE_DOMAIN_RESOURCE_UPDATE = `
resource "bigip_net_route_domain" "test-rd" {
	name              = "` + TEST_ROUTE_DOMAIN_NAME + `"
	rd_id             = 101
	description       = "Updated test route domain"
	strict            = "disabled"
	parent            = "/Common/0"
	vlans             = ["/Common/socks-tunnel", "/Common/http-tunnel"]
	connection_limit  = 2000
	routing_protocol  = ["BGP", "BFD"]
}
`

func TestAccBigipNetRouteDomain_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TEST_ROUTE_DOMAIN_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckRouteDomainExists(TEST_ROUTE_DOMAIN_NAME, true),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "name", "/Common/101"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "rd_id", "101"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "description", "Test route domain"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "strict", "enabled"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "parent", "/Common/0"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "connection_limit", "1000"),
				),
			},
		},
	})
}

func TestAccBigipNetRouteDomain_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckRouteDomainDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_ROUTE_DOMAIN_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "description", "Test route domain"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "strict", "enabled"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "connection_limit", "1000"),
				),
			},
			{
				Config: TEST_ROUTE_DOMAIN_RESOURCE_UPDATE,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "description", "Updated test route domain"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "strict", "disabled"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "vlans.#", "2"),
					resource.TestCheckTypeSetElemAttr("bigip_net_route_domain.test-rd", "vlans.*", "/Common/socks-tunnel"),
					resource.TestCheckTypeSetElemAttr("bigip_net_route_domain.test-rd", "vlans.*", "/Common/http-tunnel"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "connection_limit", "2000"),
					resource.TestCheckResourceAttr("bigip_net_route_domain.test-rd", "routing_protocol.#", "2"),
					resource.TestCheckTypeSetElemAttr("bigip_net_route_domain.test-rd", "routing_protocol.*", "BGP"),
					resource.TestCheckTypeSetElemAttr("bigip_net_route_domain.test-rd", "routing_protocol.*", "BFD"),
				),
			},
		},
	})
}

func TestAccBigipNetRouteDomain_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckRouteDomainDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TEST_ROUTE_DOMAIN_RESOURCE,
				Check: resource.ComposeTestCheckFunc(
					testCheckRouteDomainExists(TEST_ROUTE_DOMAIN_NAME, true),
				),
			},
			{
				ResourceName:      TEST_ROUTE_DOMAIN_RESOURCE_NAME,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     TEST_ROUTE_DOMAIN_NAME,
				ImportStateCheck:  testPrintImportedState,
			},
		},
	})
}

func testCheckRouteDomainExists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		p, err := client.GetRouteDomain(name)
		if err != nil {
			return err
		}
		if exists && p == nil {
			return fmt.Errorf("Route Domain %s was not created.", name)
		}
		if !exists && p != nil {
			return fmt.Errorf("Route Domain %s still exists.", name)
		}
		return nil
	}
}

func testCheckRouteDomainDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_net_route_domain" {
			continue
		}

		name := rs.Primary.ID
		routeDomain, err := client.GetRouteDomain(name)
		if err != nil {
			return err
		}
		if routeDomain != nil {
			return fmt.Errorf("Route Domain %s not destroyed.", name)
		}
		fmt.Printf("[INFO] Route Domain '%s' has been successfully destroyed\n", name)
	}
	return nil
}
