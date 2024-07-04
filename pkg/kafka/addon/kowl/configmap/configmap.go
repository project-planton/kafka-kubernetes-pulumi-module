package configmap

import (
	"github.com/pkg/errors"
	kowlconfigfiletemplate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/kowl/template"
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) (*pulumi.Context, error) {
	i := extractInput(ctx)
	kowlConfig, err := kowlconfigfiletemplate.Resources(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render kowl config file")
	}
	configMap, err := addConfigMap(ctx, i, kowlConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add configmap")
	}

	var ctxConfig = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	addKowlConfigMapToContext(&ctxConfig, configMap)
	ctx = ctx.WithValue(kafkacontextstate.Key, ctxConfig)
	return ctx, nil
}

func addConfigMap(ctx *pulumi.Context, i *input, kowlConfig []byte) (*pulumikubernetescorev1.ConfigMap, error) {
	configMap, err := corev1.NewConfigMap(ctx, i.name, &corev1.ConfigMapArgs{
		Data: pulumi.ToStringMap(map[string]string{ConfigKeyName: string(kowlConfig)}),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(i.name),
			Namespace: i.namespace.Metadata.Name(),
		},
	}, pulumi.Parent(i.namespace))
	if err != nil {
		return nil, errors.Wrap(err, "failed to add config-map")
	}
	return configMap, nil
}

func addKowlConfigMapToContext(existingConfig *kafkacontextstate.ContextState, kowlConfigMap *pulumikubernetescorev1.ConfigMap) {
	if existingConfig.Status.AddedResources == nil {
		existingConfig.Status.AddedResources = &kafkacontextstate.AddedResources{
			KowlConfigMap: kowlConfigMap,
		}
		return
	}
	existingConfig.Status.AddedResources.KowlConfigMap = kowlConfigMap
}
