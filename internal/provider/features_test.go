package provider

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
)

func TestExpandFeatures(t *testing.T) {
	testData := []struct {
		Name     string
		Input    []interface{}
		EnvVars  map[string]interface{}
		Expected features.UserFeatures
	}{
		{
			Name:  "Empty Block",
			Input: []interface{}{},
			Expected: features.UserFeatures{
				KeyVault: features.KeyVaultFeatures{
					PurgeSoftDeletedCertsOnDestroy:   true,
					PurgeSoftDeletedKeysOnDestroy:    true,
					PurgeSoftDeletedSecretsOnDestroy: true,
					PurgeSoftDeleteOnDestroy:         true,
					PurgeSoftDeletedHSMsOnDestroy:    true,
					RecoverSoftDeletedCerts:          true,
					RecoverSoftDeletedKeys:           true,
					RecoverSoftDeletedKeyVaults:      true,
					RecoverSoftDeletedSecrets:        true,
				},
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
				TemplateDeployment: features.TemplateDeploymentFeatures{
					DeleteNestedItemsDuringDeletion: false,
				},
			},
		},
		{
			Name: "Complete Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"key_vault": []interface{}{
						map[string]interface{}{
							"purge_soft_deleted_certificates_on_destroy":              true,
							"purge_soft_deleted_keys_on_destroy":                      true,
							"purge_soft_deleted_secrets_on_destroy":                   true,
							"purge_soft_deleted_hardware_security_modules_on_destroy": true,
							"purge_soft_delete_on_destroy":                            true,
							"recover_soft_deleted_certificates":                       true,
							"recover_soft_deleted_keys":                               true,
							"recover_soft_deleted_key_vaults":                         true,
							"recover_soft_deleted_secrets":                            true,
						},
					},
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": true,
						},
					},
					"template_deployment": []interface{}{
						map[string]interface{}{
							"delete_nested_items_during_deletion": true,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				KeyVault: features.KeyVaultFeatures{
					PurgeSoftDeletedCertsOnDestroy:   true,
					PurgeSoftDeletedKeysOnDestroy:    true,
					PurgeSoftDeletedSecretsOnDestroy: true,
					PurgeSoftDeleteOnDestroy:         true,
					PurgeSoftDeletedHSMsOnDestroy:    true,
					RecoverSoftDeletedCerts:          true,
					RecoverSoftDeletedKeys:           true,
					RecoverSoftDeletedKeyVaults:      true,
					RecoverSoftDeletedSecrets:        true,
				},
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: true,
				},
				TemplateDeployment: features.TemplateDeploymentFeatures{
					DeleteNestedItemsDuringDeletion: true,
				},
			},
		},
		{
			Name: "Complete Disabled",
			Input: []interface{}{
				map[string]interface{}{
					"key_vault": []interface{}{
						map[string]interface{}{
							"purge_soft_deleted_certificates_on_destroy":              false,
							"purge_soft_deleted_keys_on_destroy":                      false,
							"purge_soft_deleted_secrets_on_destroy":                   false,
							"purge_soft_deleted_hardware_security_modules_on_destroy": false,
							"purge_soft_delete_on_destroy":                            false,
							"recover_soft_deleted_certificates":                       false,
							"recover_soft_deleted_keys":                               false,
							"recover_soft_deleted_key_vaults":                         false,
							"recover_soft_deleted_secrets":                            false,
						},
					},
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": false,
						},
					},
					"template_deployment": []interface{}{
						map[string]interface{}{
							"delete_nested_items_during_deletion": false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				KeyVault: features.KeyVaultFeatures{
					PurgeSoftDeletedCertsOnDestroy:   false,
					PurgeSoftDeletedKeysOnDestroy:    false,
					PurgeSoftDeletedSecretsOnDestroy: false,
					PurgeSoftDeletedHSMsOnDestroy:    false,
					PurgeSoftDeleteOnDestroy:         false,
					RecoverSoftDeletedCerts:          false,
					RecoverSoftDeletedKeys:           false,
					RecoverSoftDeletedKeyVaults:      false,
					RecoverSoftDeletedSecrets:        false,
				},
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
				TemplateDeployment: features.TemplateDeploymentFeatures{
					DeleteNestedItemsDuringDeletion: false,
				},
			},
		},
	}

	for _, testCase := range testData {
		t.Logf("[DEBUG] Test Case: %q", testCase.Name)
		result := expandFeatures(testCase.Input)
		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Fatalf("Expected %+v but got %+v", result, testCase.Expected)
		}
	}
}

