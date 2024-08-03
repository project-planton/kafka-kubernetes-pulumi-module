package pkg

import (
	"fmt"
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/certmanager/certmanager/v1"
	istiov1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/istio/networking/v1"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	v1 "istio.io/api/networking/v1"
)

func schemaRegistryIstioIngress(ctx *pulumi.Context, locals *Locals, createdNamespace *kubernetescorev1.Namespace, labels map[string]string) error {
	svc, err := kubernetescorev1.NewService(ctx, vars.SchemaRegistryDeploymentName, &kubernetescorev1.ServiceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(vars.SchemaRegistryKubeServiceName),
			Namespace: createdNamespace.Metadata.Name(),
		},
		Spec: &kubernetescorev1.ServiceSpecArgs{
			Type: pulumi.String("ClusterIP"),
			Selector: pulumi.StringMap{
				englishword.EnglishWord_app.String(): pulumi.String(vars.SchemaRegistryDeploymentName),
			},
			Ports: kubernetescorev1.ServicePortArray{
				&kubernetescorev1.ServicePortArgs{
					Name:       pulumi.String("http"),
					Protocol:   pulumi.String("TCP"),
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(vars.SchemaRegistryContainerPort),
				},
			},
		},
	}, pulumi.Parent(createdNamespace))
	if err != nil {
		return errors.Wrapf(err, "failed to add schema registry service")
	}

	//crate new certificate
	addedCertificate, err := certmanagerv1.NewCertificate(ctx,
		"schema-registry-ingress-certificate",
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String(fmt.Sprintf("%s-schema-registry",
					locals.KafkaKubernetes.Metadata.Id)),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames:   pulumi.ToStringArray(locals.IngressSchemaRegistryHostnames),
				SecretName: pulumi.String(locals.IngressSchemaRegistryCertSecretName),
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					Name: pulumi.String(locals.IngressCertClusterIssuerName),
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{svc}), pulumi.Parent(svc))
	if err != nil {
		return errors.Wrap(err, "error creating schema registry certificate")
	}

	//create gateway
	_, err = istiov1.NewGateway(ctx,
		"schema-registry-gateway",
		&istiov1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String(fmt.Sprintf("%s-schema-registry", locals.KafkaKubernetes.Metadata.Id)),
				//all gateway resources should be created in the istio-ingress deployment namespace
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: istiov1.GatewaySpecArgs{
				//the selector labels map should match the desired istio-ingress deployment.
				Selector: pulumi.ToStringMap(vars.IstioIngressSelectorLabels),
				Servers: istiov1.GatewaySpecServersArray{
					&istiov1.GatewaySpecServersArgs{
						Name: pulumi.String("schema-registry-https"),
						Port: &istiov1.GatewaySpecServersPortArgs{
							Number:   pulumi.Int(443),
							Name:     pulumi.String("schema-registry-https"),
							Protocol: pulumi.String("HTTPS"),
						},
						Hosts: pulumi.ToStringArray(locals.IngressSchemaRegistryHostnames),
						Tls: &istiov1.GatewaySpecServersTlsArgs{
							CredentialName: addedCertificate.Spec.SecretName(),
							Mode:           pulumi.String(v1.ServerTLSSettings_SIMPLE.String()),
						},
					},
					&istiov1.GatewaySpecServersArgs{
						Name: pulumi.String("schema-registry-http"),
						Port: &istiov1.GatewaySpecServersPortArgs{
							Number:   pulumi.Int(80),
							Name:     pulumi.String("schema-registry-http"),
							Protocol: pulumi.String("HTTP"),
						},
						Hosts: pulumi.ToStringArray(locals.IngressSchemaRegistryHostnames),
						Tls: &istiov1.GatewaySpecServersTlsArgs{
							HttpsRedirect: pulumi.Bool(true),
						},
					},
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{svc}), pulumi.Parent(svc))
	if err != nil {
		return errors.Wrap(err, "error creating schema registry gateway")
	}

	//create virtual-service
	_, err = istiov1.NewVirtualService(ctx,
		"schema-registry-virtual-service",
		&istiov1.VirtualServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(fmt.Sprintf("%s-schema-registry", locals.KafkaKubernetes.Metadata.Id)),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: istiov1.VirtualServiceSpecArgs{
				Gateways: pulumi.StringArray{
					pulumi.Sprintf("%s/%s-schema-registry",
						vars.IstioIngressNamespace, locals.KafkaKubernetes.Metadata.Id),
				},
				Hosts: pulumi.ToStringArray(locals.IngressSchemaRegistryHostnames),
				Http: istiov1.VirtualServiceSpecHttpArray{
					&istiov1.VirtualServiceSpecHttpArgs{
						Name: pulumi.String(fmt.Sprintf("%s-schema-registry", locals.KafkaKubernetes.Metadata.Id)),
						Route: istiov1.VirtualServiceSpecHttpRouteArray{
							&istiov1.VirtualServiceSpecHttpRouteArgs{
								Destination: istiov1.VirtualServiceSpecHttpRouteDestinationArgs{
									Host: pulumi.String(locals.SchemaRegistryKubeServiceFqdn),
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
