package namespace

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/system/istiod"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	puluminamekubeoutput "github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/name/provider/kubernetes/output"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	pulk8scv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	v12 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	KubernetesProvider *pulumikubernetes.Provider
	Labels             map[string]string
	NamespaceName      string
}

func Resources(ctx *pulumi.Context, input *Input) (*pulk8scv1.Namespace, error) {
	namespace, err := addNamespace(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add namespace")
	}
	return namespace, nil
}

func addNamespace(ctx *pulumi.Context, input *Input) (*pulk8scv1.Namespace, error) {
	input.Labels[istiod.SidecarInjectionLabelKey] = istiod.SidecarInjectionLabelValue
	ns, err := pulk8scv1.NewNamespace(ctx, input.NamespaceName, &pulk8scv1.NamespaceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("AddedNamespace"),
		Metadata: v12.ObjectMetaPtrInput(&v12.ObjectMetaArgs{
			Name:   pulumi.String(input.NamespaceName),
			Labels: pulumi.ToStringMap(input.Labels),
		}),
	}, pulumi.Provider(input.KubernetesProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add %s namespace", input.NamespaceName)
	}
	ctx.Export(GetNamespaceNameOutputName(), ns.Metadata.Name())
	return ns, nil
}

func GetNamespaceNameOutputName() string {
	return puluminamekubeoutput.Name(pulk8scv1.Namespace{}, englishword.EnglishWord_namespace.String())
}
