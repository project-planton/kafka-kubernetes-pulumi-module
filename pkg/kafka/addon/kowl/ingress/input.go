package ingress

import (
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/kowl/deployment"
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	namespace                     *pulumikubernetescorev1.Namespace
	kafkaKubernetesId             string
	kowlDeploymentName            string
	namespaceName                 string
	workspaceDir                  string
	kafkaIngressDomain            string
	environmentName               string
	externalKowlDashboardHostname string
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		namespace:                     contextState.Status.AddedResources.Namespace,
		kafkaKubernetesId:             contextState.Spec.ResourceId,
		kowlDeploymentName:            deployment.Name,
		namespaceName:                 contextState.Spec.NamespaceName,
		workspaceDir:                  contextState.Spec.WorkspaceDir,
		kafkaIngressDomain:            contextState.Spec.EndpointDomainName,
		environmentName:               contextState.Spec.EnvironmentInfo.EnvironmentName,
		externalKowlDashboardHostname: contextState.Spec.ExternalKowlDashboardHostname,
	}
}
