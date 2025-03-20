// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/azure/terraform-provider-msgraph/internal/clients"
	"github.com/azure/terraform-provider-msgraph/internal/myvalidator"
	"github.com/azure/terraform-provider-msgraph/internal/services"
	"github.com/azure/terraform-provider-msgraph/version"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &MSGraphProvider{}

type MSGraphProvider struct{}

type MSGraphProviderModel struct {
	ClientID                     types.String `tfsdk:"client_id"`
	ClientIDFilePath             types.String `tfsdk:"client_id_file_path"`
	TenantID                     types.String `tfsdk:"tenant_id"`
	ClientCertificatePath        types.String `tfsdk:"client_certificate_path"`
	ClientCertificate            types.String `tfsdk:"client_certificate"`
	ClientCertificatePassword    types.String `tfsdk:"client_certificate_password"`
	ClientSecret                 types.String `tfsdk:"client_secret"`
	ClientSecretFilePath         types.String `tfsdk:"client_secret_file_path"`
	OIDCRequestToken             types.String `tfsdk:"oidc_request_token"`
	OIDCRequestURL               types.String `tfsdk:"oidc_request_url"`
	OIDCToken                    types.String `tfsdk:"oidc_token"`
	OIDCTokenFilePath            types.String `tfsdk:"oidc_token_file_path"`
	OIDCAzureServiceConnectionID types.String `tfsdk:"oidc_azure_service_connection_id"`
	UseOIDC                      types.Bool   `tfsdk:"use_oidc"`
	UseCLI                       types.Bool   `tfsdk:"use_cli"`
	UseMSI                       types.Bool   `tfsdk:"use_msi"`
	UseAKSWorkloadIdentity       types.Bool   `tfsdk:"use_aks_workload_identity"`
	PartnerID                    types.String `tfsdk:"partner_id"`
	CustomCorrelationRequestID   types.String `tfsdk:"custom_correlation_request_id"`
	DisableCorrelationRequestID  types.Bool   `tfsdk:"disable_correlation_request_id"`
	DisableTerraformPartnerID    types.Bool   `tfsdk:"disable_terraform_partner_id"`
}

func New() func() provider.Provider {
	return func() provider.Provider {
		return &MSGraphProvider{}
	}
}

func (model MSGraphProviderModel) GetClientId() (*string, error) {
	clientId := strings.TrimSpace(model.ClientID.ValueString())

	if path := model.ClientIDFilePath.ValueString(); path != "" {
		// #nosec G304
		fileClientIdRaw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading Client ID from file %q: %v", path, err)
		}

		fileClientId := strings.TrimSpace(string(fileClientIdRaw))

		if clientId != "" && clientId != fileClientId {
			return nil, fmt.Errorf("mismatch between supplied Client ID and supplied Client ID file contents - please either remove one or ensure they match")
		}

		clientId = fileClientId
	}

	if model.UseAKSWorkloadIdentity.ValueBool() && os.Getenv("AZURE_CLIENT_ID") != "" {
		aksClientId := os.Getenv("AZURE_CLIENT_ID")
		if clientId != "" && clientId != aksClientId {
			return nil, fmt.Errorf("mismatch between supplied Client ID and that provided by AKS Workload Identity - please remove, ensure they match, or disable use_aks_workload_identity")
		}
		clientId = aksClientId
	}

	return &clientId, nil
}

func (model MSGraphProviderModel) GetClientSecret() (*string, error) {
	clientSecret := strings.TrimSpace(model.ClientSecret.ValueString())

	if path := model.ClientSecretFilePath.ValueString(); path != "" {
		// #nosec G304
		fileSecretRaw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading Client Secret from file %q: %v", path, err)
		}

		fileSecret := strings.TrimSpace(string(fileSecretRaw))

		if clientSecret != "" && clientSecret != fileSecret {
			return nil, fmt.Errorf("mismatch between supplied Client Secret and supplied Client Secret file contents - please either remove one or ensure they match")
		}

		clientSecret = fileSecret
	}

	return &clientSecret, nil
}

