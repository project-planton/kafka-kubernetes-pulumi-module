package pkg

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8scorev1 "k8s.io/api/core/v1"
)

func schemaRegistryDeployment(ctx *pulumi.Context, locals *Locals, createdNamespace *kubernetescorev1.Namespace,
	labels map[string]string) error {
	labels[englishword.EnglishWord_app.String()] = vars.SchemaRegistryDeploymentName
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
					englishword.EnglishWord_app.String(): pulumi.String(vars.SchemaRegistryDeploymentName),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						englishword.EnglishWord_app.String(): pulumi.String(vars.SchemaRegistryDeploymentName),
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
	}, pulumi.Parent(createdNamespace), pulumi.Timeouts(&pulumi.CustomTimeouts{
		Create: "10s",
		Update: "10s",
		Delete: "10s",
	}))
	if err != nil {
		return errors.Wrap(err, "failed to add deployment")
	}
	return nil
}
