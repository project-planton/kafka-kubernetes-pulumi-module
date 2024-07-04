package kafka

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/adminuser"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/cluster"
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	kafkanamespace "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/namespace"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/network/istio"
	kafkaoutputs "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/outputs"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/topics"
	kafkakubernetesstackmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/model"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ResourceStack struct {
	WorkspaceDir     string
	Input            *kafkakubernetesstackmodel.KafkaKubernetesStackInput
	KubernetesLabels map[string]string
}

func (resourceStack *ResourceStack) Resources(ctx *pulumi.Context) error {
	//load context state
	var contextState, err = loadConfig(ctx, resourceStack)
	if err != nil {
		return errors.Wrap(err, "failed to initiate context state")
	}
	ctx = ctx.WithValue(kafkacontextstate.Key, *contextState)

	// Create the namespace resource
	ctx, err = kafkanamespace.Resources(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace resource")
	}

	if err := cluster.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add kafka kubernetes resources")
	}
	if err := adminuser.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add admin user resources")
	}
	if err := istio.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add ingress resources")
	}
	if err := topics.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to topics resources")
	}

	if err := addon.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add addons resources")
	}

	err = kafkaoutputs.Export(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to export kafka kubernetes outputs")
	}

	return nil
}