func (model MSGraphProviderModel) GetOIDCTokenFilePath() string {
	if !model.OIDCTokenFilePath.IsNull() && model.OIDCTokenFilePath.ValueString() != "" {
		return model.OIDCTokenFilePath.ValueString()
	}

	if model.UseAKSWorkloadIdentity.ValueBool() && os.Getenv("AZURE_FEDERATED_TOKEN_FILE") != "" {
		return os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	}

	return ""
}

func (p *MSGraphProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "msgraph"
}

func (p *MSGraphProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "MSGraph provider",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Client ID which should be used. This can also be sourced from the `ARM_CLIENT_ID` Environment Variable.",
			},

			"client_id_file_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The path to a file containing the Client ID which should be used. This can also be sourced from the `ARM_CLIENT_ID_FILE_PATH` Environment Variable.",
			},

			"tenant_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Tenant ID should be used. This can also be sourced from the `ARM_TENANT_ID` Environment Variable.",
			},

			// Client Certificate specific fields
			"client_certificate_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The path to the Client Certificate associated with the Service Principal which should be used. This can also be sourced from the `ARM_CLIENT_CERTIFICATE_PATH` Environment Variable.",
			},

			"client_certificate": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A base64-encoded PKCS#12 bundle to be used as the client certificate for authentication. This can also be sourced from the `ARM_CLIENT_CERTIFICATE` environment variable.",
			},

			"client_certificate_password": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The password associated with the Client Certificate. This can also be sourced from the `ARM_CLIENT_CERTIFICATE_PASSWORD` Environment Variable.",
			},

			// Client Secret specific fields
			"client_secret": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Client Secret which should be used. This can also be sourced from the `ARM_CLIENT_SECRET` Environment Variable.",
			},

			"client_secret_file_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The path to a file containing the Client Secret which should be used. For use When authenticating as a Service Principal using a Client Secret. This can also be sourced from the `ARM_CLIENT_SECRET_FILE_PATH` Environment Variable.",
			},

			// OIDC specific fields
			"oidc_request_token": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The bearer token for the request to the OIDC provider. This can also be sourced from the `ARM_OIDC_REQUEST_TOKEN` or `ACTIONS_ID_TOKEN_REQUEST_TOKEN` Environment Variables.",
			},

			"oidc_request_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The URL for the OIDC provider from which to request an ID token. This can also be sourced from the `ARM_OIDC_REQUEST_URL` or `ACTIONS_ID_TOKEN_REQUEST_URL` Environment Variables.",
			},

			"oidc_token": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID token when authenticating using OpenID Connect (OIDC). This can also be sourced from the `ARM_OIDC_TOKEN` environment Variable.",
			},

			"oidc_token_file_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The path to a file containing an ID token when authenticating using OpenID Connect (OIDC). This can also be sourced from the `ARM_OIDC_TOKEN_FILE_PATH` environment Variable.",
			},

			"oidc_azure_service_connection_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Azure Pipelines Service Connection ID to use for authentication. This can also be sourced from the `ARM_OIDC_AZURE_SERVICE_CONNECTION_ID` environment variable.",
			},

			"use_oidc": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Should OIDC be used for Authentication? This can also be sourced from the `ARM_USE_OIDC` Environment Variable. Defaults to `false`.",
			},

			// Azure CLI specific fields
			"use_cli": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Should Azure CLI be used for authentication? This can also be sourced from the `ARM_USE_CLI` environment variable. Defaults to `true`.",
			},

			// Managed Service Identity specific fields
			"use_msi": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Should Managed Identity be used for Authentication? This can also be sourced from the `ARM_USE_MSI` Environment Variable. Defaults to `false`.",
			},

			"use_aks_workload_identity": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Should AKS Workload Identity be used for Authentication? This can also be sourced from the `ARM_USE_AKS_WORKLOAD_IDENTITY` Environment Variable. Defaults to `false`. When set, `client_id`, `tenant_id` and `oidc_token_file_path` will be detected from the environment and do not need to be specified.",
			},

			// Managed Tracking GUID for User-agent
			"partner_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.Any(myvalidator.StringIsUUID(), myvalidator.StringIsEmpty()),
				},
				MarkdownDescription: "A GUID/UUID that is [registered](https://docs.microsoft.com/azure/marketplace/azure-partner-customer-usage-attribution#register-guids-and-offers) with Microsoft to facilitate partner resource usage attribution. This can also be sourced from the `ARM_PARTNER_ID` Environment Variable.",
			},

			"custom_correlation_request_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The value of the `x-ms-correlation-request-id` header, otherwise an auto-generated UUID will be used. This can also be sourced from the `ARM_CORRELATION_REQUEST_ID` environment variable.",
			},

			"disable_correlation_request_id": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "This will disable the x-ms-correlation-request-id header.",
			},

			"disable_terraform_partner_id": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Disable sending the Terraform Partner ID if a custom `partner_id` isn't specified, which allows Microsoft to better understand the usage of Terraform. The Partner ID does not give HashiCorp any direct access to usage information. This can also be sourced from the `ARM_DISABLE_TERRAFORM_PARTNER_ID` environment variable. Defaults to `false`.",
			},
		},
	}
}

