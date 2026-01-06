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

func resourceBigipSysHaGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipSysHaGroupCreate,
		UpdateContext: resourceBigipSysHaGroupUpdate,
		ReadContext:   resourceBigipSysHaGroupRead,
		DeleteContext: resourceBigipSysHaGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateF5NameNoPartition,
				Description:  "Name of the HA Group (e.g., ha_group1)",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the HA Group",
			},
			"active_bonus": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      10,
				ValidateFunc: validation.IntBetween(0, 100),
				Description:  "Bonus score for the active device (0-100, default: 0)",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable or disable the HA Group (default: true)",
			},
			"pools": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the pool to monitor",
						},
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntBetween(10, 100),
							Description:  "Weight of this pool in HA scoring (10-100, default: 10)",
						},
						"attribute": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "percent-up-members",
							ValidateFunc: validation.StringInSlice([]string{"percent-up-members"}, false),
							Description:  "Attribute to monitor (default: percent-up-members). Percent-up-members is the only available attribute for HA scoring for the clusters, pools, and trunks options.",
						},
						"minimum_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntAtLeast(0),
							Description:  "Minimum number of up members required for this component to contribute to HA score (default: 0). Value may not exceed the actual number of members in the referenced resource.",
						},
						"percent_up": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(0, 100),
							Description:  "Percentage of pool members that must be up",
						},
						"sufficient_threshold": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "all",
							Description: "Sufficient number of up members above which this component is considered at 100% for HA score contribution (default: all). Value may not exceed the actual number of members in the referenced resource.",
						},
					},
				},
				Description: "List of pools to monitor for HA scoring",
			},
			"clusters": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the cluster to monitor",
						},
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntBetween(0, 100),
							Description:  "Weight of this cluster in HA scoring (0-100, default: 10)",
						},
						"attribute": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "percent-up-members",
							ValidateFunc: validation.StringInSlice([]string{"percent-up-members"}, false),
							Description:  "Attribute to monitor (default: percent-up-members). Percent-up-members is the only available attribute for HA scoring for the clusters, pools, and trunks options.",
						},
						"minimum_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntAtLeast(0),
							Description:  "Minimum number of up members required for this component to contribute to HA score (default: 0). Value may not exceed the actual number of members in the referenced resource.",
						},
						"percent_up": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(0, 100),
							Description:  "Percentage of cluster members that must be up",
						},
						"sufficient_threshold": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "all",
							Description: "Sufficient number of up members above which this component is considered at 100% for HA score contribution (default: all). Value may not exceed the actual number of members in the referenced resource.",
						},
					},
				},
				Description: "List of clusters to monitor for HA scoring",
			},
			"trunks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the trunk to monitor",
						},
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntBetween(0, 100),
							Description:  "Weight of this trunk in HA scoring (0-100, default: 10)",
						},
						"attribute": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "percent-up-members",
							ValidateFunc: validation.StringInSlice([]string{"percent-up-members"}, false),
							Description:  "Attribute to monitor (default: percent-up-members). Percent-up-members is the only available attribute for HA scoring for the clusters, pools, and trunks options.",
						},
						"minimum_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntAtLeast(0),
							Description:  "Minimum number of up members required for this component to contribute to HA score (default: 0). Value may not exceed the actual number of members in the referenced resource.",
						},
						"percent_up": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(0, 100),
							Description:  "Percentage of trunk members that must be up",
						},
						"sufficient_threshold": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "all",
							Description: "Sufficient number of up members above which this component is considered at 100% for HA score contribution (default: all). Value may not exceed the actual number of members in the referenced resource.",
						},
					},
				},
				Description: "List of trunks to monitor for HA scoring",
			},
		},
	}
}

