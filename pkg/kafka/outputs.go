package kafka

import (
	"context"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/outputs"

	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/iac/v1/stackjob/enums/stackjoboperationtype"

	"github.com/pkg/errors"
	kafkastatemodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
	kafkastackmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/model"
	"github.com/plantoncloud/stack-job-runner-golang-sdk/pkg/stack/output/backend"
)

func Outputs(ctx context.Context, input *kafkastackmodel.KafkaKubernetesStackInput) (*kafkastatemodel.KafkaKubernetesStatusStackOutputs, error) {
	stackOutput, err := backend.StackOutput(input.StackJob)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stack output")
	}
	return OutputMapTransformer(stackOutput, input), nil
}

func OutputMapTransformer(stackOutput map[string]interface{}, input *kafkastackmodel.KafkaKubernetesStackInput) *kafkastatemodel.KafkaKubernetesStatusStackOutputs {
	if input.StackJob.Spec.OperationType != stackjoboperationtype.StackJobOperationType_apply || stackOutput == nil {
		return &kafkastatemodel.KafkaKubernetesStatusStackOutputs{}
	}
	return &kafkastatemodel.KafkaKubernetesStatusStackOutputs{
		Namespace:                       backend.GetVal(stackOutput, outputs.GetNamespaceNameOutputName()),
		KafkaSaslUsername:               backend.GetVal(stackOutput, outputs.GetSaslUsernameOutputName()),
		ExternalBootstrapServerHostname: backend.GetVal(stackOutput, outputs.GetExternalBootstrapServerHostnameOutputName()),
		InternalBootstrapServerHostname: backend.GetVal(stackOutput, outputs.GetInternalBootstrapServerHostnameOutputName()),
		ExternalSchemaRegistryUrl:       backend.GetVal(stackOutput, outputs.GetExternalSchemaRegistryUrlOutputName()),
		InternalSchemaRegistryUrl:       backend.GetVal(stackOutput, outputs.GetInternalSchemaRegistryUrlOutputName()),
		ExternalKowlDashboardUrl:        backend.GetVal(stackOutput, outputs.GetExternalKowlDashboardUrlOutputName()),
		InternalKowlDashboardUrl:        backend.GetVal(stackOutput, outputs.GetInternalKowlDashboardUrlOutputName()),
	}
}
