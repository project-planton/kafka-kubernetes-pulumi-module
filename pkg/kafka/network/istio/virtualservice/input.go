package virtualservice

import (
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	kafkahostname "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/network/istio/hostname"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	namespace          *pulumikubernetescorev1.Namespace
	labels             map[string]string
	namespaceName      string
	hostnames          []string
	workspaceDir       string
	environmentName    string
	endpointDomainName string
	resourceId         string
	brokerReplicas     int32
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	hostnames := kafkahostname.GetKafkaHostnames(contextState.Spec.EnvironmentInfo.EnvironmentName, contextState.Spec.EndpointDomainName,
		contextState.Spec.ResourceId, contextState.Spec.BrokerContainerSpec.Replicas)

	return &input{
		namespace:          contextState.Status.AddedResources.Namespace,
		labels:             contextState.Spec.Labels,
		namespaceName:      contextState.Spec.NamespaceName,
		hostnames:          hostnames,
		workspaceDir:       contextState.Spec.WorkspaceDir,
		environmentName:    contextState.Spec.EnvironmentInfo.EnvironmentName,
		endpointDomainName: contextState.Spec.EndpointDomainName,
		resourceId:         contextState.Spec.ResourceId,
		brokerReplicas:     contextState.Spec.BrokerContainerSpec.Replicas,
	}
}