// buildHaGroupConfig constructs an HaGroup config from ResourceData
func buildHaGroupConfig(d *schema.ResourceData) *bigip.HaGroup {
	config := &bigip.HaGroup{
		Name: d.Get("name").(string),
	}

	if val, ok := d.GetOk("description"); ok {
		config.Description = val.(string)
	}
	if val, ok := d.GetOk("active_bonus"); ok {
		config.ActiveBonus = val.(int)
	}

	// Enable/Disable needs special handling
	// needs to be sent in pair with opposite values: Enable=true Disable=false / Enable=false Disable=true
	config.Enabled = d.Get("enabled").(bool)
	config.Disabled = !config.Enabled

	if val, ok := d.GetOk("pools"); ok {
		poolsList := val.([]interface{})
		pools := make([]bigip.HaGroupPool, len(poolsList))
		for i, poolInterface := range poolsList {
			pool := poolInterface.(map[string]interface{})
			haPool := bigip.HaGroupPool{
				Name: pool["name"].(string),
			}
			if weight, ok := pool["weight"].(int); ok {
				haPool.Weight = weight
			}
			if attr, ok := pool["attribute"].(string); ok && attr != "" {
				haPool.Attribute = attr
			}
			if minThresh, ok := pool["minimum_threshold"].(int); ok {
				haPool.MinimumThreshold = minThresh
			}
			if percentUp, ok := pool["percent_up"].(int); ok && percentUp > 0 {
				haPool.PercentUp = percentUp
			}
			if suffThresh, ok := pool["sufficient_threshold"].(string); ok && suffThresh != "" {
				haPool.SufficientThreshold = suffThresh
			}
			pools[i] = haPool
		}
		config.Pools = pools
	}

	if val, ok := d.GetOk("clusters"); ok {
		clustersList := val.([]interface{})
		clusters := make([]bigip.HaGroupCluster, len(clustersList))
		for i, clusterInterface := range clustersList {
			cluster := clusterInterface.(map[string]interface{})
			haCluster := bigip.HaGroupCluster{
				Name: cluster["name"].(string),
			}
			if weight, ok := cluster["weight"].(int); ok {
				haCluster.Weight = weight
			}
			if attr, ok := cluster["attribute"].(string); ok && attr != "" {
				haCluster.Attribute = attr
			}
			if minThresh, ok := cluster["minimum_threshold"].(int); ok {
				haCluster.MinimumThreshold = minThresh
			}
			if percentUp, ok := cluster["percent_up"].(int); ok && percentUp > 0 {
				haCluster.PercentUp = percentUp
			}
			if suffThresh, ok := cluster["sufficient_threshold"].(string); ok && suffThresh != "" {
				haCluster.SufficientThreshold = suffThresh
			}
			clusters[i] = haCluster
		}
		config.Clusters = clusters
	}

	if val, ok := d.GetOk("trunks"); ok {
		trunksList := val.([]interface{})
		trunks := make([]bigip.HaGroupTrunk, len(trunksList))
		for i, trunkInterface := range trunksList {
			trunk := trunkInterface.(map[string]interface{})
			haTrunk := bigip.HaGroupTrunk{
				Name: trunk["name"].(string),
			}
			if weight, ok := trunk["weight"].(int); ok {
				haTrunk.Weight = weight
			}
			if attr, ok := trunk["attribute"].(string); ok && attr != "" {
				haTrunk.Attribute = attr
			}
			if minThresh, ok := trunk["minimum_threshold"].(int); ok {
				haTrunk.MinimumThreshold = minThresh
			}
			if percentUp, ok := trunk["percent_up"].(int); ok && percentUp > 0 {
				haTrunk.PercentUp = percentUp
			}
			if suffThresh, ok := trunk["sufficient_threshold"].(string); ok && suffThresh != "" {
				haTrunk.SufficientThreshold = suffThresh
			}
			trunks[i] = haTrunk
		}
		config.Trunks = trunks
	}

	return config
}

func resourceBigipSysHaGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Printf("[INFO] Creating HA Group: %s", name)
	config := buildHaGroupConfig(d)

	if err := client.CreateHaGroup(config); err != nil {
		log.Printf("[ERROR] Unable to Create HA Group (%s): %v", name, err)
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceBigipSysHaGroupRead(ctx, d, meta)
}

func resourceBigipSysHaGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	log.Printf("[INFO] Updating HA Group: %s", name)
	config := buildHaGroupConfig(d)

	if d.HasChange("pools") {
		if val, ok := d.GetOk("pools"); !ok || len(val.([]interface{})) == 0 {
			log.Printf("[INFO] Pools removed from config, clearing them")
			config.Pools = []bigip.HaGroupPool{}
		}
	}
	if d.HasChange("clusters") {
		if val, ok := d.GetOk("clusters"); !ok || len(val.([]interface{})) == 0 {
			log.Printf("[INFO] Clusters removed from config, clearing them")
			config.Clusters = []bigip.HaGroupCluster{}
		}
	}
	if d.HasChange("trunks") {
		if val, ok := d.GetOk("trunks"); !ok || len(val.([]interface{})) == 0 {
			log.Printf("[INFO] Trunks removed from config, clearing them")
			config.Trunks = []bigip.HaGroupTrunk{}
		}
	}

	if err := client.ModifyHaGroup(name, config); err != nil {
		log.Printf("[ERROR] Unable to Modify HA Group (%s): %v", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipSysHaGroupRead(ctx, d, meta)
}

func resourceBigipSysHaGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	log.Printf("[INFO] Reading HA Group: %s", name)

	obj, err := client.GetHaGroup(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Retrieve HA Group (%s): %v", name, err)
		return diag.FromErr(err)
	}
	if obj == nil {
		log.Printf("[WARN] HA Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics

	if err := d.Set("name", obj.FullPath); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", obj.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("active_bonus", obj.ActiveBonus); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("enabled", obj.Enabled); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	pools := make([]map[string]interface{}, len(obj.Pools))
	for i, pool := range obj.Pools {
		poolMap := make(map[string]interface{})
		poolMap["name"] = pool.Name
		poolMap["weight"] = pool.Weight
		poolMap["attribute"] = pool.Attribute
		poolMap["minimum_threshold"] = pool.MinimumThreshold
		if pool.PercentUp > 0 {
			poolMap["percent_up"] = pool.PercentUp
		}
		if pool.SufficientThreshold != "" {
			poolMap["sufficient_threshold"] = pool.SufficientThreshold
		}
		pools[i] = poolMap
	}
	if err := d.Set("pools", pools); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	clusters := make([]map[string]interface{}, len(obj.Clusters))
	for i, cluster := range obj.Clusters {
		clusterMap := make(map[string]interface{})
		clusterMap["name"] = cluster.Name
		clusterMap["weight"] = cluster.Weight
		clusterMap["attribute"] = cluster.Attribute
		clusterMap["minimum_threshold"] = cluster.MinimumThreshold
		if cluster.PercentUp > 0 {
			clusterMap["percent_up"] = cluster.PercentUp
		}
		if cluster.SufficientThreshold != "" {
			clusterMap["sufficient_threshold"] = cluster.SufficientThreshold
		}
		clusters[i] = clusterMap
	}
	if err := d.Set("clusters", clusters); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	trunks := make([]map[string]interface{}, len(obj.Trunks))
	for i, trunk := range obj.Trunks {
		trunkMap := make(map[string]interface{})
		trunkMap["name"] = trunk.Name
		trunkMap["weight"] = trunk.Weight
		trunkMap["attribute"] = trunk.Attribute
		trunkMap["minimum_threshold"] = trunk.MinimumThreshold
		if trunk.PercentUp > 0 {
			trunkMap["percent_up"] = trunk.PercentUp
		}
		if trunk.SufficientThreshold != "" {
			trunkMap["sufficient_threshold"] = trunk.SufficientThreshold
		}
		trunks[i] = trunkMap
	}
	if err := d.Set("trunks", trunks); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceBigipSysHaGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Printf("[INFO] Deleting HA Group: %s", name)

	err := client.DeleteHaGroup(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete HA Group (%s) (%v)", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
