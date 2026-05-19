/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2019 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package bigip

import (
	"context"
	"fmt"
	"log"
	"strings"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBigipNetVlan() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceBigipNetVlanCreate,
		ReadContext:   resourceBigipNetVlanRead,
		UpdateContext: resourceBigipNetVlanUpdate,
		DeleteContext: resourceBigipNetVlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the VLAN",
			},

			"tag": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "VLAN ID (tag)",
			},

			"interfaces": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Interface(s) attached to the VLAN",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vlanport": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Vlan name",
						},
						"tagged": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Interface tagged",
						},
					},
				},
			},
			"mtu": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Maximum Transmission Unit (MTU) for the VLAN",
				Default:      1500,
				ValidateFunc: validation.IntBetween(576, 9198),
			},
			"cmp_hash": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"default", "src-ip", "dst-ip"}, false),
				Description:  "Specifies how the traffic on the VLAN will be disaggregated. The value selected determines the traffic disaggregation method",
			},
		},
	}

}

func resourceBigipNetVlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Get("name").(string)
	tag := d.Get("tag").(int)
	mtu := d.Get("mtu").(int)

	log.Printf("[INFO] Creating VLAN %s", name)

	marketingName, err := client.GetSelfDeviceMarketingName()
	if err != nil {
		return diag.FromErr(fmt.Errorf("error detecting device type: %v", err))
	}
	log.Printf("[DEBUG] VLAN create: detected device marketingName=%q", marketingName)
	if marketingName == "BIG-IP Tenant" {
		existing, err := client.Vlan(name)
		if err != nil && !strings.Contains(err.Error(), fmt.Sprintf("The requested VLAN (%s) was not found", name)) {
			return diag.FromErr(fmt.Errorf("error checking if VLAN %s exists on tenant: %v", name, err))
		}
		if existing != nil {
			log.Printf("[INFO] VLAN %s already exists on tenant, skipping creation and reading state", name)
			d.SetId(name)
			return resourceBigipNetVlanRead(ctx, d, meta)
		}
		log.Printf("[INFO] VLAN %s does not exist on tenant, proceeding with creation", name)
	}

	d.Partial(true)

	r := &bigip.Vlan{
		Name:    name,
		Tag:     tag,
		MTU:     mtu,
		CMPHash: d.Get("cmp_hash").(string),
	}

	err = client.CreateVlan(r)

	if err != nil {
		return diag.FromErr(fmt.Errorf("Error creating VLAN %s: %v ", name, err))
	}

	d.SetId(name)

	ifaceCount := d.Get("interfaces.#").(int)
	for i := 0; i < ifaceCount; i++ {
		prefix := fmt.Sprintf("interfaces.%d", i)
		iface := d.Get(prefix + ".vlanport").(string)
		tagged := d.Get(prefix + ".tagged").(bool)

		err = client.AddInterfaceToVlan(name, iface, tagged)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error adding Interface %s to VLAN %s: %v", iface, name, err))
		}
	}

	d.Partial(false)

	return resourceBigipNetVlanRead(ctx, d, meta)
}

func resourceBigipNetVlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	log.Printf("[INFO] Reading VLAN %s", name)

	vlan, err := client.Vlan(name)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving VLAN %s: %v", name, err))
	}
	if vlan == nil {
		log.Printf("[DEBUG] VLAN %s not found, removing from state", name)
		d.SetId("")
		return nil
	}

	_ = d.Set("name", vlan.FullPath)
	_ = d.Set("tag", vlan.Tag)
	_ = d.Set("cmp_hash", vlan.CMPHash)
	_ = d.Set("mtu", vlan.MTU)

	log.Printf("[DEBUG] Reading VLAN %s Interfaces", name)

	vlanInterfaces, err := client.GetVlanInterfaces(name)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving VLAN %s Interfaces: %v", name, err))
	}

	var interfaces []map[string]interface{}
	var ifaceTagged bool
	for _, iface := range vlanInterfaces.VlanInterfaces {
		if iface.Tagged {
			ifaceTagged = true
		} else {
			ifaceTagged = false
		}
		log.Printf("[DEBUG] Retrieved VLAN Interface %s, tagging is set to %t", iface.Name, ifaceTagged)

		vlanIface := map[string]interface{}{
			"vlanport": iface.Name,
			"tagged":   ifaceTagged,
		}

		interfaces = append(interfaces, vlanIface)
	}

	if err := d.Set("interfaces", interfaces); err != nil {
		return diag.FromErr(fmt.Errorf("error updating Interfaces in state for VLAN %s: %v", name, err))
	}

	return nil
}

func resourceBigipNetVlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	log.Printf("[INFO] Updating VLAN %s", name)

	marketingName, err := client.GetSelfDeviceMarketingName()
	if err != nil {
		return diag.FromErr(fmt.Errorf("error detecting device type: %v", err))
	}
	if marketingName == "BIG-IP Tenant" {
		if d.HasChange("tag") || d.HasChange("interfaces") {
			return diag.FromErr(fmt.Errorf("VLAN %s: tag and interfaces cannot be modified on a BIG-IP Tenant, these are managed by the host", name))
		}
	}

	r := &bigip.Vlan{
		Name:    name,
		Tag:     d.Get("tag").(int),
		MTU:     d.Get("mtu").(int),
		CMPHash: d.Get("cmp_hash").(string),
	}

	err = client.ModifyVlan(name, r)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error modifying VLAN %s: %v", name, err))
	}

	return resourceBigipNetVlanRead(ctx, d, meta)
}

func resourceBigipNetVlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	log.Printf("[INFO] Deleting VLAN %s", name)

	err := client.DeleteVlan(name)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error Deleting Vlan : %s", err))
	}
	d.SetId("")
	return nil
}
