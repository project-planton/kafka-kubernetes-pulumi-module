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
    - {{.BootstrapServerHostname}}
  clientId: kowl-on-cluster
  sasl:
    enabled: true
    username: "{{.SaslUsername}}"
    mechanism: SCRAM-SHA-512
  tls:
    enabled: true
  schemaRegistry:
    enabled: true
    urls: ["http://{{.SchemaRegistryHostname}}"]
  protobuf:
    enabled: true
    schemaRegistry:
      enabled: true
      refreshInterval: {{.RefreshIntervalMinutes}}m
`
)

type input struct {
	BootstrapServerHostname string
	BootstrapServerPort     int32
	SaslUsername            string
	SchemaRegistryHostname  string
	RefreshIntervalMinutes  int32
}

func extractInput(ctx *pulumi.Context) *input {
	var contextState = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	return &input{
		BootstrapServerHostname: contextState.Spec.ExternalBootstrapHostname,
		BootstrapServerPort:     listener.ExternalPublicListenerPortNumber,
		SaslUsername:            adminuser.Username,
		SchemaRegistryHostname:  schemaregistryingress.GetKubeServiceNameFqdn(contextState.Spec.NamespaceName),
		RefreshIntervalMinutes:  RefreshIntervalMinutes,
	}
}
