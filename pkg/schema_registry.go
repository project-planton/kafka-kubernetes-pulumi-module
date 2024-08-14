package pkg

import (
	"fmt"
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/certmanager/certmanager/v1"
	istiov1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/istio/networking/v1"
	"github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/strimzioperator/kafka/v1beta2"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	v1 "istio.io/api/networking/v1"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8scorev1 "k8s.io/api/core/v1"
)

func schemaRegistry(ctx *pulumi.Context, locals *Locals, kubernetesProvider *kubernetes.Provider,
	createdNamespace *kubernetescorev1.Namespace, createdKafkaCluster *v1beta2.Kafka, labels map[string]string) error {

	labels["app"] = vars.SchemaRegistryDeploymentName

	//create schema-registry deployment
	_, err := appsv1.NewDeployment(ctx, vars.SchemaRegistryDeploymentName, &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(vars.SchemaRegistryDeploymentName),
			Namespace: createdNamespace.Metadata.Name(),
			Labels:    pulumi.ToStringMap(labels),
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String(vars.SchemaRegistryDeploymentName),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String(vars.SchemaRegistryDeploymentName),
					},
				},
				Spec: &corev1.PodSpecArgs{
					//InitContainers: corev1.ContainerArray{
					//	common.GetKafkaReadyCheckContainer(),
					//},
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String(vars.SchemaRegistryDeploymentName),
							Image: pulumi.String(vars.SchemaRegistryDockerImage),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									Name:          pulumi.String("http"),
									ContainerPort: pulumi.Int(vars.SchemaRegistryContainerPort),
								},
							},
							Resources: corev1.ResourceRequirementsArgs{
								Limits: pulumi.ToStringMap(map[string]string{
									string(k8scorev1.ResourceCPU):    locals.KafkaKubernetes.Spec.SchemaRegistryContainer.Resources.Limits.Cpu,
									string(k8scorev1.ResourceMemory): locals.KafkaKubernetes.Spec.SchemaRegistryContainer.Resources.Limits.Memory,
								}),
								Requests: pulumi.ToStringMap(map[string]string{
									string(k8scorev1.ResourceCPU):    locals.KafkaKubernetes.Spec.SchemaRegistryContainer.Resources.Requests.Cpu,
									string(k8scorev1.ResourceMemory): locals.KafkaKubernetes.Spec.SchemaRegistryContainer.Resources.Requests.Memory,
								}),
							},
							Env: corev1.EnvVarArray{
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name: pulumi.String("SCHEMA_REGISTRY_HOST_NAME"),
									ValueFrom: &corev1.EnvVarSourceArgs{
										FieldRef: &corev1.ObjectFieldSelectorArgs{
											FieldPath: pulumi.String("status.podIP"),
										},
									},
								}),
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name:  pulumi.String("SCHEMA_REGISTRY_LISTENERS"),
									Value: pulumi.String("http://0.0.0.0:8081"),
								}),
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name:  pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_SASL_MECHANISM"),
									Value: pulumi.String("SCRAM-SHA-512"),
								}),
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name:  pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_SECURITY_PROTOCOL"),
									Value: pulumi.String("SASL_SSL"),
								}),
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name:  pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_TOPIC"),
									Value: pulumi.String(vars.SchemaRegistryKafkaStoreTopicName),
								}),
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name:  pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS"),
									Value: pulumi.Sprintf("%s:%d", locals.IngressExternalBootstrapHostname, vars.ExternalPublicListenerPortNumber),
								}),
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name: pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_SASL_JAAS_CONFIG"),
									ValueFrom: &corev1.EnvVarSourceArgs{
										SecretKeyRef: &corev1.SecretKeySelectorArgs{
											Name: pulumi.String(vars.SaslPasswordSecretName),
											Key:  pulumi.String(vars.SaslJaasConfigKeyInSecret),
										},
									},
								}),
							},
						},
					},
				},
			},
		},
	}, pulumi.Parent(createdKafkaCluster))

	//create kubernetes service
	createdService, err := kubernetescorev1.NewService(ctx,
		vars.SchemaRegistryDeploymentName,
		&kubernetescorev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(vars.SchemaRegistryKubeServiceName),
				Namespace: createdNamespace.Metadata.Name(),
			},
			Spec: &kubernetescorev1.ServiceSpecArgs{
				Type: pulumi.String("ClusterIP"),
				Selector: pulumi.StringMap{
					"app": pulumi.String(vars.SchemaRegistryDeploymentName),
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
		}, pulumi.Parent(createdKafkaCluster))
	if err != nil {
		return errors.Wrapf(err, "failed to add schema registry service")
	}

	if !locals.KafkaKubernetes.Spec.Ingress.IsEnabled {
		//skip creating ingress for schema-registry if the ingress is not enabled for kafka itself.
		return nil
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
		}, pulumi.Provider(kubernetesProvider))
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
		}, pulumi.Parent(createdKafkaCluster), pulumi.DependsOn([]pulumi.Resource{createdService}))
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
		}, pulumi.Parent(createdKafkaCluster), pulumi.DependsOn([]pulumi.Resource{createdService}))
	if err != nil {
		return errors.Wrap(err, "error creating schema virtual-service")
	}

	if err != nil {
		return errors.Wrap(err, "failed to add deployment")
	}
	return nil
}
