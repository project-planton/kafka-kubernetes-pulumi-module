package outputs

import (
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kafkakubernetes/model"
	"github.com/plantoncloud/stack-job-runner-golang-sdk/pkg/automationapi/autoapistackoutput"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

const (
	Namespace                        = "namespace"
	IngressExternalBootStrapHostname = "external-bootstrap-hostname"
	IngressInternalBootStrapHostname = "internal-bootstrap-hostname"
	IngressExternalSchemaRegistryUrl = "external-schema-registry-url"
	IngressInternalSchemaRegistryUrl = "internal-schema-registry-url"
	IngressExternalKowlUrl           = "external-kowl-url"
	KafkaSaslUsername                = "kafka-sasl-username"
)

func PulumiOutputsToStackOutputsConverter(pulumiOutputs auto.OutputMap,
	input *model.KafkaKubernetesStackInput) *model.KafkaKubernetesStackOutputs {
	return &model.KafkaKubernetesStackOutputs{
		Namespace:                       autoapistackoutput.GetVal(pulumiOutputs, Namespace),
		KafkaSaslUsername:               autoapistackoutput.GetVal(pulumiOutputs, KafkaSaslUsername),
		ExternalBootstrapServerHostname: autoapistackoutput.GetVal(pulumiOutputs, IngressExternalBootStrapHostname),
		InternalBootstrapServerHostname: autoapistackoutput.GetVal(pulumiOutputs, IngressInternalBootStrapHostname),
		ExternalSchemaRegistryUrl:       autoapistackoutput.GetVal(pulumiOutputs, IngressExternalSchemaRegistryUrl),
		InternalSchemaRegistryUrl:       autoapistackoutput.GetVal(pulumiOutputs, IngressInternalSchemaRegistryUrl),
		ExternalKowlDashboardUrl:        autoapistackoutput.GetVal(pulumiOutputs, IngressExternalKowlUrl),
	}
}
