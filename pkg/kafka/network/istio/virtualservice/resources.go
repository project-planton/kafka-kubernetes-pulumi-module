package virtualservice

import (
	"fmt"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/network/istio/hostname"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	kubernetesdns "github.com/plantoncloud-inc/go-commons/kubernetes/network/dns"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/bootstrap"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/broker"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/listener"
	ingressnamespace "github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/ingress/namespace"
	"github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/ingress/gateway/kafka"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "kafka"
)

func Resources(ctx *pulumi.Context) error {
	i := extractInput(ctx)
	tlsMatchingRoutes := buildTlsMatchRoutes(i.resourceId, i.environmentName, i.endpointDomainName, i.namespaceName, i.brokerReplicas)
	bootstrapVirtualServiceObject := buildVirtualServiceObject(i.namespaceName, i.hostnames, tlsMatchingRoutes)
	manifestPath := filepath.Join(i.workspaceDir, fmt.Sprintf("%s.yaml", bootstrapVirtualServiceObject.Name))
	if err := manifest.Create(manifestPath, bootstrapVirtualServiceObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, bootstrapVirtualServiceObject.Name, &pulumik8syaml.ConfigFileArgs{File: manifestPath}, pulumi.Parent(i.namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add virtual-service manifest")
	}
	return nil
}

func buildVirtualServiceObject(namespaceName string, hostnames []string, tlsRoutes []*networkingv1beta1.TLSRoute) *v1beta1.VirtualService {
	return &v1beta1.VirtualService{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "VirtualService",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      Name,
			Namespace: namespaceName,
		},
		Spec: networkingv1beta1.VirtualService{
			Gateways: []string{fmt.Sprintf("%s/%s", ingressnamespace.Name, kafka.GatewayName)},
			Hosts:    hostnames,
			Tls:      tlsRoutes,
		},
	}
}

// buildTlsMatchRoutes builds tls matching routes for all the ingress hostnames
func buildTlsMatchRoutes(kafkaKubernetesId, productEnvName, domainName, namespaceName string, brokerReplicas int32) []*networkingv1beta1.TLSRoute {
	tlsMatchingRoutes := make([]*networkingv1beta1.TLSRoute, 0)

	//external tls routes
	externalBootstrapKubeServiceName := bootstrap.GetKubeServiceName(kafkaKubernetesId, listener.ExternalPublicListenerName)
	externalBootstrapHostname := hostname.GetExternalBootstrapHostname(kafkaKubernetesId, productEnvName, domainName)
	tlsMatchingRoutes = append(tlsMatchingRoutes, &networkingv1beta1.TLSRoute{
		Match: []*networkingv1beta1.TLSMatchAttributes{{
			Port:     uint32(listener.ExternalPublicListenerPortNumber),
			SniHosts: []string{externalBootstrapHostname},
		}},
		Route: []*networkingv1beta1.RouteDestination{{
			Destination: &networkingv1beta1.Destination{
				Host: fmt.Sprintf("%s.%s.%s", externalBootstrapKubeServiceName, namespaceName, kubernetesdns.DefaultDomain),
				Port: &networkingv1beta1.PortSelector{Number: uint32(listener.ExternalPublicListenerPortNumber)},
			},
		}},
	})
	for i := 0; i < int(brokerReplicas); i++ {
		externalBrokerKubeServiceName := broker.GetKubeServiceName(kafkaKubernetesId, listener.ExternalPublicListenerName, broker.Id(i))
		externalBrokerHostname := hostname.GetExternalBrokerHostname(kafkaKubernetesId, productEnvName, domainName, broker.Id(i))
		tlsMatchingRoutes = append(tlsMatchingRoutes, &networkingv1beta1.TLSRoute{
			Match: []*networkingv1beta1.TLSMatchAttributes{{
				Port:     uint32(listener.ExternalPublicListenerPortNumber),
				SniHosts: []string{externalBrokerHostname},
			}},
			Route: []*networkingv1beta1.RouteDestination{{
				Destination: &networkingv1beta1.Destination{
					Host: fmt.Sprintf("%s.%s.%s", externalBrokerKubeServiceName, namespaceName, kubernetesdns.DefaultDomain),
					Port: &networkingv1beta1.PortSelector{Number: uint32(listener.ExternalPublicListenerPortNumber)},
				},
			}},
		})
	}

	//internal tls matching routes
	internalBootstrapKubeServiceName := bootstrap.GetKubeServiceName(kafkaKubernetesId, listener.ExternalPrivateListenerName)
	internalBootstrapHostname := hostname.GetInternalBootstrapHostname(kafkaKubernetesId, productEnvName, domainName)
	tlsMatchingRoutes = append(tlsMatchingRoutes, &networkingv1beta1.TLSRoute{
		Match: []*networkingv1beta1.TLSMatchAttributes{{
			//private endpoints also listen on same port as public endpoints
			Port:     uint32(listener.ExternalPublicListenerPortNumber),
			SniHosts: []string{internalBootstrapHostname},
		}},
		Route: []*networkingv1beta1.RouteDestination{{
			Destination: &networkingv1beta1.Destination{
				Host: fmt.Sprintf("%s.%s.%s", internalBootstrapKubeServiceName, namespaceName, kubernetesdns.DefaultDomain),
				//requests are forwarded to the internal listener port
				Port: &networkingv1beta1.PortSelector{Number: uint32(listener.ExternalPrivateListenerPortNumber)},
			},
		}},
	})
	for i := 0; i < int(brokerReplicas); i++ {
		internalBrokerKubeServiceName := broker.GetKubeServiceName(kafkaKubernetesId, listener.ExternalPrivateListenerName, broker.Id(i))
		internalBrokerHostname := hostname.GetInternalBrokerHostname(kafkaKubernetesId, productEnvName, domainName, broker.Id(i))
		tlsMatchingRoutes = append(tlsMatchingRoutes, &networkingv1beta1.TLSRoute{
			Match: []*networkingv1beta1.TLSMatchAttributes{{
				//private endpoints also listen on same port as public endpoints
				Port:     uint32(listener.ExternalPublicListenerPortNumber),
				SniHosts: []string{internalBrokerHostname},
			}},
			Route: []*networkingv1beta1.RouteDestination{{
				Destination: &networkingv1beta1.Destination{
					Host: fmt.Sprintf("%s.%s.%s", internalBrokerKubeServiceName, namespaceName, kubernetesdns.DefaultDomain),
					//requests are forwarded to the internal listener port
					Port: &networkingv1beta1.PortSelector{Number: uint32(listener.ExternalPrivateListenerPortNumber)},
				},
			}},
		})
	}
	return tlsMatchingRoutes
}
