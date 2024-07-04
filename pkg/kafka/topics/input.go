package topics

import (
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	kafkakubernetesstatemodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	kubernetesProvider *pulumikubernetes.Provider
	labels             map[string]string
	namespaceName      string
	workspaceDir       string
	resourceId         string
	topics             []*kafkakubernetesstatemodel.KafkaTopic
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		kubernetesProvider: contextState.Spec.KubeProvider,
		labels:             contextState.Spec.Labels,
		namespaceName:      contextState.Spec.NamespaceName,
		workspaceDir:       contextState.Spec.WorkspaceDir,
		resourceId:         contextState.Spec.ResourceId,
		topics:             contextState.Spec.Topics,
	}
}
