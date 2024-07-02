package endpoint

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/endpoint/cert"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/endpoint/hostname"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/endpoint/virtualservice"
	code2cloudv1deploykfcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/kubernetes/model"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	KafkaKubernetesKubernetesStackInput *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackInput
	Labels                           map[string]string
	NamespaceName                    string
	WorkspaceDir                     string
	Namespace                        *pulumikubernetescorev1.Namespace
}

func Resources(ctx *pulumi.Context, input *Input) error {
	certHostnames := hostname.GetKafkaHostnames(
		input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.EnvironmentInfo.EnvironmentName,
		input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.Kubernetes.Ingress.EndpointDomainName,
		input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes)
	if err := cert.Resources(ctx, &cert.Input{
		Namespace:          input.Namespace,
		Labels:             input.Labels,
		NamespaceName:      input.NamespaceName,
		Hostnames:          certHostnames,
		WorkspaceDir:       input.WorkspaceDir,
		EnvironmentName:    input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.EnvironmentInfo.EnvironmentName,
		EndpointDomainName: input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.Kubernetes.Ingress.EndpointDomainName,
	}); err != nil {
		return errors.Wrap(err, "failed to add cert resources")
	}
	if err := virtualservice.Resources(ctx, &virtualservice.Input{
		Namespace: input.Namespace,
		KafkaHostnames: hostname.GetKafkaHostnames(
			input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.EnvironmentInfo.EnvironmentName,
			input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.Kubernetes.Ingress.EndpointDomainName,
			input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes),
		KafkaKubernetesKubernetesStackInput: input.KafkaKubernetesKubernetesStackInput,
		Labels:                           input.Labels,
		NamespaceName:                    input.NamespaceName,
		WorkspaceDir:                     input.WorkspaceDir,
	}); err != nil {
		return errors.Wrap(err, "failed to add virtual service resources")
	}
	return nil
}
