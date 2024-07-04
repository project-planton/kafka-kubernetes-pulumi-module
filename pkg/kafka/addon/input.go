package addon

import (
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	isKowlDashboardEnabled  bool
	isSchemaRegistryEnabled bool
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		isKowlDashboardEnabled:  contextState.Spec.IsKowlDashboardEnabled,
		isSchemaRegistryEnabled: contextState.Spec.IsSchemaRegistryEnabled,
	}
}