func (p *MSGraphProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model MSGraphProviderModel
	if resp.Diagnostics.Append(req.Config.Get(ctx, &model)...); resp.Diagnostics.HasError() {
		return
	}

	// set the defaults from environment variables
	if model.ClientID.IsNull() {
		if v := os.Getenv("ARM_CLIENT_ID"); v != "" {
			model.ClientID = types.StringValue(v)
		}
	}
	if model.ClientIDFilePath.IsNull() {
		if v := os.Getenv("ARM_CLIENT_ID_FILE_PATH"); v != "" {
			model.ClientIDFilePath = types.StringValue(v)
		}
	}

	if model.UseAKSWorkloadIdentity.IsNull() {
		if v := os.Getenv("ARM_USE_AKS_WORKLOAD_IDENTITY"); v != "" {
			model.UseAKSWorkloadIdentity = types.BoolValue(v == "true")
		} else {
			model.UseAKSWorkloadIdentity = types.BoolValue(false)
		}
	}

	if model.TenantID.IsNull() {
		if v := os.Getenv("ARM_TENANT_ID"); v != "" {
			model.TenantID = types.StringValue(v)
		}
		if model.UseAKSWorkloadIdentity.ValueBool() && os.Getenv("AZURE_TENANT_ID") != "" {
			aksTenantID := os.Getenv("AZURE_TENANT_ID")
			if model.TenantID.ValueString() != "" && model.TenantID.ValueString() != aksTenantID {
				resp.Diagnostics.AddError("Invalid `tenant_id` value", "mismatch between supplied Tenant ID and that provided by AKS Workload Identity - please remove, ensure they match, or disable use_aks_workload_identity")
				return
			}
			model.TenantID = types.StringValue(aksTenantID)
		}
	}

	if model.ClientCertificate.IsNull() {
		if v := os.Getenv("ARM_CLIENT_CERTIFICATE"); v != "" {
			model.ClientCertificate = types.StringValue(v)
		}
	}

	if model.ClientCertificatePath.IsNull() {
		if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PATH"); v != "" {
			model.ClientCertificatePath = types.StringValue(v)
		}
	}

	if model.ClientCertificatePassword.IsNull() {
		if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PASSWORD"); v != "" {
			model.ClientCertificatePassword = types.StringValue(v)
		}
	}

	if model.ClientSecret.IsNull() {
		if v := os.Getenv("ARM_CLIENT_SECRET"); v != "" {
			model.ClientSecret = types.StringValue(v)
		}
	}

	if model.ClientSecretFilePath.IsNull() {
		if v := os.Getenv("ARM_CLIENT_SECRET_FILE_PATH"); v != "" {
			model.ClientSecretFilePath = types.StringValue(v)
		}
	}

	if model.OIDCRequestToken.IsNull() {
		if v := os.Getenv("ARM_OIDC_REQUEST_TOKEN"); v != "" {
			model.OIDCRequestToken = types.StringValue(v)
		} else if v := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN"); v != "" {
			model.OIDCRequestToken = types.StringValue(v)
		}
	}

	if model.OIDCRequestURL.IsNull() {
		if v := os.Getenv("ARM_OIDC_REQUEST_URL"); v != "" {
			model.OIDCRequestURL = types.StringValue(v)
		} else if v := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL"); v != "" {
			model.OIDCRequestURL = types.StringValue(v)
		}
	}

	if model.OIDCToken.IsNull() {
		if v := os.Getenv("ARM_OIDC_TOKEN"); v != "" {
			model.OIDCToken = types.StringValue(v)
		}
	}

	if model.OIDCTokenFilePath.IsNull() {
		if v := os.Getenv("ARM_OIDC_TOKEN_FILE_PATH"); v != "" {
			model.OIDCTokenFilePath = types.StringValue(v)
		}
	}

	if model.OIDCAzureServiceConnectionID.IsNull() {
		if v := os.Getenv("ARM_OIDC_AZURE_SERVICE_CONNECTION_ID"); v != "" {
			model.OIDCAzureServiceConnectionID = types.StringValue(v)
		}
	}

	if model.UseOIDC.IsNull() {
		if v := os.Getenv("ARM_USE_OIDC"); v != "" {
			model.UseOIDC = types.BoolValue(v == "true")
		} else {
			model.UseOIDC = types.BoolValue(false)
		}
	}

	if model.UseCLI.IsNull() {
		if v := os.Getenv("ARM_USE_CLI"); v != "" {
			model.UseCLI = types.BoolValue(v == "true")
		} else {
			model.UseCLI = types.BoolValue(true)
		}
	}

	if model.UseMSI.IsNull() {
		if v := os.Getenv("ARM_USE_MSI"); v != "" {
			model.UseMSI = types.BoolValue(v == "true")
		} else {
			model.UseMSI = types.BoolValue(false)
		}
	}

	if model.PartnerID.IsNull() {
		if v := os.Getenv("ARM_PARTNER_ID"); v != "" {
			model.PartnerID = types.StringValue(v)
		}
	}

	if model.CustomCorrelationRequestID.IsNull() {
		if v := os.Getenv("ARM_CORRELATION_REQUEST_ID"); v != "" {
			model.CustomCorrelationRequestID = types.StringValue(v)
		}
	}

	if model.DisableCorrelationRequestID.IsNull() {
		if v := os.Getenv("ARM_DISABLE_CORRELATION_REQUEST_ID"); v != "" {
			model.DisableCorrelationRequestID = types.BoolValue(v == "true")
		} else {
			model.DisableCorrelationRequestID = types.BoolValue(false)
		}
	}

	if model.DisableTerraformPartnerID.IsNull() {
		if v := os.Getenv("ARM_DISABLE_TERRAFORM_PARTNER_ID"); v != "" {
			model.DisableTerraformPartnerID = types.BoolValue(v == "true")
		} else {
			model.DisableTerraformPartnerID = types.BoolValue(false)
		}
	}

	option := azidentity.DefaultAzureCredentialOptions{
		TenantID: model.TenantID.ValueString(),
	}

	cred, err := BuildChainedTokenCredential(model, option)
	if err != nil {
		resp.Diagnostics.AddError("Failed to obtain a credential.", err.Error())
		return
	}

	copt := &clients.Option{
		Cred:                        cred,
		ApplicationUserAgent:        buildUserAgent(req.TerraformVersion, model.PartnerID.ValueString(), model.DisableTerraformPartnerID.ValueBool()),
		DisableCorrelationRequestID: model.DisableCorrelationRequestID.ValueBool(),
		CustomCorrelationRequestID:  model.CustomCorrelationRequestID.ValueString(),
		CloudCfg:                    cloud.Configuration{},
		TenantId:                    model.TenantID.ValueString(),
	}
	client := &clients.Client{}
	if err = client.Build(ctx, copt); err != nil {
		resp.Diagnostics.AddError("Error Building Client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *MSGraphProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		services.NewMSGraphResource,
	}
}

