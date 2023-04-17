package network

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
	"github.com/tombuildsstuff/kermit/sdk/network/2022-07-01/network"
)

var SubnetResourceName = "azurerm_subnet"

var subnetDelegationServiceNames = []string{
	"Microsoft.ApiManagement/service",
	"Microsoft.AzureCosmosDB/clusters",
	"Microsoft.BareMetal/AzureVMware",
	"Microsoft.BareMetal/CrayServers",
	"Microsoft.Batch/batchAccounts",
	"Microsoft.ContainerInstance/containerGroups",
	"Microsoft.ContainerService/managedClusters",
	"Microsoft.Databricks/workspaces",
	"Microsoft.DBforMySQL/flexibleServers",
	"Microsoft.DBforMySQL/serversv2",
	"Microsoft.DBforPostgreSQL/flexibleServers",
	"Microsoft.DBforPostgreSQL/serversv2",
	"Microsoft.DBforPostgreSQL/singleServers",
	"Microsoft.HardwareSecurityModules/dedicatedHSMs",
	"Microsoft.Kusto/clusters",
	"Microsoft.Logic/integrationServiceEnvironments",
	"Microsoft.LabServices/labplans",
	"Microsoft.MachineLearningServices/workspaces",
	"Microsoft.Netapp/volumes",
	"Microsoft.Network/dnsResolvers",
	"Microsoft.Network/managedResolvers",
	"Microsoft.PowerPlatform/vnetaccesslinks",
	"Microsoft.ServiceFabricMesh/networks",
	"Microsoft.Sql/managedInstances",
	"Microsoft.Sql/servers",
	"Microsoft.StoragePool/diskPools",
	"Microsoft.StreamAnalytics/streamingJobs",
	"Microsoft.Synapse/workspaces",
	"Microsoft.Web/hostingEnvironments",
	"Microsoft.Web/serverFarms",
	"Microsoft.Orbital/orbitalGateways",
	"NGINX.NGINXPLUS/nginxDeployments",
	"PaloAltoNetworks.Cloudngfw/firewalls",
	"Qumulo.Storage/fileSystems",
}

func resourceSubnet() *pluginsdk.Resource {
	resource := &pluginsdk.Resource{
		Read: resourceSubnetRead,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SubnetID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"virtual_network_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"address_prefixes": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},

			"service_endpoints": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
			},

			"service_endpoint_policy_ids": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validate.SubnetServiceEndpointStoragePolicyID,
				},
			},

			"delegation": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"service_delegation": {
							Type:     pluginsdk.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(subnetDelegationServiceNames, false),
									},

									"actions": {
										Type:       pluginsdk.TypeList,
										Optional:   true,
										ConfigMode: pluginsdk.SchemaConfigModeAttr,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												"Microsoft.Network/networkinterfaces/*",
												"Microsoft.Network/publicIPAddresses/join/action",
												"Microsoft.Network/publicIPAddresses/read",
												"Microsoft.Network/virtualNetworks/read",
												"Microsoft.Network/virtualNetworks/subnets/action",
												"Microsoft.Network/virtualNetworks/subnets/join/action",
												"Microsoft.Network/virtualNetworks/subnets/prepareNetworkPolicies/action",
												"Microsoft.Network/virtualNetworks/subnets/unprepareNetworkPolicies/action",
											}, false),
										},
									},
								},
							},
						},
					},
				},
			},

			"private_endpoint_network_policies_enabled": {
				Type: pluginsdk.TypeBool,
				Computed: func() bool {
					return !features.FourPointOh()
				}(),
				Optional: true,
				Default: func() interface{} {
					if !features.FourPointOh() {
						return nil
					}
					return !features.FourPointOh()
				}(),
				ConflictsWith: func() []string {
					if !features.FourPointOh() {
						return []string{"enforce_private_link_endpoint_network_policies"}
					}
					return []string{}
				}(),
			},

			"private_link_service_network_policies_enabled": {
				Type: pluginsdk.TypeBool,
				Computed: func() bool {
					return !features.FourPointOh()
				}(),
				Optional: true,
				Default: func() interface{} {
					if !features.FourPointOh() {
						return nil
					}
					return features.FourPointOh()
				}(),
				ConflictsWith: func() []string {
					if !features.FourPointOh() {
						return []string{"enforce_private_link_service_network_policies"}
					}
					return []string{}
				}(),
			},
		},
	}

	if !features.FourPointOhBeta() {
		resource.Schema["enforce_private_link_endpoint_network_policies"] = &pluginsdk.Schema{
			Type:          pluginsdk.TypeBool,
			Computed:      true,
			Optional:      true,
			Deprecated:    "`enforce_private_link_endpoint_network_policies` will be removed in favour of the property `private_endpoint_network_policies_enabled` in version 4.0 of the AzureRM Provider",
			ConflictsWith: []string{"private_endpoint_network_policies_enabled"},
		}

		resource.Schema["enforce_private_link_service_network_policies"] = &pluginsdk.Schema{
			Type:          pluginsdk.TypeBool,
			Computed:      true,
			Optional:      true,
			Deprecated:    "`enforce_private_link_service_network_policies` will be removed in favour of the property `private_link_service_network_policies_enabled` in version 4.0 of the AzureRM Provider",
			ConflictsWith: []string{"private_link_service_network_policies_enabled"},
		}
	}

	return resource
}

func resourceSubnetRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SubnetsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SubnetID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("virtual_network_name", id.VirtualNetworkName)
	d.Set("resource_group_name", id.ResourceGroup)

	if props := resp.SubnetPropertiesFormat; props != nil {
		if props.AddressPrefixes == nil {
			if props.AddressPrefix != nil && len(*props.AddressPrefix) > 0 {
				d.Set("address_prefixes", []string{*props.AddressPrefix})
			} else {
				d.Set("address_prefixes", []string{})
			}
		} else {
			d.Set("address_prefixes", props.AddressPrefixes)
		}

		delegation := flattenSubnetDelegation(props.Delegations)
		if err := d.Set("delegation", delegation); err != nil {
			return fmt.Errorf("flattening `delegation`: %+v", err)
		}

		if !features.FourPointOhBeta() {
			d.Set("enforce_private_link_endpoint_network_policies", flattenEnforceSubnetNetworkPolicy(string(props.PrivateEndpointNetworkPolicies)))
			d.Set("enforce_private_link_service_network_policies", flattenEnforceSubnetNetworkPolicy(string(props.PrivateLinkServiceNetworkPolicies)))
		}

		d.Set("private_endpoint_network_policies_enabled", flattenSubnetNetworkPolicy(string(props.PrivateEndpointNetworkPolicies)))
		d.Set("private_link_service_network_policies_enabled", flattenSubnetNetworkPolicy(string(props.PrivateLinkServiceNetworkPolicies)))

		serviceEndpoints := flattenSubnetServiceEndpoints(props.ServiceEndpoints)
		if err := d.Set("service_endpoints", serviceEndpoints); err != nil {
			return fmt.Errorf("setting `service_endpoints`: %+v", err)
		}

		serviceEndpointPolicies := flattenSubnetServiceEndpointPolicies(props.ServiceEndpointPolicies)
		if err := d.Set("service_endpoint_policy_ids", serviceEndpointPolicies); err != nil {
			return fmt.Errorf("setting `service_endpoint_policy_ids`: %+v", err)
		}
	}

	return nil
}

func expandSubnetServiceEndpoints(input []interface{}) *[]network.ServiceEndpointPropertiesFormat {
	endpoints := make([]network.ServiceEndpointPropertiesFormat, 0)

	for _, svcEndpointRaw := range input {
		if svc, ok := svcEndpointRaw.(string); ok {
			endpoint := network.ServiceEndpointPropertiesFormat{
				Service: &svc,
			}
			endpoints = append(endpoints, endpoint)
		}
	}

	return &endpoints
}

func flattenSubnetServiceEndpoints(serviceEndpoints *[]network.ServiceEndpointPropertiesFormat) []interface{} {
	endpoints := make([]interface{}, 0)

	if serviceEndpoints == nil {
		return endpoints
	}

	for _, endpoint := range *serviceEndpoints {
		if endpoint.Service != nil {
			endpoints = append(endpoints, *endpoint.Service)
		}
	}

	return endpoints
}

