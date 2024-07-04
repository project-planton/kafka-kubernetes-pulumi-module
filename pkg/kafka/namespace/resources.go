package namespace

import (
	"github.com/pkg/errors"
	kafkacontextstate "github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/contextstate"
	"github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/system/istiod"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	"github.com/plantoncloud/pulumi-blueprint-golang-commons/pkg/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	v12 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) (*pulumi.Context, error) {
	namespace, err := addNamespace(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add namespace")
	}

	var ctxConfig = ctx.Value(kafkacontextstate.Key).(kafkacontextstate.ContextState)

	addNamespaceToContext(&ctxConfig, namespace)
	ctx = ctx.WithValue(kafkacontextstate.Key, ctxConfig)
	return ctx, nil
}

func addNamespace(ctx *pulumi.Context) (*kubernetescorev1.Namespace, error) {
	i := extractInput(ctx)
	i.labels[istiod.SidecarInjectionLabelKey] = istiod.SidecarInjectionLabelValue
	ns, err := kubernetescorev1.NewNamespace(ctx, i.namespaceName, &kubernetescorev1.NamespaceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Namespace"),
		Metadata: v12.ObjectMetaPtrInput(&v12.ObjectMetaArgs{
			Name:   pulumi.String(i.namespaceName),
			Labels: pulumi.ToStringMap(i.labels),
		}),
	}, pulumi.Provider(i.kubeProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add %s namespace", i.namespaceName)
	}
	return ns, nil
}

func addNamespaceToContext(existingConfig *kafkacontextstate.ContextState, namespace *kubernetescorev1.Namespace) {
	if existingConfig.Status.AddedResources == nil {
		existingConfig.Status.AddedResources = &kafkacontextstate.AddedResources{
			Namespace: namespace,
		}
		return
	}
	existingConfig.Status.AddedResources.Namespace = namespace
}

func GetNamespaceNameOutputName() string {
	return pulumikubernetesprovider.PulumiOutputName(kubernetescorev1.Namespace{}, englishword.EnglishWord_namespace.String())
}
