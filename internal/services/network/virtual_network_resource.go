package network

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tags"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
	"github.com/tombuildsstuff/kermit/sdk/network/2022-07-01/network"
)

var VirtualNetworkResourceName = "azurerm_virtual_network"

func resourceVirtualNetwork() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: resourceVirtualNetworkRead,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VirtualNetworkID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: resourceVirtualNetworkSchema(),
	}
}

func resourceVirtualNetworkSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"resource_group_name": commonschema.ResourceGroupName(),

		"location": commonschema.Location(),

		"address_space": {
			Type:     pluginsdk.TypeList,
			Required: true,
			MinItems: 1,
			Elem: &pluginsdk.Schema{
				Type:         pluginsdk.TypeString,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},

		// Optional

		"dns_servers": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type:         pluginsdk.TypeString,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},

		"edge_zone": commonschema.EdgeZoneOptionalForceNew(),

		"flow_timeout_in_minutes": {
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(4, 30),
		},

		"guid": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"subnet": {
			Type:       pluginsdk.TypeSet,
			Optional:   true,
			Computed:   true,
			ConfigMode: pluginsdk.SchemaConfigModeAttr,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"address_prefix": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"security_group": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},

					"id": {
						Type:     pluginsdk.TypeString,
						Computed: true,
					},
				},
			},
			Set: resourceAzureSubnetHash,
		},

		"tags": tags.Schema(),
	}
}

func resourceVirtualNetworkRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualNetworkID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	d.Set("location", location.NormalizeNilable(resp.Location))
	d.Set("edge_zone", flattenEdgeZone(resp.ExtendedLocation))

	if props := resp.VirtualNetworkPropertiesFormat; props != nil {
		d.Set("guid", props.ResourceGUID)
		d.Set("flow_timeout_in_minutes", props.FlowTimeoutInMinutes)

		if space := props.AddressSpace; space != nil {
			d.Set("address_space", utils.FlattenStringSlice(space.AddressPrefixes))
		}

		if err := d.Set("ddos_protection_plan", flattenVirtualNetworkDDoSProtectionPlan(props)); err != nil {
			return fmt.Errorf("setting `ddos_protection_plan`: %+v", err)
		}

		if err := d.Set("subnet", flattenVirtualNetworkSubnets(props.Subnets)); err != nil {
			return fmt.Errorf("setting `subnets`: %+v", err)
		}

		if err := d.Set("dns_servers", flattenVirtualNetworkDNSServers(props.DhcpOptions)); err != nil {
			return fmt.Errorf("setting `dns_servers`: %+v", err)
		}

		bgpCommunity := ""
		if p := props.BgpCommunities; p != nil {
			if v := p.VirtualNetworkCommunity; v != nil {
				bgpCommunity = *v
			}
		}
		if err := d.Set("bgp_community", bgpCommunity); err != nil {
			return fmt.Errorf("setting `bgp_community`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func expandVirtualNetworkProperties(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) (*network.VirtualNetworkPropertiesFormat, error) {
	subnets := make([]network.Subnet, 0)
	if subs := d.Get("subnet").(*pluginsdk.Set); subs.Len() > 0 {
		for _, subnet := range subs.List() {
			subnet := subnet.(map[string]interface{})

			name := subnet["name"].(string)
			log.Printf("[INFO] setting subnets inside vNet, processing %q", name)
			// since subnets can also be created outside of vNet definition (as root objects)
			// do a GET on subnet properties from the server before setting them
			resGroup := d.Get("resource_group_name").(string)
			vnetName := d.Get("name").(string)
			subnetObj, err := getExistingSubnet(ctx, resGroup, vnetName, name, meta)
			if err != nil {
				return nil, err
			}
			log.Printf("[INFO] Completed GET of Subnet props ")

			prefix := subnet["address_prefix"].(string)
			secGroup := subnet["security_group"].(string)

			// set the props from config and leave the rest intact
			subnetObj.Name = &name
			if subnetObj.SubnetPropertiesFormat == nil {
				subnetObj.SubnetPropertiesFormat = &network.SubnetPropertiesFormat{}
			}

			subnetObj.SubnetPropertiesFormat.AddressPrefix = &prefix

			if secGroup != "" {
				subnetObj.SubnetPropertiesFormat.NetworkSecurityGroup = &network.SecurityGroup{
					ID: &secGroup,
				}
			} else {
				subnetObj.SubnetPropertiesFormat.NetworkSecurityGroup = nil
			}

			subnets = append(subnets, *subnetObj)
		}
	}

	properties := &network.VirtualNetworkPropertiesFormat{
		AddressSpace: &network.AddressSpace{
			AddressPrefixes: utils.ExpandStringSlice(d.Get("address_space").([]interface{})),
		},
		DhcpOptions: &network.DhcpOptions{
			DNSServers: utils.ExpandStringSlice(d.Get("dns_servers").([]interface{})),
		},
		Subnets: &subnets,
	}

	if v, ok := d.GetOk("ddos_protection_plan"); ok {
		rawList := v.([]interface{})

		var ddosPPlan map[string]interface{}
		if len(rawList) > 0 {
			ddosPPlan = rawList[0].(map[string]interface{})
		}

		if v, ok := ddosPPlan["id"]; ok {
			id := v.(string)
			properties.DdosProtectionPlan = &network.SubResource{
				ID: &id,
			}
		}

		if v, ok := ddosPPlan["enable"]; ok {
			enable := v.(bool)
			properties.EnableDdosProtection = &enable
		}
	}

	if v, ok := d.GetOk("bgp_community"); ok {
		properties.BgpCommunities = &network.VirtualNetworkBgpCommunities{VirtualNetworkCommunity: utils.String(v.(string))}
	}

	return properties, nil
}

func flattenVirtualNetworkDDoSProtectionPlan(input *network.VirtualNetworkPropertiesFormat) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	if input.DdosProtectionPlan == nil || input.DdosProtectionPlan.ID == nil || input.EnableDdosProtection == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"id":     *input.DdosProtectionPlan.ID,
			"enable": *input.EnableDdosProtection,
		},
	}
}

