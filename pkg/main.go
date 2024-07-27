package pkg

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon"
	kafkaoutputs "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/outputs"
	kafkakubernetesmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kafkakubernetes/model"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ResourceStack struct {
	Input  *kafkakubernetesmodel.KafkaKubernetesStackInput
	Labels map[string]string
}

func (s *ResourceStack) Resources(ctx *pulumi.Context) error {
	locals := initializeLocals(ctx, s.Input)

	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesClusterCredential(ctx,
		s.Input.KubernetesClusterCredential)
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	//create namespace resource
	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: kubernetesmetav1.ObjectMetaPtrInput(
				&kubernetesmetav1.ObjectMetaArgs{
					Name:   pulumi.String(locals.Namespace),
					Labels: pulumi.ToStringMap(s.Labels),
				}),
		},
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "5s", Update: "5s", Delete: "5s"}),
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	//create kafka cluster custom resource
	if err := kafkaCluster(ctx, locals, createdNamespace, s.Labels); err != nil {
		return errors.Wrap(err, "failed to create kafka-cluster resources")
	}

	//create kafka admin user
	if err := adminUser(ctx, locals, createdNamespace, s.Labels); err != nil {
		return errors.Wrap(err, "failed to create kafka admin user")
	}

	//create kafka istio ingress
	if err := istioIngress(ctx, locals, createdNamespace, s.Labels); err != nil {
		return errors.Wrap(err, "failed to create kafka istio ingress")
	}

	//create kafka topics
	if err := kafkaTopics(ctx, locals, createdNamespace, s.Labels); err != nil {
		return errors.Wrap(err, "failed to create kafka topics")
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
