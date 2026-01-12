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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBigipSysAuthSource() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the system-wide authentication source configuration on BIG-IP. " +
			"NOTE: Only one auth source resource should be created per BIG-IP device. " +
			"Multiple resources will reference the same underlying configuration and overwrite each other.",
		CreateContext: resourceBigipSysAuthSourceCreate,
		UpdateContext: resourceBigipSysAuthSourceUpdate,
		ReadContext:   resourceBigipSysAuthSourceRead,
		DeleteContext: resourceBigipSysAuthSourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"local",
					"ldap",
					"radius",
					"tacacs",
					"active-directory",
				}, false),
				Description: "Authentication source type (local, ldap, radius, tacacs, active-directory)",
			},
			"fallback": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "false",
				ValidateFunc: validation.StringInSlice([]string{"true", "false"}, false),
				Description:  "Fallback to local authentication if remote authentication fails (default: false)",
			},
		},
	}
}

func resourceBigipSysAuthSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Creating Authentication Source configuration")
	config := &bigip.AuthSource{
		Type:     d.Get("type").(string),
		Fallback: d.Get("fallback").(string),
	}

	if err := client.CreateAuthSource(config); err != nil {
		log.Printf("[ERROR] Unable to Create Auth Source: %v", err)
		return diag.FromErr(err)
	}

	d.SetId("auth_source")
	return resourceBigipSysAuthSourceRead(ctx, d, meta)
}

func resourceBigipSysAuthSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Updating Authentication Source configuration")
	config := &bigip.AuthSource{
		Type:     d.Get("type").(string),
		Fallback: d.Get("fallback").(string),
	}

	if err := client.ModifyAuthSource(config); err != nil {
		log.Printf("[ERROR] Unable to Modify Auth Source: %v", err)
		return diag.FromErr(err)
	}

	return resourceBigipSysAuthSourceRead(ctx, d, meta)
}

func resourceBigipSysAuthSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*bigip.BigIP)
	log.Printf("[INFO] Reading Authentication Source configuration")

	obj, err := client.GetAuthSource()
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Auth Source: %v", err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] Auth Source not found, removing from state")
		d.SetId("")
		return nil
	}

	if err := d.Set("type", obj.Type); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("fallback", obj.Fallback); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysAuthSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Resetting Authentication Source to default (local)")

	err := client.DeleteAuthSource()
	if err != nil {
		log.Printf("[ERROR] Unable to Reset Auth Source: %v", err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