func flattenVirtualNetworkSubnets(input *[]network.Subnet) *pluginsdk.Set {
	results := &pluginsdk.Set{
		F: resourceAzureSubnetHash,
	}

	if subnets := input; subnets != nil {
		for _, subnet := range *input {
			output := map[string]interface{}{}

			if id := subnet.ID; id != nil {
				output["id"] = *id
			}

			if name := subnet.Name; name != nil {
				output["name"] = *name
			}

			if props := subnet.SubnetPropertiesFormat; props != nil {
				if prefix := props.AddressPrefix; prefix != nil {
					output["address_prefix"] = *prefix
				}

				if nsg := props.NetworkSecurityGroup; nsg != nil {
					if nsg.ID != nil {
						output["security_group"] = *nsg.ID
					}
				}
			}

			results.Add(output)
		}
	}

	return results
}

func flattenVirtualNetworkDNSServers(input *network.DhcpOptions) []string {
	results := make([]string, 0)

	if input != nil {
		if servers := input.DNSServers; servers != nil {
			results = *servers
		}
	}

	return results
}

func resourceAzureSubnetHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(m["name"].(string))
		if v, ok := m["address_prefix"]; ok {
			buf.WriteString(v.(string))
		}
		if v, ok := m["security_group"]; ok {
			buf.WriteString(v.(string))
		}
	}

	return pluginsdk.HashString(buf.String())
}

func getExistingSubnet(ctx context.Context, resGroup string, vnetName string, subnetName string, meta interface{}) (*network.Subnet, error) {
	subnetClient := meta.(*clients.Client).Network.SubnetsClient
	resp, err := subnetClient.Get(ctx, resGroup, vnetName, subnetName, "")
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return &network.Subnet{}, nil
		}
		// raise an error if there was an issue other than 404 in getting subnet properties
		return nil, err
	}

	// Return it directly rather than copy the fields to prevent potential uncovered properties (for example, `ServiceEndpoints` mentioned in #1619)
	return &resp, nil
}

func VirtualNetworkProvisioningStateRefreshFunc(ctx context.Context, client *network.VirtualNetworksClient, id parse.VirtualNetworkId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			return nil, "", fmt.Errorf("polling for %s: %+v", id.String(), err)
		}

		return res, string(res.ProvisioningState), nil
	}
}
