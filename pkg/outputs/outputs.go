package outputs

import (
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kafkakubernetes"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/kubernetes"
	"github.com/plantoncloud/stack-job-runner-golang-sdk/pkg/automationapi/autoapistackoutput"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

const (
	Namespace                        = "namespace"
	IngressExternalBootStrapHostname = "external-bootstrap-hostname"
	IngressInternalBootStrapHostname = "internal-bootstrap-hostname"
	IngressExternalSchemaRegistryUrl = "external-schema-registry-url"
	IngressInternalSchemaRegistryUrl = "internal-schema-registry-url"
	IngressKafkaUiExternalUrl        = "kafka-ui-ingress-external-url"
	KafkaAdminUsername               = "kafka-admin-username"
	KafkaAdminPasswordSecretName     = "kafka-admin-user-secret-name"
	KafkaAdminPasswordSecretKey      = "kafka-admin-user-secret-key"
)

func PulumiOutputsToStackOutputsConverter(pulumiOutputs auto.OutputMap,
	input *kafkakubernetes.KafkaKubernetesStackInput) *kafkakubernetes.KafkaKubernetesStackOutputs {
	return &kafkakubernetes.KafkaKubernetesStackOutputs{
		Namespace:                       autoapistackoutput.GetVal(pulumiOutputs, Namespace),
		BootstrapServerExternalHostname: autoapistackoutput.GetVal(pulumiOutputs, IngressExternalBootStrapHostname),
		BootstrapServerInternalHostname: autoapistackoutput.GetVal(pulumiOutputs, IngressInternalBootStrapHostname),
		SchemaRegistryExternalUrl:       autoapistackoutput.GetVal(pulumiOutputs, IngressExternalSchemaRegistryUrl),
		SchemaRegistryInternalUrl:       autoapistackoutput.GetVal(pulumiOutputs, IngressInternalSchemaRegistryUrl),
		Username:                        autoapistackoutput.GetVal(pulumiOutputs, KafkaAdminUsername),
		KafkaUiExternalUrl:              autoapistackoutput.GetVal(pulumiOutputs, IngressKafkaUiExternalUrl),
		PasswordSecret: &kubernetes.KubernernetesSecretKey{
			Name: autoapistackoutput.GetVal(pulumiOutputs, KafkaAdminPasswordSecretName),
			Key:  autoapistackoutput.GetVal(pulumiOutputs, KafkaAdminPasswordSecretKey),
		},
	}
}
