/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2025 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/

package bigip

import (
	"context"
	"log"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBigipCmDeviceName() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the BIG-IP device name. This resource renames the local device.",
		CreateContext: resourceBigipCmDeviceNameCreate,
		UpdateContext: resourceBigipCmDeviceNameUpdate,
		ReadContext:   resourceBigipCmDeviceNameRead,
		DeleteContext: resourceBigipCmDeviceNameDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The desired device name for the BIG-IP device.",
			},
		},
	}
}

func resourceBigipCmDeviceNameCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Creating CM Device Name: %s", name)

	if err := client.CreateDeviceName(name); err != nil {
		log.Printf("[ERROR] Unable to Create CM Device Name: %v", err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipCmDeviceNameRead(ctx, d, meta)
}

func resourceBigipCmDeviceNameUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Updating CM Device Name: %s", name)

	if err := client.ModifyDeviceName(name); err != nil {
		log.Printf("[ERROR] Unable to Modify CM Device Name: %v", err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipCmDeviceNameRead(ctx, d, meta)
}

func resourceBigipCmDeviceNameRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	log.Printf("[INFO] Reading CM Device Name")

	name, err := client.GetSelfDeviceName()
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve CM Device Name: %v", err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBigipCmDeviceNameDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Resetting CM Device Name to default (bigip1)")

	if err := client.DeleteDeviceName(); err != nil {
		log.Printf("[ERROR] Unable to Reset CM Device Name: %v", err)
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
