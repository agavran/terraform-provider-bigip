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

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// validateSnmpCommunityOidSource validates that source cannot be "all" when oid_subset is defined
func validateSnmpCommunityOidSource(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	oidSubset, oidOk := diff.GetOk("oid_subset")
	source, sourceOk := diff.GetOk("source")

	if oidOk && oidSubset.(string) != "" && sourceOk && source.(string) == "all" {
		return fmt.Errorf("source cannot be \"all\" when oid_subset is defined")
	}

	return nil
}

func resourceBigipSysSnmpCommunity() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipSysSnmpCommunityCreate,
		UpdateContext: resourceBigipSysSnmpCommunityUpdate,
		ReadContext:   resourceBigipSysSnmpCommunityRead,
		DeleteContext: resourceBigipSysSnmpCommunityDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateSnmpCommunityOidSource,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateF5Name,
				Description:  "Name of the SNMP community",
			},
			"community_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SNMP community string",
			},
			"access": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ro",
				ValidateFunc: validation.StringInSlice([]string{"ro", "rw", "roa", "rwa"}, false),
				Description:  "Access level (ro=read-only, rw=read-write, roa=read-only-all, rwa=read-write-all, default: ro)",
			},
			"source": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "all",
				Description: "Source IP address or network (CIDR notation) allowed to use this community (default: all)",
			},
			"ipv6": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable IPv6 support (default: disabled)",
			},
			"oid_subset": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "OID subset for restricting SNMP access",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the SNMP community",
			},
		},
	}
}

// buildSnmpCommunityConfig constructs an SnmpCommunity config from ResourceData
func buildSnmpCommunityConfig(d *schema.ResourceData) *bigip.SnmpCommunity {
	config := &bigip.SnmpCommunity{
		Name:          d.Get("name").(string),
		CommunityName: d.Get("community_name").(string),
	}

	if val, ok := d.GetOk("access"); ok {
		config.Access = val.(string)
	}
	if val, ok := d.GetOk("source"); ok {
		config.Source = val.(string)
	}
	if val, ok := d.GetOk("ipv6"); ok {
		config.Ipv6 = val.(string)
	}
	if val, ok := d.GetOk("oid_subset"); ok {
		config.OidSubset = val.(string)
	}
	if val, ok := d.GetOk("description"); ok {
		config.Description = val.(string)
	}

	return config
}

func resourceBigipSysSnmpCommunityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Creating SNMP Community: %s", name)
	config := buildSnmpCommunityConfig(d)

	if err := client.CreateSnmpCommunity(config); err != nil {
		log.Printf("[ERROR] Unable to Create SNMP Community (%s): %v", name, err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipSysSnmpCommunityRead(ctx, d, meta)
}

func resourceBigipSysSnmpCommunityUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	log.Printf("[INFO] Updating SNMP Community: %s", name)
	config := buildSnmpCommunityConfig(d)
	config.Name = name

	if err := client.ModifySnmpCommunity(name, config); err != nil {
		log.Printf("[ERROR] Unable to Modify SNMP Community (%s): %v", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipSysSnmpCommunityRead(ctx, d, meta)
}

func resourceBigipSysSnmpCommunityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Reading SNMP Community: %s", name)

	obj, err := client.GetSnmpCommunity(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve SNMP Community (%s): %v", name, err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] SNMP Community (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("name", obj.FullPath); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("community_name", obj.CommunityName); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("access", obj.Access); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("source", obj.Source); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("ipv6", obj.Ipv6); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("oid_subset", obj.OidSubset); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", obj.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysSnmpCommunityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Printf("[INFO] Deleting SNMP Community: %s", name)

	err := client.DeleteSnmpCommunity(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete SNMP Community (%s) (%v)", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
