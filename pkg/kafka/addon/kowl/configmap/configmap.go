package configmap

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/util/file"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type templateInput struct {
	BootstrapServerHostname string
	BootstrapServerPort     int32
	SaslUsername            string
	SchemaRegistryHostname  string
	RefreshIntervalMinutes  int32
}

const (
	ConfigKeyName          = "kowl.yaml"
	kowlConfigFileTemplate = `
kafka:
  brokers:
    - {{.BootstrapServerHostname}}
  clientId: kowl-on-cluster
  sasl:
    enabled: true
    username: "{{.SaslUsername}}"
    mechanism: SCRAM-SHA-512
  tls:
    enabled: true
  schemaRegistry:
    enabled: true
    urls: ["http://{{.SchemaRegistryHostname}}"]
  protobuf:
    enabled: true
    schemaRegistry:
      enabled: true
      refreshInterval: {{.RefreshIntervalMinutes}}m
`
)

type Input struct {
	Namespace               *pulumikubernetescorev1.Namespace
	Name                    string
	BootstrapServerHostname string
	BootstrapServerPort     int32
	SaslUsername            string
	SchemaRegistryHostname  string
	RefreshIntervalMinutes  int32
}

func Resources(ctx *pulumi.Context, input *Input) (*corev1.ConfigMap, error) {
	kowlConfig, err := file.RenderTemplate(&templateInput{
		BootstrapServerHostname: input.BootstrapServerHostname,
		BootstrapServerPort:     input.BootstrapServerPort,
		SaslUsername:            input.SaslUsername,
		SchemaRegistryHostname:  input.SchemaRegistryHostname,
		RefreshIntervalMinutes:  input.RefreshIntervalMinutes,
	}, kowlConfigFileTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render kowl config file")
	}
	configMap, err := addConfigMap(ctx, input, kowlConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add configmap")
	}
	return configMap, nil
}

func addConfigMap(ctx *pulumi.Context, input *Input, kowlConfig []byte) (*pulumikubernetescorev1.ConfigMap, error) {
	configMap, err := corev1.NewConfigMap(ctx, input.Name, &corev1.ConfigMapArgs{
		Data: pulumi.ToStringMap(map[string]string{ConfigKeyName: string(kowlConfig)}),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(input.Name),
			Namespace: input.Namespace.Metadata.Name(),
		},
	}, pulumi.Parent(input.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "failed to add config-map")
	}
	return configMap, nil
}
