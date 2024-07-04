package configmap

import (
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	Name          = "kowl"
	ConfigKeyName = "kowl.yaml"
)

type input struct {
	namespace *pulumikubernetescorev1.Namespace
	name      string
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		namespace: contextState.Status.AddedResources.Namespace,
		name:      Name,
	}
}
