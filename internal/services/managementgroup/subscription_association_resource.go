package managementgroup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-05-01/managementgroups" // nolint: staticcheck
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/managementgroup/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/managementgroup/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceManagementGroupSubscriptionAssociation() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: resourceManagementGroupSubscriptionAssociationRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(5 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ManagementGroupSubscriptionAssociationID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"management_group_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ManagementGroupID,
			},

			"subscription_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: commonids.ValidateSubscriptionID,
			},
		},
	}
}

func resourceManagementGroupSubscriptionAssociationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	// There is no "read" function on the appropriate client so we need to check if the Subscription is in the Management Group subscription list
	client := meta.(*clients.Client).ManagementGroups.GroupsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ManagementGroupSubscriptionAssociationID(d.Id())
	if err != nil {
		return err
	}

	managementGroup, err := client.Get(ctx, id.ManagementGroup, "children", utils.Bool(false), "", "")
	if err != nil {
		return fmt.Errorf("reading Management Group %q for Subscription Associations: %+v", id.ManagementGroup, err)
	}
	found := false
	if props := managementGroup.Properties; props != nil {
		if props.Children == nil {
			return fmt.Errorf("could not read properties for Management Group %q", id.ManagementGroup)
		}

		for _, v := range *props.Children {
			if v.Type == managementgroups.Type1Subscriptions {
				if v.Name != nil && *v.Name == id.SubscriptionId {
					found = true
				}
			}
		}

		if !found {
			log.Printf("[INFO] Subscription %q not found in Management group %q, removing from state", id.SubscriptionId, id.ManagementGroup)
			d.SetId("")
			return nil
		}

		managementGroupId := parse.NewManagementGroupId(id.ManagementGroup)
		d.Set("management_group_id", managementGroupId.ID())
		subscriptionId := commonids.NewSubscriptionID(id.SubscriptionId)
		d.Set("subscription_id", subscriptionId.ID())
	}

	return nil
}

func subscriptionAssociationRefreshFunc(ctx context.Context, client *managementgroups.Client, id parse.ManagementGroupSubscriptionAssociationId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		managementGroup, err := client.Get(ctx, id.ManagementGroup, "children", utils.Bool(false), "", "")
		if err != nil {
			return nil, "", fmt.Errorf("reading Management Group %q for Subscription Associations: %+v", id.ManagementGroup, err)
		}

		if props := managementGroup.Properties; props != nil && props.Children != nil {
			for _, v := range *props.Children {
				if v.Type == managementgroups.Type1Subscriptions {
					if v.Name != nil && *v.Name == id.SubscriptionId {
						return managementGroup, "Exists", nil
					}
				}
			}
		}

		return "NotFound", "NotFound", nil
	}
}