func (p *MSGraphProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		services.NewMSGraphDataSource,
	}
}

func buildUserAgent(terraformVersion string, partnerID string, disableTerraformPartnerID bool) string {
	if terraformVersion == "" {
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing it's 0.10 or 0.11
		terraformVersion = "0.11+compatible"
	}
	tfUserAgent := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io)", terraformVersion)
	providerUserAgent := fmt.Sprintf("terraform-provider-msgraph/%s", version.ProviderVersion)
	userAgent := strings.TrimSpace(fmt.Sprintf("%s %s", tfUserAgent, providerUserAgent))

	// append the CloudShell version to the user agent if it exists
	if azureAgent := os.Getenv("AZURE_HTTP_USER_AGENT"); azureAgent != "" {
		userAgent = fmt.Sprintf("%s %s", userAgent, azureAgent)
	}

	// only one pid can be interpreted currently
	// hence, send partner ID if present, otherwise send Terraform GUID
	// unless users have opted out
	if partnerID == "" && !disableTerraformPartnerID {
		// Microsoft’s Terraform Partner ID is this specific GUID
		partnerID = "222c6c49-1b0a-5959-a213-6608f9eb8820"
	}

	if partnerID != "" {
		userAgent = fmt.Sprintf("%s pid-%s", userAgent, partnerID)
	}
	return userAgent
}

