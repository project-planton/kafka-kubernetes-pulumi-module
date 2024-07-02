package kafka

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/addon"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/adminuser"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/cluster"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/endpoint"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/endpoint/hostname"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/listener"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/namespace"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/topics"
	kafkak8sstackmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/model"
	pulumikubernetesprovider "github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/automation/provider/kubernetes"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ResourceStack struct {
	WorkspaceDir     string
	Input            *kafkak8sstackmodel.KafkaKubernetesStackInput
	KubernetesLabels map[string]string
}

func (s *ResourceStack) Resources(ctx *pulumi.Context) error {
	kubernetesProvider, err := pulumikubernetesprovider.GetWithStackCredentials(ctx, s.Input.CredentialsInput)
	if err != nil {
		return errors.Wrap(err, "failed to setup kubernetes provider")
	}

	namespaceName := s.Input.ResourceInput.Metadata.Id

	ns, err := namespace.Resources(ctx, getNamespaceInput(namespaceName, s.KubernetesLabels, kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to add namespace resources")
	}
	if err := cluster.Resources(ctx, getClusterInput(s.Input, namespaceName, ns, s.KubernetesLabels, s.WorkspaceDir)); err != nil {
		return errors.Wrap(err, "failed to add kafka kubernetes resources")
	}
	if err := adminuser.Resources(ctx, getAdminUserInput(s.Input, namespaceName, ns, s.KubernetesLabels)); err != nil {
		return errors.Wrap(err, "failed to add admin user resources")
	}
	if err := endpoint.Resources(ctx, getEndpointInput(s.Input, namespaceName, ns, s.KubernetesLabels, s.WorkspaceDir)); err != nil {
		return errors.Wrap(err, "failed to add ingress resources")
	}
	if err := topics.Resources(ctx, &topics.Input{
		KubernetesProvider:                  kubernetesProvider,
		WorkspaceDir:                        s.WorkspaceDir,
		NamespaceName:                       s.Input.ResourceInput.Metadata.Id,
		KafkaKubernetesKubernetesStackInput: s.Input,
		Labels:                              s.KubernetesLabels,
	}); err != nil {
		return errors.Wrap(err, "failed to topics resources")
	}
	if err := addon.Resources(ctx, getAddonsInput(
		s.Input.ResourceInput, namespaceName, s.WorkspaceDir, ns, s.KubernetesLabels)); err != nil {
		return errors.Wrap(err, "failed to add addons resources")
	}
	return nil
}

func getAddonsInput(kafkaKubernetesStackResourceInput *kafkak8sstackmodel.KafkaKubernetesKubernetesStackResourceInput,
	namespaceName, workspace string, ns *v1.Namespace, labels map[string]string) *addon.Input {
	return &addon.Input{
		Labels:        labels,
		Namespace:     ns,
		WorkspaceDir:  workspace,
		NamespaceName: namespaceName,
		KafkaKubernetesKubernetesStackResourceInput: kafkaKubernetesStackResourceInput,
		BootstrapServerHostname: hostname.GetExternalBootstrapHostname(
			kafkaKubernetesStackResourceInput.KafkaKubernetes.Metadata.Id,
			kafkaKubernetesStackResourceInput.KafkaKubernetes.Spec.EnvironmentInfo.EnvironmentName,
			kafkaKubernetesStackResourceInput.KafkaKubernetes.Spec.Kubernetes.Ingress.EndpointDomainName,
		),
		BootstrapServerPort: listener.ExternalPublicListenerPortNumber,
	}
}

func getNamespaceInput(namespaceName string, labels map[string]string,
	kubernetesProvider *pulumikubernetes.Provider) *namespace.Input {
	return &namespace.Input{
		KubernetesProvider: kubernetesProvider,
		NamespaceName:      namespaceName,
		Labels:             labels,
	}
}

func getAdminUserInput(input *kafkak8sstackmodel.KafkaKubernetesKubernetesStackInput, namespaceName string,
	ns *v1.Namespace, labels map[string]string) *adminuser.Input {
	return &adminuser.Input{
		KafkaKubernetesKubernetesStackInput: input,
		NamespaceName:                       namespaceName,
		Namespace:                           ns,
		Labels:                              labels,
	}
}

func getClusterInput(input *kafkak8sstackmodel.KafkaKubernetesKubernetesStackInput, namespaceName string,
	ns *v1.Namespace, labels map[string]string, workspace string) *cluster.Input {
	return &cluster.Input{
		KafkaKubernetesKubernetesStackInput: input,
		Namespace:                           ns,
		NamespaceName:                       namespaceName,
		Labels:                              labels,
		WorkspaceDir:                        workspace,
	}
}

func getEndpointInput(input *kafkak8sstackmodel.KafkaKubernetesKubernetesStackInput, namespaceName string,
	ns *v1.Namespace, labels map[string]string, workspace string) *endpoint.Input {
	return &endpoint.Input{
		KafkaKubernetesKubernetesStackInput: input,
		Labels:                              labels,
		NamespaceName:                       namespaceName,
		WorkspaceDir:                        workspace,
		Namespace:                           ns,
	}
}
