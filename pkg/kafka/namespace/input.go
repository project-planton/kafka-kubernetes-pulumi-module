package namespace

import (
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	namespaceName string
	labels        map[string]string
	kubeProvider  *pulumikubernetes.Provider
}

func extractInput(ctx *pulumi.Context) *input {
	var ctxConfig = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		namespaceName: ctxConfig.Spec.NamespaceName,
		labels:        ctxConfig.Spec.Labels,
		kubeProvider:  ctxConfig.Spec.KubeProvider,
	}
}
