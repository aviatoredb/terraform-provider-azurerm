package provider

import (
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

func schemaFeatures(supportLegacyTestSuite bool) *pluginsdk.Schema {
	// NOTE: if there's only one nested field these want to be Required (since there's no point
	//       specifying the block otherwise) - however for 2+ they should be optional
	featuresMap := map[string]*pluginsdk.Schema{
		//lintignore:XS003

		"template_deployment": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"delete_nested_items_during_deletion": {
						Type:     pluginsdk.TypeBool,
						Required: true,
					},
				},
			},
		},

		"key_vault": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"purge_soft_delete_on_destroy": {
						Description: "When enabled soft-deleted `azurerm_key_vault` resources will be permanently deleted (e.g purged), when destroyed",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
					},

					"purge_soft_deleted_certificates_on_destroy": {
						Description: "When enabled soft-deleted `azurerm_key_vault_certificate` resources will be permanently deleted (e.g purged), when destroyed",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
					},

					"purge_soft_deleted_keys_on_destroy": {
						Description: "When enabled soft-deleted `azurerm_key_vault_key` resources will be permanently deleted (e.g purged), when destroyed",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
					},

					"purge_soft_deleted_secrets_on_destroy": {
						Description: "When enabled soft-deleted `azurerm_key_vault_secret` resources will be permanently deleted (e.g purged), when destroyed",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
					},

					"purge_soft_deleted_hardware_security_modules_on_destroy": {
						Description: "When enabled soft-deleted `azurerm_key_vault_managed_hardware_security_module` resources will be permanently deleted (e.g purged), when destroyed",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
					},

					"recover_soft_deleted_certificates": {
						Description: "When enabled soft-deleted `azurerm_key_vault_certificate` resources will be restored, instead of creating new ones",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
					},

					"recover_soft_deleted_key_vaults": {
						Description: "When enabled soft-deleted `azurerm_key_vault` resources will be restored, instead of creating new ones",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
					},

					"recover_soft_deleted_keys": {
						Description: "When enabled soft-deleted `azurerm_key_vault_key` resources will be restored, instead of creating new ones",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
					},

					"recover_soft_deleted_secrets": {
						Description: "When enabled soft-deleted `azurerm_key_vault_secret` resources will be restored, instead of creating new ones",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
					},
				},
			},
		},

		"resource_group": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*schema.Schema{
					"prevent_deletion_if_contains_resources": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  os.Getenv("TF_ACC") == "",
					},
				},
			},
		},
	}

	// this is a temporary hack to enable us to gradually add provider blocks to test configurations
	// rather than doing it as a big-bang and breaking all open PR's
	if supportLegacyTestSuite {
		return &pluginsdk.Schema{
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: featuresMap,
			},
		}
	}

	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Required: true,
		MaxItems: 1,
		MinItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: featuresMap,
		},
	}
}

func expandFeatures(input []interface{}) features.UserFeatures {
	// these are the defaults if omitted from the config
	featuresMap := features.Default()

	if len(input) == 0 || input[0] == nil {
		return featuresMap
	}

	val := input[0].(map[string]interface{})

	if raw, ok := val["key_vault"]; ok {
		items := raw.([]interface{})
		if len(items) > 0 && items[0] != nil {
			keyVaultRaw := items[0].(map[string]interface{})
			if v, ok := keyVaultRaw["purge_soft_delete_on_destroy"]; ok {
				featuresMap.KeyVault.PurgeSoftDeleteOnDestroy = v.(bool)
			}
			if v, ok := keyVaultRaw["purge_soft_deleted_certificates_on_destroy"]; ok {
				featuresMap.KeyVault.PurgeSoftDeletedCertsOnDestroy = v.(bool)
			}
			if v, ok := keyVaultRaw["purge_soft_deleted_keys_on_destroy"]; ok {
				featuresMap.KeyVault.PurgeSoftDeletedKeysOnDestroy = v.(bool)
			}
			if v, ok := keyVaultRaw["purge_soft_deleted_secrets_on_destroy"]; ok {
				featuresMap.KeyVault.PurgeSoftDeletedSecretsOnDestroy = v.(bool)
			}
			if v, ok := keyVaultRaw["purge_soft_deleted_hardware_security_modules_on_destroy"]; ok {
				featuresMap.KeyVault.PurgeSoftDeletedHSMsOnDestroy = v.(bool)
			}
			if v, ok := keyVaultRaw["recover_soft_deleted_certificates"]; ok {
				featuresMap.KeyVault.RecoverSoftDeletedCerts = v.(bool)
			}
			if v, ok := keyVaultRaw["recover_soft_deleted_key_vaults"]; ok {
				featuresMap.KeyVault.RecoverSoftDeletedKeyVaults = v.(bool)
			}
			if v, ok := keyVaultRaw["recover_soft_deleted_keys"]; ok {
				featuresMap.KeyVault.RecoverSoftDeletedKeys = v.(bool)
			}
			if v, ok := keyVaultRaw["recover_soft_deleted_secrets"]; ok {
				featuresMap.KeyVault.RecoverSoftDeletedSecrets = v.(bool)
			}
		}
	}

	if raw, ok := val["template_deployment"]; ok {
		items := raw.([]interface{})
		if len(items) > 0 {
			templateRaw := items[0].(map[string]interface{})
			if v, ok := templateRaw["delete_nested_items_during_deletion"]; ok {
				featuresMap.TemplateDeployment.DeleteNestedItemsDuringDeletion = v.(bool)
			}
		}
	}

	if raw, ok := val["resource_group"]; ok {
		items := raw.([]interface{})
		if len(items) > 0 {
			resourceGroupRaw := items[0].(map[string]interface{})
			if v, ok := resourceGroupRaw["prevent_deletion_if_contains_resources"]; ok {
				featuresMap.ResourceGroup.PreventDeletionIfContainsResources = v.(bool)
			}
		}
	}

	return featuresMap
}