func TestExpandFeaturesKeyVault(t *testing.T) {
	testData := []struct {
		Name     string
		Input    []interface{}
		EnvVars  map[string]interface{}
		Expected features.UserFeatures
	}{
		{
			Name: "Empty Block",
			Input: []interface{}{
				map[string]interface{}{
					"key_vault": []interface{}{},
				},
			},
			Expected: features.UserFeatures{
				KeyVault: features.KeyVaultFeatures{
					PurgeSoftDeletedCertsOnDestroy:   true,
					PurgeSoftDeletedKeysOnDestroy:    true,
					PurgeSoftDeletedSecretsOnDestroy: true,
					PurgeSoftDeleteOnDestroy:         true,
					PurgeSoftDeletedHSMsOnDestroy:    true,
					RecoverSoftDeletedCerts:          true,
					RecoverSoftDeletedKeys:           true,
					RecoverSoftDeletedKeyVaults:      true,
					RecoverSoftDeletedSecrets:        true,
				},
			},
		},
		{
			Name: "Purge Soft Delete On Destroy and Recover Soft Deleted Key Vaults Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"key_vault": []interface{}{
						map[string]interface{}{
							"purge_soft_deleted_certificates_on_destroy":              true,
							"purge_soft_deleted_keys_on_destroy":                      true,
							"purge_soft_deleted_secrets_on_destroy":                   true,
							"purge_soft_deleted_hardware_security_modules_on_destroy": true,
							"purge_soft_delete_on_destroy":                            true,
							"recover_soft_deleted_certificates":                       true,
							"recover_soft_deleted_keys":                               true,
							"recover_soft_deleted_key_vaults":                         true,
							"recover_soft_deleted_secrets":                            true,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				KeyVault: features.KeyVaultFeatures{
					PurgeSoftDeletedCertsOnDestroy:   true,
					PurgeSoftDeletedKeysOnDestroy:    true,
					PurgeSoftDeletedSecretsOnDestroy: true,
					PurgeSoftDeletedHSMsOnDestroy:    true,
					PurgeSoftDeleteOnDestroy:         true,
					RecoverSoftDeletedCerts:          true,
					RecoverSoftDeletedKeys:           true,
					RecoverSoftDeletedKeyVaults:      true,
					RecoverSoftDeletedSecrets:        true,
				},
			},
		},
		{
			Name: "Purge Soft Delete On Destroy and Recover Soft Deleted Key Vaults Disabled",
			Input: []interface{}{
				map[string]interface{}{
					"key_vault": []interface{}{
						map[string]interface{}{
							"purge_soft_deleted_certificates_on_destroy":              false,
							"purge_soft_deleted_keys_on_destroy":                      false,
							"purge_soft_deleted_secrets_on_destroy":                   false,
							"purge_soft_deleted_hardware_security_modules_on_destroy": false,
							"purge_soft_delete_on_destroy":                            false,
							"recover_soft_deleted_certificates":                       false,
							"recover_soft_deleted_keys":                               false,
							"recover_soft_deleted_key_vaults":                         false,
							"recover_soft_deleted_secrets":                            false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				KeyVault: features.KeyVaultFeatures{
					PurgeSoftDeletedCertsOnDestroy:   false,
					PurgeSoftDeletedKeysOnDestroy:    false,
					PurgeSoftDeletedSecretsOnDestroy: false,
					PurgeSoftDeleteOnDestroy:         false,
					PurgeSoftDeletedHSMsOnDestroy:    false,
					RecoverSoftDeletedCerts:          false,
					RecoverSoftDeletedKeyVaults:      false,
					RecoverSoftDeletedKeys:           false,
					RecoverSoftDeletedSecrets:        false,
				},
			},
		},
	}

	for _, testCase := range testData {
		t.Logf("[DEBUG] Test Case: %q", testCase.Name)
		result := expandFeatures(testCase.Input)
		if !reflect.DeepEqual(result.KeyVault, testCase.Expected.KeyVault) {
			t.Fatalf("Expected %+v but got %+v", result.KeyVault, testCase.Expected.KeyVault)
		}
	}
}

func TestExpandFeaturesResourceGroup(t *testing.T) {
	testData := []struct {
		Name     string
		Input    []interface{}
		EnvVars  map[string]interface{}
		Expected features.UserFeatures
	}{
		{
			Name: "Empty Block",
			Input: []interface{}{
				map[string]interface{}{
					"resource_group": []interface{}{},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
			},
		},
		{
			Name: "Prevent Deletion If Contains Resources Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": true,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: true,
				},
			},
		},
		{
			Name: "Prevent Deletion If Contains Resources Disabled",
			Input: []interface{}{
				map[string]interface{}{
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
			},
		},
	}

	for _, testCase := range testData {
		t.Logf("[DEBUG] Test Case: %q", testCase.Name)
		result := expandFeatures(testCase.Input)
		if !reflect.DeepEqual(result.ResourceGroup, testCase.Expected.ResourceGroup) {
			t.Fatalf("Expected %+v but got %+v", result.ResourceGroup, testCase.Expected.ResourceGroup)
		}
	}
}
