package kafka

import (
	"context"

	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/iac/v1/stackjob/enums/stackjoboperationtype"

	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/endpoint/virtualservice"
	productkfcv1clusterkubernetesstackimplnamespace "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/namespace"
	code2cloudv1deploykfcmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
	code2cloudv1deploykfcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/org"
	"github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/stack/output/backend"
)

func Outputs(ctx context.Context, input *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackInput) (*code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackOutputs, error) {
	pulumiOrgName, err := org.GetOrgName()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pulumi org name")
	}
	stackOutput, err := backend.StackOutput(pulumiOrgName, input.StackJob)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stack output")
	}
	return OutputMapTransformer(stackOutput, input), nil
}

func OutputMapTransformer(stackOutput map[string]interface{}, input *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackInput) *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackOutputs {
	if input.StackJob.Spec.OperationType != stackjoboperationtype.StackJobOperationType_apply || stackOutput == nil {
		return &code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackOutputs{}
	}
	return &code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackOutputs{
		KafkaKubernetesStatus: &code2cloudv1deploykfcmodel.KafkaKubernetesStatus{
			Kubernetes: &code2cloudv1deploykfcmodel.KafkaKubernetesStatusKubernetesStatus{
				Namespace:                       backend.GetVal(stackOutput, productkfcv1clusterkubernetesstackimplnamespace.GetNamespaceNameOutputName()),
				KafkaSaslUsername:               "admin",
				ExternalBootstrapServerHostname: backend.GetVal(stackOutput, virtualservice.GetExternalBootstrapServerHostnameOutputName()),
				InternalBootstrapServerHostname: backend.GetVal(stackOutput, virtualservice.GetInternalBootstrapServerHostnameOutputName()),
				ExternalSchemaRegistryUrl:       backend.GetVal(stackOutput, virtualservice.GetExternalSchemaRegistryUrlOutputName()),
				InternalSchemaRegistryUrl:       backend.GetVal(stackOutput, virtualservice.GetInternalSchemaRegistryUrlOutputName()),
				ExternalKowlDashboardUrl:        backend.GetVal(stackOutput, virtualservice.GetExternalKowlDashboardUrlOutputName()),
				InternalKowlDashboardUrl:        backend.GetVal(stackOutput, virtualservice.GetInternalKowlDashboardUrlOutputName()),
			},
		},
	}
}
