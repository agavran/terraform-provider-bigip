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

func resourceBigipSysRemoteUser() *schema.Resource {
	validRoles := []string{
		"admin",
		"application-editor",
		"auditor",
		"certificate-manager",
		"firewall-manager",
		"fraud-protection-manager",
		"guest",
		"irule-manager",
		"manager",
		"no-access",
		"operator",
		"resource-admin",
		"user-manager",
		"web-application-security-administrator",
		"web-application-security-editor",
	}

	return &schema.Resource{
		Description: "Manages the system-wide external (remote) user configuration on BIG-IP. " +
			"NOTE: Only one remote user resource should be created per BIG-IP device. " +
			"Multiple resources will reference the same underlying configuration and overwrite each other.",
		CreateContext: resourceBigipSysRemoteUserCreate,
		UpdateContext: resourceBigipSysRemoteUserUpdate,
		ReadContext:   resourceBigipSysRemoteUserRead,
		DeleteContext: resourceBigipSysRemoteUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"default_partition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "all",
				Description: "Default partition for remote users (default: all)",
			},
			"default_role": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "no-access",
				ValidateFunc: validation.StringInSlice(validRoles, false),
				Description:  "Default role for remote users (default: no-access)",
			},
			"remote_console_access": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"disabled", "tmsh"}, false),
				Description:  "Console access for remote users (disabled or tmsh, default: disabled)",
			},
		},
	}
}

func resourceBigipSysRemoteUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Creating Remote User configuration")
	config := &bigip.RemoteUser{
		DefaultPartition:    d.Get("default_partition").(string),
		DefaultRole:         d.Get("default_role").(string),
		RemoteConsoleAccess: d.Get("remote_console_access").(string),
	}

	if err := client.CreateRemoteUser(config); err != nil {
		log.Printf("[ERROR] Unable to Create Remote User Config: %v", err)
		return diag.FromErr(err)
	}

	d.SetId("remote_user")
	return resourceBigipSysRemoteUserRead(ctx, d, meta)
}

func resourceBigipSysRemoteUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Updating Remote User configuration")
	config := &bigip.RemoteUser{
		DefaultPartition:    d.Get("default_partition").(string),
		DefaultRole:         d.Get("default_role").(string),
		RemoteConsoleAccess: d.Get("remote_console_access").(string),
	}

	if err := client.ModifyRemoteUser(config); err != nil {
		log.Printf("[ERROR] Unable to Modify Remote User Config: %v", err)
		return diag.FromErr(err)
	}

	return resourceBigipSysRemoteUserRead(ctx, d, meta)
}

func resourceBigipSysRemoteUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*bigip.BigIP)
	log.Printf("[INFO] Reading Remote User configuration")

	obj, err := client.GetRemoteUser()
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Remote User Config: %v", err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] Remote User Config not found, removing from state")
		d.SetId("")
		return nil
	}

	if err := d.Set("default_partition", obj.DefaultPartition); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("default_role", obj.DefaultRole); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("remote_console_access", obj.RemoteConsoleAccess); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysRemoteUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	log.Printf("[INFO] Resetting Remote User configuration to defaults")

	err := client.DeleteRemoteUser()
	if err != nil {
		log.Printf("[ERROR] Unable to Reset Remote User Config: %v", err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
