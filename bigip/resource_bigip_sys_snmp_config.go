/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2025 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/

package bigip

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// defaultAllowSnmpConfig returns the default value for the allowed_addresses field
func defaultAllowSnmpConfig() (interface{}, error) {
	return []interface{}{"127.0.0.0/8"}, nil
}

// validateAllowedAddresses validates that allowed_addresses contains valid IP addresses or networks
// Accepts: single IPs (192.168.1.1), CIDR notation (192.168.0.0/16), or IP/netmask pairs (192.168.0.0/255.255.0.0)
func validateAllowedSnmpAddresses(val interface{}, key string) (warns []string, errs []error) {
	list, ok := val.([]interface{})
	if !ok {
		errs = append(errs, fmt.Errorf("%q must be a list", key))
		return
	}

	if len(list) == 0 {
		return
	}

	for i, item := range list {
		str, ok := item.(string)
		if !ok {
			errs = append(errs, fmt.Errorf("%q[%d]: must be a string", key, i))
			continue
		}

		if strings.Contains(str, "/") {
			parts := strings.Split(str, "/")
			if len(parts) != 2 {
				errs = append(errs, fmt.Errorf("%q[%d]: invalid format %q, expected IP/mask or IP/CIDR", key, i, str))
				continue
			}

			if net.ParseIP(parts[0]) == nil {
				errs = append(errs, fmt.Errorf("%q[%d]: invalid IP address %q", key, i, parts[0]))
				continue
			}

			if _, _, err := net.ParseCIDR(str); err == nil {
				continue
			}

			if net.ParseIP(parts[1]) == nil {
				errs = append(errs, fmt.Errorf("%q[%d]: invalid CIDR or netmask %q", key, i, str))
			}
		} else {
			if net.ParseIP(str) == nil {
				errs = append(errs, fmt.Errorf("%q[%d]: invalid IP address %q", key, i, str))
			}
		}
	}

	return
}

// validateAllowedAddressesDiff validates the allowed_addresses during plan phase using CustomizeDiff
func validateAllowedSnmpAddressesDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if val, ok := diff.GetOk("allowed_addresses"); ok {
		_, errs := validateAllowedSnmpAddresses(val, "allowed_addresses")
		if len(errs) > 0 {
			if len(errs) == 1 {
				return errs[0]
			}
			errMsg := "Multiple validation errors for \"allowed_addresses\" field:\n"
			for i, err := range errs {
				errMsg += fmt.Sprintf("  %d. %s\n", i+1, err.Error())
			}
			return fmt.Errorf("%s", errMsg)
		}
	}
	return nil
}

func resourceBigipSysSnmpConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipSysSnmpConfigCreate,
		UpdateContext: resourceBigipSysSnmpConfigUpdate,
		ReadContext:   resourceBigipSysSnmpConfigRead,
		DeleteContext: resourceBigipSysSnmpConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateAllowedSnmpAddressesDiff,
		Schema: map[string]*schema.Schema{
			"sys_contact": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Customer Name <admin@customer.com>",
				Description: "SNMP system contact information (default: 'Customer Name <admin@customer.com>')",
			},
			"sys_location": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Network Closet 1",
				Description: "SNMP system location (default: 'Network Closet 1')",
			},
			"allowed_addresses": {
				Type:        schema.TypeList,
				Optional:    true,
				DefaultFunc: defaultAllowSnmpConfig,
				Description: "List of allowed IP addresses or networks for SNMP access (single IP address, CIDR or network mask notation)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// buildSnmpConfig constructs an SnmpConfig from ResourceData
func buildSnmpConfig(d *schema.ResourceData) *bigip.SnmpConfig {
	config := &bigip.SnmpConfig{}

	if val, ok := d.GetOk("sys_contact"); ok {
		config.SysContact = val.(string)
	}
	if val, ok := d.GetOk("sys_location"); ok {
		config.SysLocation = val.(string)
	}
	if val, ok := d.GetOk("allowed_addresses"); ok {
		addresses := val.([]interface{})
		config.AllowedAddresses = make([]string, len(addresses))
		for i, v := range addresses {
			config.AllowedAddresses[i] = v.(string)
		}
	}

	return config
}

func resourceBigipSysSnmpConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Creating SNMP Configuration")
	config := buildSnmpConfig(d)

	if err := client.CreateSnmpConfig(config); err != nil {
		log.Printf("[ERROR] Unable to Create SNMP Config: %v", err)
		return diag.FromErr(err)
	}

	d.SetId("snmp_config")
	return resourceBigipSysSnmpConfigRead(ctx, d, meta)
}

func resourceBigipSysSnmpConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Updating SNMP Configuration")
	config := buildSnmpConfig(d)

	if err := client.ModifySnmpConfig(config); err != nil {
		log.Printf("[ERROR] Unable to Modify SNMP Config: %v", err)
		return diag.FromErr(err)
	}

	return resourceBigipSysSnmpConfigRead(ctx, d, meta)
}

func resourceBigipSysSnmpConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	log.Printf("[INFO] Reading SNMP Configuration")

	obj, err := client.GetSnmpConfig()
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve SNMP Config: %v", err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] SNMP Config not found, removing from state")
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("sys_contact", obj.SysContact); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("sys_location", obj.SysLocation); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("allowed_addresses", obj.AllowedAddresses); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysSnmpConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Resetting SNMP Configuration to defaults")

	err := client.DeleteSnmpConfig()
	if err != nil {
		log.Printf("[ERROR] Unable to Reset SNMP Config: %v", err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
