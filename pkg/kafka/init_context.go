package kafka

import (
	"github.com/pkg/errors"
	environmentblueprinthostnames "github.com/plantoncloud/environment-pulumi-blueprint/pkg/gcpgke/endpointdomains/hostnames"
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/network/istio/hostname"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	"github.com/plantoncloud/pulumi-blueprint-golang-commons/pkg/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func loadConfig(ctx *pulumi.Context, resourceStack *ResourceStack) (*kafkacontextstate.ContextState, error) {

	kubernetesProvider, err := pulumikubernetesprovider.GetWithStackCredentials(ctx, resourceStack.Input.CredentialsInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup kubernetes provider")
	}

	var resourceId = resourceStack.Input.ResourceInput.Metadata.Id
	var resourceName = resourceStack.Input.ResourceInput.Metadata.Name
	var environmentInfo = resourceStack.Input.ResourceInput.Spec.EnvironmentInfo
	var isIngressEnabled = false

	if resourceStack.Input.ResourceInput.Spec.Ingress != nil {
		isIngressEnabled = resourceStack.Input.ResourceInput.Spec.Ingress.IsEnabled
	}

	var endpointDomainName = ""
	var envDomainName = ""
	var ingressType = kubernetesworkloadingresstype.KubernetesWorkloadIngressType_unspecified
	var internalBootstrapHostname = ""
	var externalBootstrapHostname = ""
	var internalKowlDashboardHostname = ""
	var externalKowlDashboardHostname = ""
	var internalSchemaRegistryHostname = ""
	var externalSchemaRegistryHostname = ""

	if isIngressEnabled {
		endpointDomainName = resourceStack.Input.ResourceInput.Spec.Ingress.EndpointDomainName
		envDomainName = environmentblueprinthostnames.GetExternalEnvHostname(environmentInfo.EnvironmentName, endpointDomainName)
		ingressType = resourceStack.Input.ResourceInput.Spec.Ingress.IngressType

		internalBootstrapHostname = hostname.GetInternalBootstrapHostname(resourceId, environmentInfo.EnvironmentName, endpointDomainName)
		externalBootstrapHostname = hostname.GetExternalBootstrapHostname(resourceId, environmentInfo.EnvironmentName, endpointDomainName)
		internalKowlDashboardHostname = hostname.GetInteralKowlDashboardHostname(resourceId, environmentInfo.EnvironmentName, endpointDomainName)
		externalKowlDashboardHostname = hostname.GetExternalKowlDashboardHostname(resourceId, environmentInfo.EnvironmentName, endpointDomainName)
		internalSchemaRegistryHostname = hostname.GetInternalSchemaRegistryHostname(resourceId, environmentInfo.EnvironmentName, endpointDomainName)
		externalSchemaRegistryHostname = hostname.GetExternalSchemaRegistryHostname(resourceId, environmentInfo.EnvironmentName, endpointDomainName)
	}

	return &kafkacontextstate.ContextState{
		Spec: &kafkacontextstate.Spec{
			KubeProvider:                   kubernetesProvider,
			ResourceId:                     resourceId,
			ResourceName:                   resourceName,
			Labels:                         resourceStack.KubernetesLabels,
			EnvironmentInfo:                resourceStack.Input.ResourceInput.Spec.EnvironmentInfo,
			WorkspaceDir:                   resourceStack.WorkspaceDir,
			NamespaceName:                  resourceId,
			BrokerContainerSpec:            resourceStack.Input.ResourceInput.Spec.BrokerContainer,
			ZookeeperContainerSpec:         resourceStack.Input.ResourceInput.Spec.ZookeeperContainer,
			SchemaRegistryContainerSpec:    resourceStack.Input.ResourceInput.Spec.SchemaRegistryContainer,
			Topics:                         resourceStack.Input.ResourceInput.Spec.KafkaTopics,
			IsKowlDashboardEnabled:         resourceStack.Input.ResourceInput.Spec.IsKowlDashboardEnabled,
			IsSchemaRegistryEnabled:        resourceStack.Input.ResourceInput.Spec.SchemaRegistryContainer.IsEnabled,
			IsIngressEnabled:               isIngressEnabled,
			IngressType:                    ingressType,
			EndpointDomainName:             endpointDomainName,
			EnvDomainName:                  envDomainName,
			ExternalBootstrapHostname:      externalBootstrapHostname,
			InternalBootstrapHostname:      internalBootstrapHostname,
			InternalKowlDashboardHostname:  internalKowlDashboardHostname,
			ExternalKowlDashboardHostname:  externalKowlDashboardHostname,
			InternalSchemaRegistryHostname: internalSchemaRegistryHostname,
			ExternalSchemaRegistryHostname: externalSchemaRegistryHostname,
		},
		Status: &kafkacontextstate.Status{},
	}, nil
}