func BuildChainedTokenCredential(model MSGraphProviderModel, options azidentity.DefaultAzureCredentialOptions) (*azidentity.ChainedTokenCredential, error) {
	log.Printf("[DEBUG] building chained token credential")
	var creds []azcore.TokenCredential

	if model.UseOIDC.ValueBool() || model.UseAKSWorkloadIdentity.ValueBool() {
		log.Printf("[DEBUG] oidc credential or AKS Workload Identity enabled")
		if cred, err := buildOidcCredential(model, options); err == nil {
			creds = append(creds, cred)
		} else {
			log.Printf("[DEBUG] failed to initialize oidc credential: %v", err)
		}

		log.Printf("[DEBUG] azure pipelines credential enabled")
		if cred, err := buildAzurePipelinesCredential(model, options); err == nil {
			creds = append(creds, cred)
		} else {
			log.Printf("[DEBUG] failed to initialize azure pipelines credential: %v", err)
		}
	}

	if cred, err := buildClientSecretCredential(model, options); err == nil {
		creds = append(creds, cred)
	} else {
		log.Printf("[DEBUG] failed to initialize client secret credential: %v", err)
	}

	if cred, err := buildClientCertificateCredential(model, options); err == nil {
		creds = append(creds, cred)
	} else {
		log.Printf("[DEBUG] failed to initialize client certificate credential: %v", err)
	}

	if model.UseMSI.ValueBool() {
		log.Printf("[DEBUG] msi credential enabled")
		if cred, err := buildManagedIdentityCredential(model, options); err == nil {
			creds = append(creds, cred)
		} else {
			log.Printf("[DEBUG] failed to initialize msi credential: %v", err)
		}
	}

	if model.UseCLI.ValueBool() {
		log.Printf("[DEBUG] cli credential enabled")
		if cred, err := buildAzureCLICredential(options); err == nil {
			creds = append(creds, cred)
		} else {
			log.Printf("[DEBUG] failed to initialize cli credential: %v", err)
		}
	}

	if len(creds) == 0 {
		return nil, fmt.Errorf("no credentials were successfully initialized")
	}

	return azidentity.NewChainedTokenCredential(creds, nil)
}

func buildClientSecretCredential(model MSGraphProviderModel, options azidentity.DefaultAzureCredentialOptions) (azcore.TokenCredential, error) {
	log.Printf("[DEBUG] building client secret credential")
	clientID, err := model.GetClientId()
	if err != nil {
		return nil, err
	}
	clientSecret, err := model.GetClientSecret()
	if err != nil {
		return nil, err
	}
	o := &azidentity.ClientSecretCredentialOptions{
		AdditionallyAllowedTenants: options.AdditionallyAllowedTenants,
		ClientOptions:              options.ClientOptions,
		DisableInstanceDiscovery:   options.DisableInstanceDiscovery,
	}
	return azidentity.NewClientSecretCredential(options.TenantID, *clientID, *clientSecret, o)
}

