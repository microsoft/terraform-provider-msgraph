package provider

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
)

const (
	// Canonical environment names (kept consistent with the AzureAD provider).
	EnvironmentPublic         = "public"
	EnvironmentUSGovernmentL4 = "usgovernmentl4"
	EnvironmentUSGovernmentL5 = "usgovernmentl5"
	EnvironmentChina          = "china"

	// Aliases accepted for the canonical names above.
	EnvironmentGlobal       = "global"
	EnvironmentUSGovernment = "usgovernment"
	EnvironmentDoD          = "dod"

	GraphHostPublic         = "https://graph.microsoft.com"
	GraphHostUSGovernmentL4 = "https://graph.microsoft.us"
	GraphHostUSGovernmentL5 = "https://dod-graph.microsoft.us"
	GraphHostChina          = "https://microsoftgraph.chinacloudapi.cn"
)

// SupportedEnvironments contains every value accepted for the `environment` argument,
// including the aliases that mirror the AzureAD provider. The order is stable so it can
// back both the schema validator and the documentation.
var SupportedEnvironments = []string{
	EnvironmentPublic,
	EnvironmentGlobal,
	EnvironmentUSGovernment,
	EnvironmentUSGovernmentL4,
	EnvironmentDoD,
	EnvironmentUSGovernmentL5,
	EnvironmentChina,
}

// Environments is the single source of truth mapping every accepted environment name
// (canonical values and their AzureAD-compatible aliases) to a fully resolved
// cloud.Configuration that embeds the Microsoft Graph service endpoint and audience.
var Environments = map[string]cloud.Configuration{
	EnvironmentPublic:         withGraphService(cloud.AzurePublic, GraphHostPublic),
	EnvironmentGlobal:         withGraphService(cloud.AzurePublic, GraphHostPublic),
	EnvironmentUSGovernment:   withGraphService(cloud.AzureGovernment, GraphHostUSGovernmentL4),
	EnvironmentUSGovernmentL4: withGraphService(cloud.AzureGovernment, GraphHostUSGovernmentL4),
	EnvironmentDoD:            withGraphService(cloud.AzureGovernment, GraphHostUSGovernmentL5),
	EnvironmentUSGovernmentL5: withGraphService(cloud.AzureGovernment, GraphHostUSGovernmentL5),
	EnvironmentChina:          withGraphService(cloud.AzureChina, GraphHostChina),
}

// ResolveCloudEnvironment returns the cloud.Configuration for the given environment
// name. An empty value defaults to the public cloud.
func ResolveCloudEnvironment(environment string) (cloud.Configuration, error) {
	name := strings.ToLower(strings.TrimSpace(environment))
	if name == "" {
		name = EnvironmentPublic
	}
	cfg, ok := Environments[name]
	if !ok {
		return cloud.Configuration{}, fmt.Errorf("unsupported environment %q. Supported values are: %s", environment, strings.Join(SupportedEnvironments, ", "))
	}
	return cfg, nil
}

// withGraphService clones cfg and embeds the Microsoft Graph endpoint/audience so that
// cloud.Configuration remains the single source of truth for each environment.
func withGraphService(cfg cloud.Configuration, graphHost string) cloud.Configuration {
	services := make(map[cloud.ServiceName]cloud.ServiceConfiguration, len(cfg.Services)+1)
	for name, service := range cfg.Services {
		services[name] = service
	}
	services[clients.MicrosoftGraph] = cloud.ServiceConfiguration{
		Endpoint: graphHost,
		Audience: graphHost,
	}
	cfg.Services = services
	return cfg
}
