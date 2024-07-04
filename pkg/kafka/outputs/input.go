package outputs

import (
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	resourceId                     string
	environmentName                string
	endpointDomainName             string
	namespaceName                  string
	externalBootstrapHostname      string
	internalBootstrapHostname      string
	externalSchemaRegistryHostname string
	internalSchemaRegistryHostname string
	externalKowlDashboardHostname  string
	internalKowlDashboardHostname  string
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		resourceId:                     contextState.Spec.ResourceId,
		environmentName:                contextState.Spec.EnvironmentInfo.EnvironmentName,
		endpointDomainName:             contextState.Spec.EndpointDomainName,
		namespaceName:                  contextState.Spec.NamespaceName,
		externalBootstrapHostname:      contextState.Spec.ExternalBootstrapHostname,
		internalBootstrapHostname:      contextState.Spec.InternalBootstrapHostname,
		externalSchemaRegistryHostname: contextState.Spec.ExternalSchemaRegistryHostname,
		internalSchemaRegistryHostname: contextState.Spec.InternalSchemaRegistryHostname,
		externalKowlDashboardHostname:  contextState.Spec.ExternalKowlDashboardHostname,
		internalKowlDashboardHostname:  contextState.Spec.InternalKowlDashboardHostname,
	}
}
