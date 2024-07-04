package ingress

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	"github.com/plantoncloud-inc/go-commons/kubernetes/network/dns"
	"github.com/plantoncloud-inc/go-commons/network/dns/zone"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/schemaregistry/deployment"
	ingressnamespace "github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/ingress/namespace"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Resources(ctx *pulumi.Context) error {
	i := extractInput(ctx)
	svc, err := addService(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to add service")
	}
	externalVirtualServiceObject := buildVirtualServiceObject(fmt.Sprintf("%s-public", KubeServiceName),
		KubeServiceName,
		i.namespaceName,
		i.externalIngressHostname,
		i.environmentName,
		i.kafkaIngressDomain,
	)
	if err := addVirtualService(ctx, externalVirtualServiceObject, svc, i.workspaceDir); err != nil {
		return errors.Wrap(err, "failed to add external virtual service")
	}

	internalVirtualServiceObject := buildVirtualServiceObject(fmt.Sprintf("%s-private", KubeServiceName),
		KubeServiceName,
		i.namespaceName,
		i.internalIngressHostname,
		i.environmentName,
		i.kafkaIngressDomain,
	)
	if err := addVirtualService(ctx, internalVirtualServiceObject, svc, i.workspaceDir); err != nil {
		return errors.Wrap(err, "failed to add internal virtual service")
	}
	return nil
}

func addVirtualService(ctx *pulumi.Context, virtualServiceObject *v1beta1.VirtualService, svc *corev1.Service, workspace string) error {
	resourceName := fmt.Sprintf("virtual-service-%s", virtualServiceObject.Name)
	manifestPath := filepath.Join(workspace, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, virtualServiceObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName, &pulumik8syaml.ConfigFileArgs{
		File: manifestPath,
	}, pulumi.DependsOn([]pulumi.Resource{svc}), pulumi.Parent(svc))
	if err != nil {
		return errors.Wrap(err, "failed to add virtual-service manifest")
	}
	return nil
}

func addService(ctx *pulumi.Context) (*corev1.Service, error) {
	i := extractInput(ctx)
	svc, err := corev1.NewService(ctx, i.schemaRegistryDeploymentName, &corev1.ServiceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(KubeServiceName),
			Namespace: pulumi.String(i.namespaceName),
		},
		Spec: &corev1.ServiceSpecArgs{
			Type: pulumi.String("ClusterIP"),
			Selector: pulumi.StringMap{
				englishword.EnglishWord_app.String(): pulumi.String(i.schemaRegistryDeploymentName),
			},
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Name:       pulumi.String("http"),
					Protocol:   pulumi.String("TCP"),
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(deployment.ContainerPort),
				},
			},
		},
	}, pulumi.Parent(i.namespace))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add service")
	}
	return svc, nil
}

func buildVirtualServiceObject(name, kubeServiceName, namespaceName, ingressHostname, productEnvName, domainName string) *v1beta1.VirtualService {
	gatewayName := getGatewayName(productEnvName, domainName)
	return &v1beta1.VirtualService{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "VirtualService",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      name,
			Namespace: namespaceName,
		},
		Spec: networkingv1beta1.VirtualService{
			Gateways: []string{fmt.Sprintf("%s/%s", ingressnamespace.Name, gatewayName)},
			Hosts:    []string{ingressHostname},
			Http: []*networkingv1beta1.HTTPRoute{{
				Name: kubeServiceName,
				Route: []*networkingv1beta1.HTTPRouteDestination{
					{
						Destination: &networkingv1beta1.Destination{
							Host: fmt.Sprintf("%s.%s.%s", kubeServiceName, namespaceName, dns.DefaultDomain),
							Port: &networkingv1beta1.PortSelector{Number: 80},
						},
					},
				},
			}},
		},
	}
}

func getGatewayName(productEnvName, domainName string) string {
	envDomain := fmt.Sprintf("%s.%s", productEnvName, domainName)
	return fmt.Sprintf(zone.GetZoneName(envDomain))
}

func GetKubeServiceNameFqdn(namespace string) string {
	return fmt.Sprintf("%s.%s.%s", KubeServiceName, namespace, dns.DefaultDomain)
}
