package pkg

import (
	"fmt"
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/certmanager/certmanager/v1"
	istiov1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/istio/networking/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	v1 "istio.io/api/networking/v1"
)

func kowlIstioIngress(ctx *pulumi.Context, locals *Locals, createdNamespace *kubernetescorev1.Namespace, labels map[string]string) error {
	svc, err := kubernetescorev1.NewService(ctx, vars.KowlDeploymentName, &kubernetescorev1.ServiceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(vars.KowlKubeServiceName),
			Namespace: createdNamespace.Metadata.Name(),
		},
		Spec: &kubernetescorev1.ServiceSpecArgs{
			Type: pulumi.String("ClusterIP"),
			Selector: pulumi.StringMap{
				"app": pulumi.String(vars.KowlDeploymentName),
			},
			Ports: kubernetescorev1.ServicePortArray{
				&kubernetescorev1.ServicePortArgs{
					Name:       pulumi.String("http"),
					Protocol:   pulumi.String("TCP"),
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(vars.KowlContainerPort),
				},
			},
		},
	}, pulumi.Parent(createdNamespace))
	if err != nil {
		return errors.Wrapf(err, "failed to add kowl service")
	}

	//crate new certificate
	addedCertificate, err := certmanagerv1.NewCertificate(ctx,
		"kowl-ingress-certificate",
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(fmt.Sprintf("%s-kowl", locals.KafkaKubernetes.Metadata.Id)),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames: pulumi.StringArray{
					pulumi.String(locals.IngressExternalKowlHostname),
				},
				SecretName: pulumi.String(locals.IngressKowlCertSecretName),
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					Name: pulumi.String(locals.IngressCertClusterIssuerName),
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{svc}), pulumi.Parent(svc))
	if err != nil {
		return errors.Wrap(err, "error creating kowl certificate")
	}

	//create gateway
	_, err = istiov1.NewGateway(ctx, "kowl-gateway", &istiov1.GatewayArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("%s-kowl", locals.KafkaKubernetes.Metadata.Id)),
			//all gateway resources should be created in the istio-ingress deployment namespace
			Namespace: pulumi.String(vars.IstioIngressNamespace),
			Labels:    pulumi.ToStringMap(labels),
		},
		Spec: istiov1.GatewaySpecArgs{
			//the selector labels map should match the desired istio-ingress deployment.
			Selector: pulumi.ToStringMap(vars.IstioIngressSelectorLabels),
			Servers: istiov1.GatewaySpecServersArray{
				&istiov1.GatewaySpecServersArgs{
					Name: pulumi.String("kowl-https"),
					Port: &istiov1.GatewaySpecServersPortArgs{
						Number:   pulumi.Int(443),
						Name:     pulumi.String("kowl-https"),
						Protocol: pulumi.String("HTTPS"),
					},
					Hosts: pulumi.StringArray{
						pulumi.String(locals.IngressExternalKowlHostname),
					},
					Tls: &istiov1.GatewaySpecServersTlsArgs{
						CredentialName: addedCertificate.Spec.SecretName(),
						Mode:           pulumi.String(v1.ServerTLSSettings_SIMPLE.String()),
					},
				},
				&istiov1.GatewaySpecServersArgs{
					Name: pulumi.String("kowl-http"),
					Port: &istiov1.GatewaySpecServersPortArgs{
						Number:   pulumi.Int(80),
						Name:     pulumi.String("kowl-http"),
						Protocol: pulumi.String("HTTP"),
					},
					Hosts: pulumi.StringArray{
						pulumi.String(locals.IngressExternalKowlHostname),
					},
					Tls: &istiov1.GatewaySpecServersTlsArgs{
						HttpsRedirect: pulumi.Bool(true),
					},
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{svc}), pulumi.Parent(svc))
	if err != nil {
		return errors.Wrap(err, "error creating kowl gateway")
	}

	//create virtual-service
	_, err = istiov1.NewVirtualService(ctx,
		"kowl-virtual-service",
		&istiov1.VirtualServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(fmt.Sprintf("%s-kowl", locals.KafkaKubernetes.Metadata.Id)),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: istiov1.VirtualServiceSpecArgs{
				Gateways: pulumi.StringArray{
					pulumi.Sprintf("%s/%s-kowl",
						vars.IstioIngressNamespace, locals.KafkaKubernetes.Metadata.Id),
				},
				Hosts: pulumi.StringArray{
					pulumi.String(locals.IngressExternalKowlHostname),
				},
				Http: istiov1.VirtualServiceSpecHttpArray{
					&istiov1.VirtualServiceSpecHttpArgs{
						Name: pulumi.String(fmt.Sprintf("%s-kowl", locals.KafkaKubernetes.Metadata.Id)),
						Route: istiov1.VirtualServiceSpecHttpRouteArray{
							&istiov1.VirtualServiceSpecHttpRouteArgs{
								Destination: istiov1.VirtualServiceSpecHttpRouteDestinationArgs{
									Host: pulumi.String(locals.KowlKubeServiceFqdn),
									Port: istiov1.VirtualServiceSpecHttpRouteDestinationPortArgs{
										Number: pulumi.Int(80),
									},
								},
							},
						},
					},
				},
			},
			Status: nil,
		}, pulumi.DependsOn([]pulumi.Resource{svc}), pulumi.Parent(svc))
	if err != nil {
		return errors.Wrap(err, "error creating schema virtual-service")
	}
	return nil
}
