package ingress

import (
	schemaregistrydeployment "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/schemaregistry/deployment"
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// KubeServiceName https://github.com/confluentinc/schema-registry/issues/689#issuecomment-354485274
	//warning: service name should not be "schema-registry"
	KubeServiceName = "sr"
)

type input struct {
	namespace                    *pulumikubernetescorev1.Namespace
	schemaRegistryDeploymentName string
	kafkaKubernetesId            string
	namespaceName                string
	workspaceDir                 string
	environmentName              string
	kafkaIngressDomain           string
	externalIngressHostname      string
	internalIngressHostname      string
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		namespace:                    contextState.Status.AddedResources.Namespace,
		schemaRegistryDeploymentName: schemaregistrydeployment.Name,
		kafkaKubernetesId:            contextState.Spec.ResourceId,
		namespaceName:                contextState.Spec.NamespaceName,
		workspaceDir:                 contextState.Spec.WorkspaceDir,
		environmentName:              contextState.Spec.EnvironmentInfo.EnvironmentName,
		kafkaIngressDomain:           contextState.Spec.EndpointDomainName,
		externalIngressHostname:      contextState.Spec.ExternalSchemaRegistryHostname,
		internalIngressHostname:      contextState.Spec.InternalSchemaRegistryHostname,
	}
}
