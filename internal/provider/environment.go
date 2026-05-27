package provider

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/microsoft/terraform-provider-msgraph/internal/clients"
)

const (
	EnvironmentPublic       = "public"
	EnvironmentChina        = "china"
	EnvironmentUSGovernment = "usgovernment"

	GraphHostPublic       = "https://graph.microsoft.com"
	GraphHostChina        = "https://microsoftgraph.chinacloudapi.cn"
	GraphHostUSGovernment = "https://graph.microsoft.us"
)

type CloudEnvironment struct {
	Name  string
	Cloud cloud.Configuration
}

func ResolveCloudEnvironment(environment string) (CloudEnvironment, error) {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "", EnvironmentPublic:
		return cloudEnvironment(EnvironmentPublic, cloud.AzurePublic, GraphHostPublic), nil
	case EnvironmentChina:
		return cloudEnvironment(EnvironmentChina, cloud.AzureChina, GraphHostChina), nil
	case EnvironmentUSGovernment:
		return cloudEnvironment(EnvironmentUSGovernment, cloud.AzureGovernment, GraphHostUSGovernment), nil
	default:
		return CloudEnvironment{}, fmt.Errorf("unsupported environment %q. Supported values are: public, china, usgovernment", environment)
	}
}

func cloudEnvironment(name string, cfg cloud.Configuration, graphHost string) CloudEnvironment {
	services := make(map[cloud.ServiceName]cloud.ServiceConfiguration, len(cfg.Services)+1)
	for name, service := range cfg.Services {
		services[name] = service
	}
	services[clients.MicrosoftGraph] = cloud.ServiceConfiguration{
		Endpoint: graphHost,
		Audience: graphHost,
	}
	cfg.Services = services

	return CloudEnvironment{
		Name:  name,
		Cloud: cfg,
	}
}
