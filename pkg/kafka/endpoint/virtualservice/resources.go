package virtualservice

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	kubernetesdns "github.com/plantoncloud-inc/go-commons/kubernetes/network/dns"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/bootstrap"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/broker"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/endpoint/hostname"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/listener"
	ingressnamespace "github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/ingress/namespace"
	"github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/ingress/gateway/kafka"
	code2cloudv1deploykfcmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
	code2cloudv1deploykfcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/name/output/custom"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "kafka"
)

type Input struct {
	Namespace                        *pulumikubernetescorev1.Namespace
	KafkaHostnames                   []string
	KafkaKubernetesKubernetesStackInput *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackInput
	Labels                           map[string]string
	NamespaceName                    string
	WorkspaceDir                     string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	tlsMatchingRoutes := buildTlsMatchRoutes(
		input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes,
		input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.EnvironmentInfo.EnvironmentName,
		input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.Kubernetes.Ingress.EndpointDomainName,
		input.NamespaceName,
	)
	bootstrapVirtualServiceObject := buildVirtualServiceObject(Name, input.NamespaceName, input.KafkaHostnames, tlsMatchingRoutes)
	manifestPath := filepath.Join(input.WorkspaceDir, fmt.Sprintf("%s.yaml", bootstrapVirtualServiceObject.Name))
	if err := manifest.Create(manifestPath, bootstrapVirtualServiceObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, bootstrapVirtualServiceObject.Name, &pulumik8syaml.ConfigFileArgs{File: manifestPath}, pulumi.Parent(input.Namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add virtual-service manifest")
	}

	exportOutputs(ctx, input.KafkaKubernetesKubernetesStackInput.ResourceInput)

	return nil
}

func exportOutputs(ctx *pulumi.Context, kafkaKubernetesStackResourceInput *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackResourceInput) {
	kafkaKubernetesId := kafkaKubernetesStackResourceInput.KafkaKubernetes.Metadata.Id
	productEnvName := kafkaKubernetesStackResourceInput.KafkaKubernetes.Spec.EnvironmentInfo.EnvironmentName
	endpointDomainName := kafkaKubernetesStackResourceInput.KafkaKubernetes.Spec.Kubernetes.Ingress.EndpointDomainName
	externalBootstrapServerHostname := hostname.GetExternalBootstrapHostname(kafkaKubernetesId, productEnvName, endpointDomainName)
	internalBootstrapServerHostname := hostname.GetInternalBootstrapHostname(kafkaKubernetesId, productEnvName, endpointDomainName)

	externalSchemaRegistryHostname := hostname.GetExternalSchemaRegistryHostname(kafkaKubernetesId, productEnvName, endpointDomainName)
	internalSchemaRegistryHostname := hostname.GetInternalSchemaRegistryHostname(kafkaKubernetesId, productEnvName, endpointDomainName)

	externalKowlDashboardHostname := hostname.GetExternalKowlDashboardHostname(kafkaKubernetesId, productEnvName, endpointDomainName)
	internalKowlDashboardHostname := hostname.GetInteralKowlDashboardHostname(kafkaKubernetesId, productEnvName, endpointDomainName)

	ctx.Export(GetExternalBootstrapServerHostnameOutputName(), pulumi.String(externalBootstrapServerHostname))
	ctx.Export(GetInternalBootstrapServerHostnameOutputName(), pulumi.String(internalBootstrapServerHostname))

	ctx.Export(GetExternalSchemaRegistryUrlOutputName(), pulumi.Sprintf("https://%s", externalSchemaRegistryHostname))
	ctx.Export(GetInternalSchemaRegistryUrlOutputName(), pulumi.Sprintf("https://%s", internalSchemaRegistryHostname))

	ctx.Export(GetExternalKowlDashboardUrlOutputName(), pulumi.Sprintf("https://%s", externalKowlDashboardHostname))
	ctx.Export(GetInternalKowlDashboardUrlOutputName(), pulumi.Sprintf("https://%s", internalKowlDashboardHostname))
}

func GetExternalBootstrapServerHostnameOutputName() string {
	return custom.Name("external-bootstrap-server-hostname")
}

func GetInternalBootstrapServerHostnameOutputName() string {
	return custom.Name("internal-bootstrap-server-hostname")
}

func GetExternalSchemaRegistryUrlOutputName() string {
	return custom.Name("external-schema-registry-url")
}

func GetInternalSchemaRegistryUrlOutputName() string {
	return custom.Name("internal-schema-registry-url")
}

func GetExternalKowlDashboardUrlOutputName() string {
	return custom.Name("external-kowl-dashboard-url")
}

func GetInternalKowlDashboardUrlOutputName() string {
	return custom.Name("internal-kowl-dashboard-url")
}

//buildVirtualServiceObject builds virtual service object
/*
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: kafka
  namespace: tenant-product-kafka-dev-mon
spec:
  gateways:
    - istio-ingress/kafka
  hosts:
    - 'kc-planton-pcs-prod-mon-bootstrap.dev.example.com'
    - 'kc-planton-pcs-prod-mon-broker-b0.dev.example.com'
  tls:
    - match:
        - port: 9092
          sniHosts:
            - "kc-planton-pcs-prod-mon-bootstrap.dev.example.com"
      route:
        - destination:
            host: mon-kafka-extpub-bootstrap.tenant-product-kafka-dev-mon.svc.cluster.local
            port:
              number: 9092
    - match:
        - port: 9092
          sniHosts:
            - "kc-planton-pcs-prod-mon-broker-b0.dev.example.com"
      route:
        - destination:
            host: mon-kafka-extpub-0.tenant-product-kafka-dev-mon.svc.cluster.local
            port:
              number: 9092
*/
func buildVirtualServiceObject(name, namespaceName string, hostnames []string, tlsRoutes []*networkingv1beta1.TLSRoute) *v1beta1.VirtualService {
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
			Gateways: []string{fmt.Sprintf("%s/%s", ingressnamespace.Name, kafka.GatewayName)},
			Hosts:    hostnames,
			Tls:      tlsRoutes,
		},
	}
}

