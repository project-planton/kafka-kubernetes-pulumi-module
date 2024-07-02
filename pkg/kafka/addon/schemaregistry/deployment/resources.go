package deployment

import (
	"github.com/pkg/errors"
	kubernetesv1model "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/model"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8scorev1 "k8s.io/api/core/v1"
)

const (
	// DockerImage https://hub.docker.com/layers/confluentinc/cp-schema-registry/7.2.6
	DockerImage         = "confluentinc/cp-schema-registry:7.2.6"
	ContainerPort       = 8081
	KafkaStoreTopicName = "schema-registry"
)

type Input struct {
	Labels                       map[string]string
	Namespace                    *pulumikubernetescorev1.Namespace
	SchemaRegistryDeploymentName string
	NamespaceName                string
	BootstrapServerHostname      string
	BootstrapServerPort          int32
	SaslJaasConfigKeyInSecret    string
	SaslJaasConfigSecretName     string
	ContainerResources           *kubernetesv1model.ContainerResources
}

func Resources(ctx *pulumi.Context, input *Input) error {
	if err := addDeployment(ctx, input); err != nil {
		return errors.Wrap(err, "failed to add deployment")
	}
	return nil
}

func addDeployment(ctx *pulumi.Context, input *Input) error {
	labels := input.Labels
	labels[englishword.EnglishWord_app.String()] = input.SchemaRegistryDeploymentName
	_, err := appsv1.NewDeployment(ctx, input.SchemaRegistryDeploymentName, &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(input.SchemaRegistryDeploymentName),
			Namespace: pulumi.String(input.NamespaceName),
			Labels:    pulumi.ToStringMap(labels),
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					englishword.EnglishWord_app.String(): pulumi.String(input.SchemaRegistryDeploymentName),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						englishword.EnglishWord_app.String(): pulumi.String(input.SchemaRegistryDeploymentName),
					},
				},
				Spec: &corev1.PodSpecArgs{
					//InitContainers: corev1.ContainerArray{
					//	common.GetKafkaReadyCheckContainer(),
					//},
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String(input.SchemaRegistryDeploymentName),
							Image: pulumi.String(DockerImage),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									Name:          pulumi.String("http"),
									ContainerPort: pulumi.Int(ContainerPort),
								},
							},
							Resources: corev1.ResourceRequirementsArgs{
								Limits: pulumi.ToStringMap(map[string]string{
									string(k8scorev1.ResourceCPU):    input.ContainerResources.Limits.Cpu,
									string(k8scorev1.ResourceMemory): input.ContainerResources.Limits.Memory,
								}),
								Requests: pulumi.ToStringMap(map[string]string{
									string(k8scorev1.ResourceCPU):    input.ContainerResources.Requests.Cpu,
									string(k8scorev1.ResourceMemory): input.ContainerResources.Requests.Memory,
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
									Value: pulumi.String(KafkaStoreTopicName),
								}),
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name:  pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS"),
									Value: pulumi.Sprintf("%s:%d", input.BootstrapServerHostname, input.BootstrapServerPort),
								}),
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name: pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_SASL_JAAS_CONFIG"),
									ValueFrom: &corev1.EnvVarSourceArgs{
										SecretKeyRef: &corev1.SecretKeySelectorArgs{
											Name: pulumi.String(input.SaslJaasConfigSecretName),
											Key:  pulumi.String(input.SaslJaasConfigKeyInSecret),
										},
									},
								}),
							},
						},
					},
				},
			},
		},
	}, pulumi.Parent(input.Namespace), pulumi.Timeouts(&pulumi.CustomTimeouts{
		Create: "10m",
		Update: "10m",
		Delete: "10m",
	}))
	if err != nil {
		return errors.Wrap(err, "failed to add deployment")
	}
	return nil
}
