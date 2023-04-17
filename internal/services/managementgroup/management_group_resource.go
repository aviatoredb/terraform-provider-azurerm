package managementgroup

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-05-01/managementgroups" // nolint: staticcheck
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/managementgroup/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/managementgroup/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

var managementGroupCacheControl = "no-cache"

func resourceManagementGroup() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: resourceManagementGroupRead,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ManagementGroupID(id)
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
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validate.ManagementGroupName,
			},

			"display_name": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
			},

			"parent_management_group_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.ManagementGroupID,
			},

			"subscription_ids": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.IsUUID,
				},
				Set: pluginsdk.HashString,
			},
		},
	}
}

func resourceManagementGroupRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ManagementGroups.GroupsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ManagementGroupID(d.Id())
	if err != nil {
		return err
	}

	recurse := utils.Bool(true)
	resp, err := client.Get(ctx, id.Name, "children", recurse, "", managementGroupCacheControl)
	if err != nil {
		if utils.ResponseWasForbidden(resp.Response) || utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Management Group %q doesn't exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("unable to read Management Group %q: %+v", d.Id(), err)
	}

	d.Set("name", id.Name)

	if props := resp.Properties; props != nil {
		d.Set("display_name", props.DisplayName)

		subscriptionIds, err := flattenManagementGroupSubscriptionIds(props.Children)
		if err != nil {
			return fmt.Errorf("unable to flatten `subscription_ids`: %+v", err)
		}
		d.Set("subscription_ids", subscriptionIds)

		parentId := ""
		if details := props.Details; details != nil {
			if parent := details.Parent; parent != nil {
				if pid := parent.ID; pid != nil {
					parentId = *pid
				}
			}
		}
		d.Set("parent_management_group_id", parentId)
	}

	return nil
}

func expandManagementGroupSubscriptionIds(input *pluginsdk.Set) []string {
	output := make([]string, 0)

	if input != nil {
		for _, v := range input.List() {
			output = append(output, v.(string))
		}
	}

	return output
}

func flattenManagementGroupSubscriptionIds(input *[]managementgroups.ChildInfo) (*pluginsdk.Set, error) {
	subscriptionIds := &pluginsdk.Set{F: pluginsdk.HashString}
	if input == nil {
		return subscriptionIds, nil
	}

	for _, child := range *input {
		if child.ID == nil {
			continue
		}

		id, err := parseManagementGroupSubscriptionID(*child.ID)
		if err != nil {
			return nil, fmt.Errorf("unable to parse child Subscription ID %+v", err)
		}

		if id != nil {
			subscriptionIds.Add(id.subscriptionId)
		}
	}

	return subscriptionIds, nil
}

type subscriptionId struct {
	subscriptionId string
}

func parseManagementGroupSubscriptionID(input string) (*subscriptionId, error) {
	// this is either:
	// /subscriptions/00000000-0000-0000-0000-000000000000

	// we skip out the child managementGroup ID's
	if strings.HasPrefix(input, "/providers/Microsoft.Management/managementGroups/") {
		return nil, nil
	}

	components := strings.Split(input, "/")

	if len(components) == 0 {
		return nil, fmt.Errorf("subscription Id is empty or not formatted correctly: %s", input)
	}

	if len(components) != 3 {
		return nil, fmt.Errorf("subscription Id should have 2 segments, got %d: %q", len(components)-1, input)
	}

	id := subscriptionId{
		subscriptionId: components[2],
	}
	return &id, nil
}

func determineManagementGroupSubscriptionsIdsToRemove(existing *[]managementgroups.ChildInfo, updated []string) (*[]string, error) {
	subscriptionIdsToRemove := make([]string, 0)
	if existing == nil {
		return &subscriptionIdsToRemove, nil
	}

	for _, v := range *existing {
		if v.ID == nil {
			continue
		}

		id, err := parseManagementGroupSubscriptionID(*v.ID)
		if err != nil {
			return nil, fmt.Errorf("unable to parse Subscription ID %q: %+v", *v.ID, err)
		}

		// not a Subscription - so let's skip it
		if id == nil {
			continue
		}

		found := false
		for _, subId := range updated {
			if id.subscriptionId == subId {
				found = true
				break
			}
		}

		if !found {
			subscriptionIdsToRemove = append(subscriptionIdsToRemove, id.subscriptionId)
		}
	}

	return &subscriptionIdsToRemove, nil
}

func managementgroupCreateStateRefreshFunc(ctx context.Context, client *managementgroups.Client, groupName string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := client.Get(ctx, groupName, "children", utils.Bool(true), "", managementGroupCacheControl)
		if err != nil {
			if utils.ResponseWasForbidden(resp.Response) {
				return resp, "pending", nil
			}
			return resp, "failed", err
		}

		return resp, "succeeded", nil
	}
}