func buildClientCertificateCredential(model MSGraphProviderModel, options azidentity.DefaultAzureCredentialOptions) (azcore.TokenCredential, error) {
	log.Printf("[DEBUG] building client certificate credential")
	clientID, err := model.GetClientId()
	if err != nil {
		return nil, err
	}

	var certData []byte
	if certPath := model.ClientCertificatePath.ValueString(); certPath != "" {
		log.Printf("[DEBUG] reading certificate from file %s", certPath)
		// #nosec G304
		certData, err = os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf(`failed to read certificate file "%s": %v`, certPath, err)
		}
	}
	if certBase64 := model.ClientCertificate.ValueString(); certBase64 != "" {
		log.Printf("[DEBUG] decoding certificate from base64")
		certData, err = decodeCertificate(certBase64)
		if err != nil {
			return nil, err
		}
	}

	if len(certData) == 0 {
		return nil, fmt.Errorf("no certificate data provided")
	}

	var password []byte
	if v := model.ClientCertificatePassword.ValueString(); v != "" {
		password = []byte(v)
	}
	certs, key, err := azidentity.ParseCertificates(certData, password)
	if err != nil {
		return nil, fmt.Errorf(`failed to load certificate": %v`, err)
	}
	o := &azidentity.ClientCertificateCredentialOptions{
		AdditionallyAllowedTenants: options.AdditionallyAllowedTenants,
		ClientOptions:              options.ClientOptions,
		DisableInstanceDiscovery:   options.DisableInstanceDiscovery,
	}
	return azidentity.NewClientCertificateCredential(options.TenantID, *clientID, certs, key, o)
}

func buildOidcCredential(model MSGraphProviderModel, options azidentity.DefaultAzureCredentialOptions) (azcore.TokenCredential, error) {
	log.Printf("[DEBUG] building oidc credential")
	clientId, err := model.GetClientId()
	if err != nil {
		return nil, err
	}
	if model.OIDCToken.ValueString() == "" && model.GetOIDCTokenFilePath() == "" && (model.OIDCRequestToken.ValueString() == "" || model.OIDCRequestURL.ValueString() == "") {
		return nil, fmt.Errorf("missing required OIDC configuration")
	}
	o := &OidcCredentialOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: options.Cloud,
		},
		AdditionallyAllowedTenants: options.AdditionallyAllowedTenants,
		TenantID:                   options.TenantID,
		ClientID:                   *clientId,
		RequestToken:               model.OIDCRequestToken.ValueString(),
		RequestUrl:                 model.OIDCRequestURL.ValueString(),
		Token:                      model.OIDCToken.ValueString(),
		TokenFilePath:              model.GetOIDCTokenFilePath(),
	}
	return NewOidcCredential(o)
}

func buildManagedIdentityCredential(model MSGraphProviderModel, options azidentity.DefaultAzureCredentialOptions) (azcore.TokenCredential, error) {
	log.Printf("[DEBUG] building managed identity credential")
	clientId, err := model.GetClientId()
	if err != nil {
		return nil, err
	}
	o := &azidentity.ManagedIdentityCredentialOptions{
		ClientOptions: options.ClientOptions,
		ID:            azidentity.ClientID(*clientId),
	}
	return NewManagedIdentityCredential(o)
}

func buildAzureCLICredential(options azidentity.DefaultAzureCredentialOptions) (azcore.TokenCredential, error) {
	log.Printf("[DEBUG] building azure cli credential")
	o := &azidentity.AzureCLICredentialOptions{
		AdditionallyAllowedTenants: options.AdditionallyAllowedTenants,
		TenantID:                   options.TenantID,
	}
	return azidentity.NewAzureCLICredential(o)
}

func buildAzurePipelinesCredential(model MSGraphProviderModel, options azidentity.DefaultAzureCredentialOptions) (azcore.TokenCredential, error) {
	log.Printf("[DEBUG] building azure pipeline credential")
	o := &azidentity.AzurePipelinesCredentialOptions{
		ClientOptions:              options.ClientOptions,
		AdditionallyAllowedTenants: options.AdditionallyAllowedTenants,
		DisableInstanceDiscovery:   options.DisableInstanceDiscovery,
	}
	clientId, err := model.GetClientId()
	if err != nil {
		return nil, err
	}
	return azidentity.NewAzurePipelinesCredential(options.TenantID, *clientId, model.OIDCAzureServiceConnectionID.ValueString(), model.OIDCRequestToken.ValueString(), o)
}

func decodeCertificate(clientCertificate string) ([]byte, error) {
	var pfx []byte
	if clientCertificate != "" {
		out := make([]byte, base64.StdEncoding.DecodedLen(len(clientCertificate)))
		n, err := base64.StdEncoding.Decode(out, []byte(clientCertificate))
		if err != nil {
			return pfx, fmt.Errorf("could not decode client certificate data: %v", err)
		}
		pfx = out[:n]
	}
	return pfx, nil
}
