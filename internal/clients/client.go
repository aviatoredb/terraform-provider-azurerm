package clients

import (
	"context"
	"fmt"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
	keyvault "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault/client"
	managementgroup "github.com/hashicorp/terraform-provider-azurerm/internal/services/managementgroup/client"
	network "github.com/hashicorp/terraform-provider-azurerm/internal/services/network/client"
	resource "github.com/hashicorp/terraform-provider-azurerm/internal/services/resource/client"
	storage "github.com/hashicorp/terraform-provider-azurerm/internal/services/storage/client"
)

type Client struct {
	autoClient

	// StopContext is used for propagating control from Terraform Core (e.g. Ctrl/Cmd+C)
	StopContext context.Context

	Account  *ResourceManagerAccount
	Features features.UserFeatures

	KeyVault         *keyvault.Client
	ManagementGroups *managementgroup.Client
	Network          *network.Client
	Resource         *resource.Client
	Storage          *storage.Client
}

// NOTE: it should be possible for this method to become Private once the top level Client's removed

func (client *Client) Build(ctx context.Context, o *common.ClientOptions) error {
	autorest.Count429AsRetry = false
	// Disable the Azure SDK for Go's validation since it's unhelpful for our use-case
	validation.Disabled = true

	if err := buildAutoClients(&client.autoClient, o); err != nil {
		return fmt.Errorf("building auto-clients: %+v", err)
	}

	client.Features = o.Features
	client.StopContext = ctx

	client.KeyVault = keyvault.NewClient(o)
	client.ManagementGroups = managementgroup.NewClient(o)
	client.Network = network.NewClient(o)
	client.Storage = storage.NewClient(o)
	client.Resource = resource.NewClient(o)

	return nil
}
