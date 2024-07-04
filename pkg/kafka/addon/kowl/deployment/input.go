package deployment

import (
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	Name = "kowl"
)

type input struct {
	labels             map[string]string
	kowlDeploymentName string
	namespaceName      string
	kowlConfigMap      *pulumikubernetescorev1.ConfigMap
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		labels:             contextState.Spec.Labels,
		kowlDeploymentName: Name,
		namespaceName:      contextState.Spec.NamespaceName,
		kowlConfigMap:      contextState.Status.AddedResources.KowlConfigMap,
	}
}
