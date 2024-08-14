package pkg

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/util/file"
	certmanagerv1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/certmanager/certmanager/v1"
	istiov1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/istio/networking/v1"
	"github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/strimzioperator/kafka/v1beta2"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	v1 "istio.io/api/networking/v1"
)

func kowl(ctx *pulumi.Context, locals *Locals, kubernetesProvider *kubernetes.Provider,
	createdNamespace *kubernetescorev1.Namespace, createdKafkaCluster *v1beta2.Kafka, labels map[string]string) error {

	type kowlConfigTemplateInput struct {
		BootstrapServerHostname string
		BootstrapServerPort     int
		SaslUsername            string
		SchemaRegistryHostname  string
		RefreshIntervalMinutes  int
	}

	kowlConfig, err := file.RenderTemplate(&kowlConfigTemplateInput{
		BootstrapServerHostname: locals.IngressExternalBootstrapHostname,
		BootstrapServerPort:     vars.ExternalPublicListenerPortNumber,
		SaslUsername:            vars.AdminUsername,
		SchemaRegistryHostname:  locals.SchemaRegistryKubeServiceFqdn,
		RefreshIntervalMinutes:  vars.KowlRefreshIntervalMinutes,
	}, vars.KowlConfigFileTemplate)
	if err != nil {
		return errors.Wrap(err, "failed to render kowl config file")
	}

	createdConfigMap, err := kubernetescorev1.NewConfigMap(ctx,
		vars.KowlConfigMapName,
		&kubernetescorev1.ConfigMapArgs{
			Data: pulumi.ToStringMap(map[string]string{vars.KowlConfigKeyName: string(kowlConfig)}),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(vars.KowlConfigMapName),
				Namespace: createdNamespace.Metadata.Name(),
			},
		}, pulumi.Parent(createdNamespace))
	if err != nil {
		return errors.Wrap(err, "failed to add config-map")
	}

	labels["app"] = vars.KowlDeploymentName

	_, err = appsv1.NewDeployment(ctx, vars.KowlDeploymentName, &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(vars.KowlDeploymentName),
			Namespace: createdNamespace.Metadata.Name(),
			Labels:    pulumi.ToStringMap(labels),
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String(vars.KowlDeploymentName),
				},
			},
			Template: &kubernetescorev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String(vars.KowlDeploymentName),
					},
				},
				Spec: &kubernetescorev1.PodSpecArgs{
					Volumes: kubernetescorev1.VolumeArray{
						kubernetescorev1.VolumeArgs{
							ConfigMap: kubernetescorev1.ConfigMapVolumeSourceArgs{
								Name: createdConfigMap.Metadata.Name(),
							},
							Name: pulumi.String(vars.KowlConfigVolumeName),
						},
					},
					Containers: kubernetescorev1.ContainerArray{
						&kubernetescorev1.ContainerArgs{
							Name:  pulumi.String(vars.KowlDeploymentName),
							Image: pulumi.String(vars.KowlDockerImage),
							Args: pulumi.ToStringArray([]string{
								//https://github.com/cloudhut/charts/blob/master/kowl/templates/deployment.yaml
								fmt.Sprintf("--config.filepath=%s", vars.KowlConfigVolumeMountPath),
								fmt.Sprintf("--kafka.sasl.password=$%s", vars.KowlEnvVarNameKafkaSaslPassword),
							}),
							Ports: kubernetescorev1.ContainerPortArray{
								&kubernetescorev1.ContainerPortArgs{
									Name:          pulumi.String("http"),
									ContainerPort: pulumi.Int(vars.KowlContainerPort),
								},
							},
							Env: kubernetescorev1.EnvVarArray{
								kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
									Name: pulumi.String(vars.KowlEnvVarNameKafkaSaslPassword),
									ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
										SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
											Name: pulumi.String(vars.SaslPasswordSecretName),
											Key:  pulumi.String(vars.SaslPasswordKeyInSecret),
										},
									},
								}),
							},
							VolumeMounts: kubernetescorev1.VolumeMountArray{
								kubernetescorev1.VolumeMountArgs{
									MountPath: pulumi.String(vars.KowlConfigVolumeMountPath),
									Name:      pulumi.String(vars.KowlConfigVolumeName),
									SubPath:   pulumi.String(vars.KowlConfigKeyName),
								},
							},
							Resources: kubernetescorev1.ResourceRequirementsArgs{
								Limits: pulumi.ToStringMap(map[string]string{
									"cpu":    vars.KowlCpuLimits,
									"memory": vars.KowlMemoryLimits,
								}),
								Requests: pulumi.ToStringMap(map[string]string{
									"cpu":    vars.KowlCpuRequests,
									"memory": vars.KowlMemoryRequests,
								}),
							},
						},
					},
				},
			},
		},
	}, pulumi.Parent(createdNamespace), pulumi.DependsOn([]pulumi.Resource{createdKafkaCluster, createdConfigMap}),
		pulumi.IgnoreChanges([]string{"metadata", "status"}))
	if err != nil {
		return errors.Wrap(err, "failed to add kowl deployment")
	}

	//create service
	createdService, err := kubernetescorev1.NewService(ctx,
		vars.KowlDeploymentName,
		&kubernetescorev1.ServiceArgs{
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

	if !locals.KafkaKubernetes.Spec.Ingress.IsEnabled {
		//skip creating ingress for kowl if the ingress is not enabled for kafka itself.
		return nil
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
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "error creating kowl certificate")
	}

	//create gateway
	_, err = istiov1.NewGateway(ctx,
		"kowl-gateway",
		&istiov1.GatewayArgs{
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
		}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{createdService}))
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
		}, pulumi.Parent(createdNamespace),
		pulumi.DependsOn([]pulumi.Resource{createdService}))
	if err != nil {
		return errors.Wrap(err, "error creating schema virtual-service")
	}

	return nil
}
