package adminuser

import (
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	pulk8scv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	resourceId    string
	labels        map[string]string
	workspaceDir  string
	namespace     *pulk8scv1.Namespace
	namespaceName string
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		resourceId:    contextState.Spec.ResourceId,
		labels:        contextState.Spec.Labels,
		workspaceDir:  contextState.Spec.WorkspaceDir,
		namespace:     contextState.Status.AddedResources.Namespace,
		namespaceName: contextState.Spec.NamespaceName,
	}
}
