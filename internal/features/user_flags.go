package features

type UserFeatures struct {
	KeyVault           KeyVaultFeatures
	ResourceGroup      ResourceGroupFeatures
	TemplateDeployment TemplateDeploymentFeatures
}

type KeyVaultFeatures struct {
	PurgeSoftDeleteOnDestroy         bool
	PurgeSoftDeletedKeysOnDestroy    bool
	PurgeSoftDeletedCertsOnDestroy   bool
	PurgeSoftDeletedSecretsOnDestroy bool
	PurgeSoftDeletedHSMsOnDestroy    bool
	RecoverSoftDeletedKeyVaults      bool
	RecoverSoftDeletedKeys           bool
	RecoverSoftDeletedCerts          bool
	RecoverSoftDeletedSecrets        bool
}

type ResourceGroupFeatures struct {
	PreventDeletionIfContainsResources bool
}

type TemplateDeploymentFeatures struct {
	DeleteNestedItemsDuringDeletion bool
}
