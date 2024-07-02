package addon

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/addon/kowl"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/addon/schemaregistry"
	schemaregistryingress "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/addon/schemaregistry/ingress"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/adminuser"
	code2cloudv1deploykfcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/kubernetes/model"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	Labels                                   map[string]string
	Namespace                                *pulumikubernetescorev1.Namespace
	WorkspaceDir                             string
	NamespaceName                            string
	KafkaKubernetesKubernetesStackResourceInput *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackResourceInput
	BootstrapServerHostname                  string
	BootstrapServerPort                      int32
}

func Resources(ctx *pulumi.Context, input *Input) error {
	if input.KafkaKubernetesKubernetesStackResourceInput.KafkaKubernetes.Spec.Kubernetes.SchemaRegistryContainer.IsEnabled {
		if err := schemaregistry.Resources(ctx, &schemaregistry.Input{
			Labels:                  input.Labels,
			Namespace:               input.Namespace,
			KafkaKubernetesId:          input.KafkaKubernetesKubernetesStackResourceInput.KafkaKubernetes.Metadata.Id,
			WorkspaceDir:            input.WorkspaceDir,
			NamespaceName:           input.NamespaceName,
			BootstrapServerHostname: input.BootstrapServerHostname,
			BootstrapServerPort:     input.BootstrapServerPort,
			EnvironmentName:         input.KafkaKubernetesKubernetesStackResourceInput.KafkaKubernetes.Spec.EnvironmentInfo.EnvironmentName,
			KafkaIngressDomain:      input.KafkaKubernetesKubernetesStackResourceInput.KafkaKubernetes.Spec.Kubernetes.Ingress.EndpointDomainName,
			ContainerResources:      input.KafkaKubernetesKubernetesStackResourceInput.KafkaKubernetes.Spec.Kubernetes.SchemaRegistryContainer.Resources,
		}); err != nil {
			return errors.Wrap(err, "failed to add schema-registry resources")
		}
	}

	if input.KafkaKubernetesKubernetesStackResourceInput.KafkaKubernetes.Spec.Kubernetes.IsKowlDashboardEnabled {
		if err := kowl.Resources(ctx, &kowl.Input{
			Labels:                  input.Labels,
			Namespace:               input.Namespace,
			WorkspaceDir:            input.WorkspaceDir,
			KafkaKubernetesId:          input.KafkaKubernetesKubernetesStackResourceInput.KafkaKubernetes.Metadata.Id,
			NamespaceName:           input.NamespaceName,
			BootstrapServerHostname: input.BootstrapServerHostname,
			BootstrapServerPort:     input.BootstrapServerPort,
			SaslUsername:            adminuser.Username,
			SchemaRegistryHostname:  schemaregistryingress.GetKubeServiceNameFqdn(input.NamespaceName),
			KafkaIngressDomain:      input.KafkaKubernetesKubernetesStackResourceInput.KafkaKubernetes.Spec.Kubernetes.Ingress.EndpointDomainName,
			EnvironmentName:         input.KafkaKubernetesKubernetesStackResourceInput.KafkaKubernetes.Spec.EnvironmentInfo.EnvironmentName,
		}); err != nil {
			return errors.Wrap(err, "failed to add kowl resources")
		}
	}
	return nil
}
