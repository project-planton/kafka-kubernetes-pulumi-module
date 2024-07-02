package kowl

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/addon/kowl/configmap"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/addon/kowl/deployment"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/addon/kowl/ingress"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	Name                   = "kowl"
	RefreshIntervalMinutes = 5
)

type Input struct {
	Labels                  map[string]string
	Namespace               *pulumikubernetescorev1.Namespace
	WorkspaceDir            string
	KafkaKubernetesId          string
	NamespaceName           string
	BootstrapServerHostname string
	BootstrapServerPort     int32
	SaslUsername            string
	SchemaRegistryHostname  string
	KafkaIngressDomain      string
	EnvironmentName         string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	configMap, err := configmap.Resources(ctx, getConfigMapInput(input))
	if err != nil {
		return err
	}
	if err := deployment.Resources(ctx, getDeploymentInput(input, configMap)); err != nil {
		return errors.Wrap(err, "failed to add deployment resources")
	}
	if err := ingress.Resources(ctx, &ingress.Input{
		Namespace:          input.Namespace,
		KafkaKubernetesId:     input.KafkaKubernetesId,
		KowlDeploymentName: Name,
		NamespaceName:      input.NamespaceName,
		WorkspaceDir:       input.WorkspaceDir,
		KafkaIngressDomain: input.KafkaIngressDomain,
		EnvironmentName:    input.EnvironmentName,
	}); err != nil {
		return errors.Wrap(err, "failed to add ingress resources")
	}
	return nil
}

func getDeploymentInput(kowlInput *Input, configMap *pulumikubernetescorev1.ConfigMap) *deployment.Input {
	return &deployment.Input{
		Labels:             kowlInput.Labels,
		KowlDeploymentName: Name,
		NamespaceName:      kowlInput.NamespaceName,
		KowlConfigMap:      configMap,
	}
}

func getConfigMapInput(kowlInput *Input) *configmap.Input {
	return &configmap.Input{
		Namespace:               kowlInput.Namespace,
		Name:                    Name,
		BootstrapServerHostname: kowlInput.BootstrapServerHostname,
		BootstrapServerPort:     kowlInput.BootstrapServerPort,
		SaslUsername:            kowlInput.SaslUsername,
		SchemaRegistryHostname:  kowlInput.SchemaRegistryHostname,
		RefreshIntervalMinutes:  RefreshIntervalMinutes,
	}
}
