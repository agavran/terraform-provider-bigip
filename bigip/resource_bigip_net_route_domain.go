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

func resourceBigipNetRouteDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipNetRouteDomainCreate,
		UpdateContext: resourceBigipNetRouteDomainUpdate,
		ReadContext:   resourceBigipNetRouteDomainRead,
		DeleteContext: resourceBigipNetRouteDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateF5Name,
				Description:  "Name of the route domain (e.g., /Common/100)",
			},
			"rd_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Route domain ID (must be unique, 0-65534)",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the route domain",
			},
			"strict": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "enabled",
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
				Description:  "Enable strict isolation (default: enabled)",
			},
			"parent": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Parent route domain (example: /Common/0)",
			},
			"vlans": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of VLANs associated with this route domain",
			},
			"routing_protocol": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"BFD", "BGP", "IS-IS", "OSPFv2", "OSPFv3", "PIM", "RIP", "RIPng", "none"}, false),
				},
				Description: "Set of routing protocols (BFD, BGP, IS-IS, OSPFv2, OSPFv3, PIM, RIP, RIPng, none)",
			},
			"bwc_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Bandwidth control policy",
			},
			"connection_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Maximum number of concurrent connections (0 = unlimited)",
			},
			"flow_eviction_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Flow eviction policy",
			},
			"fw_enforced_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Enforced firewall policy",
			},
			"fw_staged_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Staged firewall policy",
			},
			"ip_intelligence_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP intelligence policy",
			},
			"security_nat_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Security NAT policy",
			},
			"security_packet_filter_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Security packet filter policy",
			},
			"service_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Service policy",
			},
			"throughput_capacity": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "infinite",
				Description: "Specifies the max throughput capacity of this route domain in Mbps (default: infinite)",
			},
		},
	}
}

// buildRouteDomainConfig constructs a RouteDomain config from ResourceData
func buildRouteDomainConfig(d *schema.ResourceData) *bigip.RouteDomain {
	config := &bigip.RouteDomain{
		Name: d.Get("name").(string),
		ID:   d.Get("rd_id").(int),
	}

	setStringIfOk(d, "description", &config.Description)
	setStringIfOk(d, "strict", &config.Strict)
	setStringIfOk(d, "parent", &config.Parent)
	setStringIfOk(d, "bwc_policy", &config.BwcPolicy)
	setStringIfOk(d, "flow_eviction_policy", &config.FlowEvictionPolicy)
	setStringIfOk(d, "fw_enforced_policy", &config.FwEnforcedPolicy)
	setStringIfOk(d, "fw_staged_policy", &config.FwStagedPolicy)
	setStringIfOk(d, "ip_intelligence_policy", &config.IpIntelligencePolicy)
	setStringIfOk(d, "security_nat_policy", &config.SecurityNatPolicy)
	setStringIfOk(d, "security_packet_filter_policy", &config.SecurityPacketFilterPolicy)
	setStringIfOk(d, "service_policy", &config.ServicePolicy)
	setStringIfOk(d, "throughput_capacity", &config.ThroughputCapacity)

	if val, ok := d.GetOk("connection_limit"); ok {
		config.ConnectionLimit = val.(int)
	}

	if val, ok := d.GetOk("vlans"); ok {
		config.Vlans = setToStringSlice(val.(*schema.Set))
	}
	if val, ok := d.GetOk("routing_protocol"); ok {
		config.RoutingProtocol = setToStringSlice(val.(*schema.Set))
	}

	return config
}

func resourceBigipNetRouteDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Creating Route Domain: %s", name)
	config := buildRouteDomainConfig(d)

	if err := client.CreateRouteDomain(config); err != nil {
		log.Printf("[ERROR] Unable to Create Route Domain (%s): %v", name, err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipNetRouteDomainRead(ctx, d, meta)
}

func resourceBigipNetRouteDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	log.Printf("[INFO] Updating Route Domain: %s", name)
	config := buildRouteDomainConfig(d)

	if d.HasChange("vlans") {
		if val, ok := d.GetOk("vlans"); !ok || val.(*schema.Set).Len() == 0 {
			log.Printf("[INFO] VLANs removed from config, clearing them")
			config.Vlans = []string{}
		}
	}
	if d.HasChange("routing_protocol") {
		if val, ok := d.GetOk("routing_protocol"); !ok || val.(*schema.Set).Len() == 0 {
			log.Printf("[INFO] Routing protocols removed from config, clearing them")
			config.RoutingProtocol = []string{}
		}
	}

	if err := client.ModifyRouteDomain(name, config); err != nil {
		log.Printf("[ERROR] Unable to Modify Route Domain (%s): %v", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipNetRouteDomainRead(ctx, d, meta)
}

func resourceBigipNetRouteDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Reading Route Domain: %s", name)

	obj, err := client.GetRouteDomain(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve Route Domain (%s): %v", name, err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] Route Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	setResourceData(d, "name", obj.FullPath, &diags)
	setResourceData(d, "rd_id", obj.ID, &diags)
	setResourceData(d, "description", obj.Description, &diags)
	setResourceData(d, "strict", obj.Strict, &diags)
	setResourceData(d, "parent", obj.Parent, &diags)
	setResourceData(d, "vlans", obj.Vlans, &diags)
	setResourceData(d, "routing_protocol", obj.RoutingProtocol, &diags)
	setResourceData(d, "bwc_policy", obj.BwcPolicy, &diags)
	setResourceData(d, "connection_limit", obj.ConnectionLimit, &diags)
	setResourceData(d, "flow_eviction_policy", obj.FlowEvictionPolicy, &diags)
	setResourceData(d, "fw_enforced_policy", obj.FwEnforcedPolicy, &diags)
	setResourceData(d, "fw_staged_policy", obj.FwStagedPolicy, &diags)
	setResourceData(d, "ip_intelligence_policy", obj.IpIntelligencePolicy, &diags)
	setResourceData(d, "security_nat_policy", obj.SecurityNatPolicy, &diags)
	setResourceData(d, "security_packet_filter_policy", obj.SecurityPacketFilterPolicy, &diags)
	setResourceData(d, "service_policy", obj.ServicePolicy, &diags)
	setResourceData(d, "throughput_capacity", obj.ThroughputCapacity, &diags)

	return diags
}

func resourceBigipNetRouteDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Printf("[INFO] Deleting Route Domain: %s", name)

	if name == "/Common/0" || d.Get("rd_id").(int) == 0 {
		return diag.FromErr(fmt.Errorf("cannot delete default route domain 0"))
	}

	err := client.DeleteRouteDomain(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete Route Domain (%s) (%v)", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