// buildTlsMatchRoutes builds tls matching routes for all the ingress hostnames
func buildTlsMatchRoutes(kafkaKubernetes *code2cloudv1deploykfcmodel.KafkaKubernetes, productEnvName, domainName string, namespaceName string) []*networkingv1beta1.TLSRoute {
	tlsMatchingRoutes := make([]*networkingv1beta1.TLSRoute, 0)

	//external tls routes
	externalBootstrapKubeServiceName := bootstrap.GetKubeServiceName(kafkaKubernetes.Metadata.Id, listener.ExternalPublicListenerName)
	externalBootstrapHostname := hostname.GetExternalBootstrapHostname(kafkaKubernetes.Metadata.Id, productEnvName, domainName)
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
	for i := 0; i < int(kafkaKubernetes.Spec.Kubernetes.KafkaBrokerContainer.Replicas); i++ {
		externalBrokerKubeServiceName := broker.GetKubeServiceName(kafkaKubernetes.Metadata.Id, listener.ExternalPublicListenerName, broker.Id(i))
		externalBrokerHostname := hostname.GetExternalBrokerHostname(kafkaKubernetes.Metadata.Id, productEnvName, domainName, broker.Id(i))
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
	internalBootstrapKubeServiceName := bootstrap.GetKubeServiceName(kafkaKubernetes.Metadata.Id, listener.ExternalPrivateListenerName)
	internalBootstrapHostname := hostname.GetInternalBootstrapHostname(kafkaKubernetes.Metadata.Id, productEnvName, domainName)
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
	for i := 0; i < int(kafkaKubernetes.Spec.Kubernetes.KafkaBrokerContainer.Replicas); i++ {
		internalBrokerKubeServiceName := broker.GetKubeServiceName(kafkaKubernetes.Metadata.Id, listener.ExternalPrivateListenerName, broker.Id(i))
		internalBrokerHostname := hostname.GetInternalBrokerHostname(kafkaKubernetes.Metadata.Id, productEnvName, domainName, broker.Id(i))
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
