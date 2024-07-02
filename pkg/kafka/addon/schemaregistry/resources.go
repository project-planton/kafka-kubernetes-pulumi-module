package schemaregistry

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/addon/schemaregistry/deployment"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/addon/schemaregistry/ingress"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/adminuser"
	kubernetesv1model "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/model"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	Name = "schema-registry"
)

type Input struct {
	Labels                  map[string]string
	Namespace               *pulumikubernetescorev1.Namespace
	KafkaKubernetesId          string
	WorkspaceDir            string
	NamespaceName           string
	BootstrapServerHostname string
	BootstrapServerPort     int32
	EnvironmentName         string
	KafkaIngressDomain      string
	ContainerResources      *kubernetesv1model.ContainerResources
}

func Resources(ctx *pulumi.Context, input *Input) error {
	if err := deployment.Resources(ctx, &deployment.Input{
		Labels:                       input.Labels,
		Namespace:                    input.Namespace,
		SchemaRegistryDeploymentName: Name,
		NamespaceName:                input.NamespaceName,
		BootstrapServerHostname:      input.BootstrapServerHostname,
		BootstrapServerPort:          input.BootstrapServerPort,
		SaslJaasConfigSecretName:     adminuser.SaslPasswordSecretName,
		SaslJaasConfigKeyInSecret:    adminuser.SaslJaasConfigKeyInSecret,
		ContainerResources:           input.ContainerResources,
	}); err != nil {
		return errors.Wrap(err, "failed to add deployment resources")
	}
	if err := ingress.Resources(ctx, &ingress.Input{
		Namespace:                    input.Namespace,
		SchemaRegistryDeploymentName: Name,
		KafkaKubernetesId:               input.KafkaKubernetesId,
		NamespaceName:                input.NamespaceName,
		WorkspaceDir:                 input.WorkspaceDir,
		KafkaIngressDomain:           input.KafkaIngressDomain,
		EnvironmentName:              input.EnvironmentName,
	}); err != nil {
		return errors.Wrap(err, "failed to add ingress resources")
	}
	return nil
}
