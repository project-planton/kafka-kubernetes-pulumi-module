package template

import (
	schemaregistryingress "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/schemaregistry/ingress"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/adminuser"
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/listener"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	RefreshIntervalMinutes = 5
	kowlConfigFileTemplate = `
kafka:
  brokers:
    - {{.bootstrapServerHostname}}
  clientId: kowl-on-cluster
  sasl:
    enabled: true
    username: "{{.saslUsername}}"
    mechanism: SCRAM-SHA-512
  tls:
    enabled: true
  schemaRegistry:
    enabled: true
    urls: ["http://{{.schemaRegistryHostname}}"]
  protobuf:
    enabled: true
    schemaRegistry:
      enabled: true
      refreshInterval: {{.refreshIntervalMinutes}}m
`
)

type input struct {
	bootstrapServerHostname string
	bootstrapServerPort     int32
	saslUsername            string
	schemaRegistryHostname  string
	refreshIntervalMinutes  int32
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		bootstrapServerHostname: contextState.Spec.ExternalBootstrapHostname,
		bootstrapServerPort:     listener.ExternalPublicListenerPortNumber,
		saslUsername:            adminuser.Username,
		schemaRegistryHostname:  schemaregistryingress.GetKubeServiceNameFqdn(contextState.Spec.NamespaceName),
		refreshIntervalMinutes:  RefreshIntervalMinutes,
	}
}