func expandSubnetDelegation(input []interface{}) *[]network.Delegation {
	retDelegations := make([]network.Delegation, 0)

	for _, deleValue := range input {
		deleData := deleValue.(map[string]interface{})
		deleName := deleData["name"].(string)
		srvDelegations := deleData["service_delegation"].([]interface{})
		srvDelegation := srvDelegations[0].(map[string]interface{})
		srvName := srvDelegation["name"].(string)
		srvActions := srvDelegation["actions"].([]interface{})

		retSrvActions := make([]string, 0)
		for _, srvAction := range srvActions {
			srvActionData := srvAction.(string)
			retSrvActions = append(retSrvActions, srvActionData)
		}

		retDelegation := network.Delegation{
			Name: &deleName,
			ServiceDelegationPropertiesFormat: &network.ServiceDelegationPropertiesFormat{
				ServiceName: &srvName,
				Actions:     &retSrvActions,
			},
		}

		retDelegations = append(retDelegations, retDelegation)
	}

	return &retDelegations
}

func flattenSubnetDelegation(delegations *[]network.Delegation) []interface{} {
	if delegations == nil {
		return []interface{}{}
	}

	retDeles := make([]interface{}, 0)

	normalizeServiceName := map[string]string{}
	for _, normName := range subnetDelegationServiceNames {
		normalizeServiceName[strings.ToLower(normName)] = normName
	}

	for _, dele := range *delegations {
		retDele := make(map[string]interface{})
		if v := dele.Name; v != nil {
			retDele["name"] = *v
		}

		svcDeles := make([]interface{}, 0)
		svcDele := make(map[string]interface{})
		if props := dele.ServiceDelegationPropertiesFormat; props != nil {
			if v := props.ServiceName; v != nil {
				name := *v
				if nv, ok := normalizeServiceName[strings.ToLower(name)]; ok {
					name = nv
				}
				svcDele["name"] = name
			}

			if v := props.Actions; v != nil {
				svcDele["actions"] = *v
			}
		}

		svcDeles = append(svcDeles, svcDele)

		retDele["service_delegation"] = svcDeles

		retDeles = append(retDeles, retDele)
	}

	return retDeles
}

// TODO 4.0: Remove expandEnforceSubnetPrivateLinkNetworkPolicy function
func expandEnforceSubnetNetworkPolicy(enabled bool) string {
	// This is strange logic, but to get the schema to make sense for the end user
	// I exposed it with the same name that the Azure CLI does to be consistent
	// between the tool sets, which means true == Disabled.
	if enabled {
		return string(network.VirtualNetworkPrivateEndpointNetworkPoliciesDisabled)
	}

	return string(network.VirtualNetworkPrivateEndpointNetworkPoliciesEnabled)
}

func expandSubnetNetworkPolicy(enabled bool) string {
	if enabled {
		return string(network.VirtualNetworkPrivateEndpointNetworkPoliciesEnabled)
	}

	return string(network.VirtualNetworkPrivateEndpointNetworkPoliciesDisabled)
}

// TODO 4.0: Remove flattenEnforceSubnetPrivateLinkNetworkPolicy function
func flattenEnforceSubnetNetworkPolicy(input string) bool {
	// This is strange logic, but to get the schema to make sense for the end user
	// I exposed it with the same name that the Azure CLI does to be consistent
	// between the tool sets, which means true == Disabled.
	return strings.EqualFold(input, string(network.VirtualNetworkPrivateEndpointNetworkPoliciesDisabled))
}

func flattenSubnetNetworkPolicy(input string) bool {
	return strings.EqualFold(input, string(network.VirtualNetworkPrivateEndpointNetworkPoliciesEnabled))
}

func expandSubnetServiceEndpointPolicies(input []interface{}) *[]network.ServiceEndpointPolicy {
	output := make([]network.ServiceEndpointPolicy, 0)
	for _, policy := range input {
		policy := policy.(string)
		output = append(output, network.ServiceEndpointPolicy{ID: &policy})
	}
	return &output
}

func flattenSubnetServiceEndpointPolicies(input *[]network.ServiceEndpointPolicy) []interface{} {
	if input == nil {
		return nil
	}

	var output []interface{}
	for _, policy := range *input {
		id := ""
		if policy.ID != nil {
			id = *policy.ID
		}
		output = append(output, id)
	}
	return output
}

func SubnetProvisioningStateRefreshFunc(ctx context.Context, client *network.SubnetsClient, id parse.SubnetId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
		if err != nil {
			return nil, "", fmt.Errorf("polling for %s: %+v", id.String(), err)
		}

		return res, string(res.ProvisioningState), nil
	}
}
