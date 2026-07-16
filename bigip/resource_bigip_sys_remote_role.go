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
	"math"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBigipSysRemoteRole() *schema.Resource {
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
		CreateContext: resourceBigipSysRemoteRoleCreate,
		UpdateContext: resourceBigipSysRemoteRoleUpdate,
		ReadContext:   resourceBigipSysRemoteRoleRead,
		DeleteContext: resourceBigipSysRemoteRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateF5Name,
				Description:  "Name of the remote role mapping",
			},
			"attribute": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute string used to match the remote group (e.g., 'memberof=cn=group,dc=example,dc=com')",
			},
			"line_order": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, math.MaxInt),
				Description:  "Order in which the remote role is processed (recommended starting at 1000)",
			},
			"role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validRoles, false),
				Description:  "BIG-IP role to assign to users matching this remote role",
			},
			"user_partition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "All",
				Description: "Partition to which the user is restricted (default: All)",
			},
			"console": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"disabled", "tmsh"}, false),
				Description:  "Console access level (disabled or tmsh, default: disabled)",
			},
			"deny": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Deny access to users matching this attribute (default: disabled)",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the remote role mapping",
			},
		},
	}
}

// buildRemoteRoleConfig constructs a RemoteRole config from ResourceData
func buildRemoteRoleConfig(d *schema.ResourceData) *bigip.RemoteRole {
	config := &bigip.RemoteRole{
		Name:      d.Get("name").(string),
		Attribute: d.Get("attribute").(string),
		LineOrder: d.Get("line_order").(int),
	}

	if val, ok := d.GetOk("role"); ok {
		config.Role = val.(string)
	}
	if val, ok := d.GetOk("user_partition"); ok {
		config.UserPartition = val.(string)
	}
	if val, ok := d.GetOk("console"); ok {
		config.Console = val.(string)
	}
	if val, ok := d.GetOk("deny"); ok {
		config.Deny = val.(string)
	}
	if val, ok := d.GetOk("description"); ok {
		config.Description = val.(string)
	}

	return config
}

func resourceBigipSysRemoteRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Creating Remote Role: %s", name)
	config := buildRemoteRoleConfig(d)

	if err := client.CreateRemoteRole(config); err != nil {
		log.Printf("[ERROR] Unable to Create Remote Role (%s): %v", name, err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipSysRemoteRoleRead(ctx, d, meta)
}

func resourceBigipSysRemoteRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	log.Printf("[INFO] Updating Remote Role: %s", name)
	config := buildRemoteRoleConfig(d)
	config.Name = name // Ensure we use the ID for updates

	if err := client.ModifyRemoteRole(name, config); err != nil {
		log.Printf("[ERROR] Unable to Modify Remote Role (%s): %v", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipSysRemoteRoleRead(ctx, d, meta)
}

func resourceBigipSysRemoteRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Reading Remote Role config: %s", name)

	obj, err := client.GetRemoteRole(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Remote Role (%s): %v", name, err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] Remote Role (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("name", obj.FullPath); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("attribute", obj.Attribute); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("line_order", obj.LineOrder); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("role", obj.Role); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("user_partition", obj.UserPartition); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("console", obj.Console); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("deny", obj.Deny); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", obj.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysRemoteRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Printf("[INFO] Deleting Remote Role: %s", name)

	err := client.DeleteRemoteRole(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete Remote Role (%s) (%v)", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
