package pkg

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/listener"
	certmanagerv1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/certmanager/certmanager/v1"
	istiov1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/istio/networking/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	v1 "istio.io/api/networking/v1"
)

func istioIngress(ctx *pulumi.Context, locals *Locals, createdNamespace *kubernetescorev1.Namespace, labels map[string]string) error {
	//crate new certificate
	addedCertificate, err := certmanagerv1.NewCertificate(ctx,
		"ingress-certificate",
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.KafkaKubernetes.Metadata.Id),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames:   pulumi.ToStringArray(locals.IngressHostnames),
				SecretName: pulumi.String(locals.IngressCertSecretName),
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					Name: pulumi.String(locals.IngressCertClusterIssuerName),
				},
			},
		})
	if err != nil {
		return errors.Wrap(err, "error creating certificate")
	}

	//create gateway
	_, err = istiov1.NewGateway(ctx, fmt.Sprintf("gateway-%s", locals.KafkaKubernetes.Metadata.Id), &istiov1.GatewayArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name: pulumi.String(locals.KafkaKubernetes.Metadata.Id),
			//all gateway resources should be created in the istio-ingress deployment namespace
			Namespace: pulumi.String(vars.IstioIngressNamespace),
			Labels:    pulumi.ToStringMap(labels),
		},
		Spec: istiov1.GatewaySpecArgs{
			//the selector labels map should match the desired istio-ingress deployment.
			Selector: pulumi.ToStringMap(vars.IstioIngressSelectorLabels),
			Servers: istiov1.GatewaySpecServersArray{
				&istiov1.GatewaySpecServersArgs{
					Port: &istiov1.GatewaySpecServersPortArgs{
						Number:   pulumi.Int(vars.ExternalPublicListenerPortNumber),
						Name:     pulumi.String("tcp-kafka"),
						Protocol: pulumi.String("TLS"),
					},
					Hosts: pulumi.ToStringArray(locals.IngressHostnames),
					Tls: &istiov1.GatewaySpecServersTlsArgs{
						CredentialName: addedCertificate.Spec.SecretName(),
						Mode:           pulumi.String(v1.ServerTLSSettings_PASSTHROUGH.String()),
					},
				},
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "error creating gateway")
	}

	// create external hostnames array
	var externalHostnames = pulumi.StringArray{
		pulumi.String(locals.IngressExternalBootstrapHostname),
	}

	for _, externalBrokerHostname := range locals.IngressExternalBrokerHostnames {
		externalHostnames = append(externalHostnames, pulumi.String(externalBrokerHostname))
	}

	// create internal hostnames array
	var internalHostnames = pulumi.StringArray{
		pulumi.String(locals.IngressInternalBootstrapHostname),
	}

	for _, internalBrokerHostname := range locals.IngressInternalBrokerHostnames {
		internalHostnames = append(internalHostnames, pulumi.String(internalBrokerHostname))
	}

	// create tls matching routes
	tlsMatchingRoutes := istiov1.VirtualServiceSpecTlsArray{
		//external tls routes
		istiov1.VirtualServiceSpecTlsArgs{
			Match: istiov1.VirtualServiceSpecTlsMatchArray{
				istiov1.VirtualServiceSpecTlsMatchArgs{
					Port:     pulumi.Int(listener.ExternalPublicListenerPortNumber),
					SniHosts: externalHostnames,
				},
			},
			Route: istiov1.VirtualServiceSpecTlsRouteArray{
				istiov1.VirtualServiceSpecTlsRouteArgs{
					Destination: istiov1.VirtualServiceSpecTlsRouteDestinationArgs{
						Host: pulumi.String(locals.BootstrapKubeServiceFqdn),
						Port: istiov1.VirtualServiceSpecTlsRouteDestinationPortArgs{
							Number: pulumi.Int(listener.ExternalPublicListenerPortNumber),
						},
					},
				},
			},
		},

		//internal tls routes
		istiov1.VirtualServiceSpecTlsArgs{
			Match: istiov1.VirtualServiceSpecTlsMatchArray{
				istiov1.VirtualServiceSpecTlsMatchArgs{
					//private endpoints also listen on same port as public endpoints
					Port:     pulumi.Int(listener.ExternalPublicListenerPortNumber),
					SniHosts: internalHostnames,
				},
			},
			Route: istiov1.VirtualServiceSpecTlsRouteArray{
				istiov1.VirtualServiceSpecTlsRouteArgs{
					Destination: istiov1.VirtualServiceSpecTlsRouteDestinationArgs{
						Host: pulumi.String(locals.BootstrapKubeServiceFqdn),
						Port: istiov1.VirtualServiceSpecTlsRouteDestinationPortArgs{
							//requests are forwarded to the internal listener port
							Number: pulumi.Int(listener.ExternalPrivateListenerPortNumber),
						},
					},
				},
			},
		},
	}

	//create virtual-service
	_, err = istiov1.NewVirtualService(ctx,
		locals.KafkaKubernetes.Metadata.Id,
		&istiov1.VirtualServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.KafkaKubernetes.Metadata.Id),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: istiov1.VirtualServiceSpecArgs{
				Gateways: pulumi.StringArray{
					pulumi.Sprintf("%s/%s", vars.IstioIngressNamespace,
						locals.KafkaKubernetes.Metadata.Id),
				},
				Hosts: pulumi.ToStringArray(locals.IngressHostnames),
				Tls:   tlsMatchingRoutes,
			},
		})
	if err != nil {
		return errors.Wrap(err, "error creating virtual-service")
	}
	return nil
}
