package acceptance

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/azure/terraform-provider-msgraph/internal/clients"
	"github.com/azure/terraform-provider-msgraph/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_client    *clients.Client
	clientLock = &sync.Mutex{}
)

func BuildTestClient() (*clients.Client, error) {
	clientLock.Lock()
	defer clientLock.Unlock()

	if _client == nil {
		var cloudConfig cloud.Configuration
		env := os.Getenv("ARM_ENVIRONMENT")
		switch strings.ToLower(env) {
		case "public":
			cloudConfig = cloud.AzurePublic
		case "usgovernment":
			cloudConfig = cloud.AzureGovernment
		case "china":
			cloudConfig = cloud.AzureChina
		default:
			cloudConfig = cloud.AzurePublic
		}

		model := provider.MSGraphProviderModel{}

		// set the defaults from environment variables
		if v := os.Getenv("ARM_CLIENT_ID"); v != "" {
			model.ClientID = types.StringValue(v)
		}
		if v := os.Getenv("ARM_CLIENT_ID_FILE_PATH"); v != "" {
			model.ClientIDFilePath = types.StringValue(v)
		}

		if v := os.Getenv("ARM_USE_AKS_WORKLOAD_IDENTITY"); v != "" {
			model.UseAKSWorkloadIdentity = types.BoolValue(v == "true")
		} else {
			model.UseAKSWorkloadIdentity = types.BoolValue(false)
		}
		if v := os.Getenv("ARM_TENANT_ID"); v != "" {
			model.TenantID = types.StringValue(v)
		}
		if model.UseAKSWorkloadIdentity.ValueBool() && os.Getenv("AZURE_TENANT_ID") != "" {
			aksTenantID := os.Getenv("AZURE_TENANT_ID")
			if model.TenantID.ValueString() != "" && model.TenantID.ValueString() != aksTenantID {
				return nil, fmt.Errorf("invalid `tenant_id` value: mismatch between supplied Tenant ID and that provided by AKS Workload Identity - please remove, ensure they match, or disable use_aks_workload_identity")
			}
			model.TenantID = types.StringValue(aksTenantID)
		}

		if v := os.Getenv("ARM_CLIENT_CERTIFICATE"); v != "" {
			model.ClientCertificate = types.StringValue(v)
		}
		if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PATH"); v != "" {
			model.ClientCertificatePath = types.StringValue(v)
		}
		if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PASSWORD"); v != "" {
			model.ClientCertificatePassword = types.StringValue(v)
		}
		if v := os.Getenv("ARM_CLIENT_SECRET"); v != "" {
			model.ClientSecret = types.StringValue(v)
		}
		if v := os.Getenv("ARM_CLIENT_SECRET_FILE_PATH"); v != "" {
			model.ClientSecretFilePath = types.StringValue(v)
		}
		if v := os.Getenv("ARM_OIDC_REQUEST_TOKEN"); v != "" {
			model.OIDCRequestToken = types.StringValue(v)
		} else if v := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN"); v != "" {
			model.OIDCRequestToken = types.StringValue(v)
		}
		if v := os.Getenv("ARM_OIDC_REQUEST_URL"); v != "" {
			model.OIDCRequestURL = types.StringValue(v)
		} else if v := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL"); v != "" {
			model.OIDCRequestURL = types.StringValue(v)
		}
		if v := os.Getenv("ARM_OIDC_TOKEN"); v != "" {
			model.OIDCToken = types.StringValue(v)
		}
		if v := os.Getenv("ARM_OIDC_TOKEN_FILE_PATH"); v != "" {
			model.OIDCTokenFilePath = types.StringValue(v)
		}
		if v := os.Getenv("ARM_OIDC_AZURE_SERVICE_CONNECTION_ID"); v != "" {
			model.OIDCAzureServiceConnectionID = types.StringValue(v)
		}
		if v := os.Getenv("ARM_USE_OIDC"); v != "" {
			model.UseOIDC = types.BoolValue(v == "true")
		} else {
			model.UseOIDC = types.BoolValue(false)
		}
		if v := os.Getenv("ARM_USE_CLI"); v != "" {
			model.UseCLI = types.BoolValue(v == "true")
		} else {
			model.UseCLI = types.BoolValue(true)
		}

		option := azidentity.DefaultAzureCredentialOptions{
			TenantID: model.TenantID.ValueString(),
		}
		cred, err := provider.BuildChainedTokenCredential(model, option)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain a credential: %v", err)
		}

		copt := &clients.Option{
			Cred:     cred,
			CloudCfg: cloudConfig,
			TenantId: os.Getenv("ARM_TENANT_ID"),
		}

		client := &clients.Client{}
		if err := client.Build(context.TODO(), copt); err != nil {
			return nil, err
		}
		_client = client
	}

	return _client, nil
}
