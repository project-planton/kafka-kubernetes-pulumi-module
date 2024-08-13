package pkg

import (
	"github.com/pkg/errors"
	kafkakubernetesmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kafkakubernetes"
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

	//create kubernetes-provider from the credential in the stack-kowlConfigTemplateInput
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesClusterCredential(ctx,
		s.Input.KubernetesClusterCredential, "kubernetes")
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
	createdKafkaCluster, err := kafkaCluster(ctx, locals, createdNamespace, s.Labels)
	if err != nil {
		return errors.Wrap(err, "failed to create kafka-cluster resources")
	}

	//create kafka admin user
	if err := kafkaAdminUser(ctx, locals, createdNamespace, createdKafkaCluster, s.Labels); err != nil {
		return errors.Wrap(err, "failed to create kafka admin user")
	}

	//create kafka topics
	if err := kafkaTopics(ctx, locals, createdNamespace, createdKafkaCluster, s.Labels); err != nil {
		return errors.Wrap(err, "failed to create kafka topics")
	}

	//create kafka istio ingress
	if locals.KafkaKubernetes.Spec.Ingress.IsEnabled {
		if err := kafkaIstioIngress(ctx, locals, kubernetesProvider, createdNamespace, s.Labels); err != nil {
			return errors.Wrap(err, "failed to create kafka istio ingress")
		}
	}

	//create schema-registry
	if locals.KafkaKubernetes.Spec.SchemaRegistryContainer != nil &&
		locals.KafkaKubernetes.Spec.SchemaRegistryContainer.IsEnabled {
		if err := schemaRegistry(ctx, locals, createdNamespace, createdKafkaCluster, s.Labels); err != nil {
			return errors.Wrap(err, "failed to create schema registry deployment")
		}

		if err := schemaRegistryIstioIngress(ctx, locals, createdNamespace, createdKafkaCluster, s.Labels); err != nil {
			return errors.Wrap(err, "failed to create schema registry ingress")
		}
	}

	//create kowl
	if locals.KafkaKubernetes.Spec.IsKowlDashboardEnabled {
		if err := kowl(ctx, locals, createdNamespace, createdKafkaCluster, s.Labels); err != nil {
			return errors.Wrap(err, "failed to create kowl deployment")
		}

		if err := kowlIstioIngress(ctx, locals, kubernetesProvider, createdNamespace, s.Labels); err != nil {
			return errors.Wrap(err, "failed to create kowl ingress")
		}
	}
	return nil
}
