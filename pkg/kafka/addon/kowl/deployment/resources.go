package deployment

import (
	"fmt"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"

	"github.com/pkg/errors"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/addon/kowl/configmap"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/adminuser"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	DockerImage                 = "quay.io/cloudhut/kowl:master-59f68da"
	ContainerPort               = 8080
	EnvVarNameKafkaSaslPassword = "KAFKA_SASL_PASSWORD"
	KowlConfigVolumeName        = "kowl-config"
	KowlConfigVolumeMountPath   = "/var/kowl/config.yaml"
	CpuRequests                 = "25m"
	CpuLimits                   = "150m"
	MemoryRequests              = "90Mi"
	MemoryLimits                = "180Mi"
)

func Resources(ctx *pulumi.Context) error {
	if err := addDeployment(ctx); err != nil {
		return errors.Wrap(err, "failed to add deployment")
	}
	return nil
}

func addDeployment(ctx *pulumi.Context) error {
	i := extractInput(ctx)
	labels := i.labels
	labels[englishword.EnglishWord_app.String()] = i.kowlDeploymentName
	_, err := appsv1.NewDeployment(ctx, i.kowlDeploymentName, &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(i.kowlDeploymentName),
			Namespace: pulumi.String(i.namespaceName),
			Labels:    pulumi.ToStringMap(labels),
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					englishword.EnglishWord_app.String(): pulumi.String(i.kowlDeploymentName),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						englishword.EnglishWord_app.String(): pulumi.String(i.kowlDeploymentName),
					},
				},
				Spec: &corev1.PodSpecArgs{
					//InitContainers: corev1.ContainerArray{
					//	common.GetKafkaReadyCheckContainer(),
					//},
					Volumes: corev1.VolumeArray{
						corev1.VolumeArgs{
							ConfigMap: corev1.ConfigMapVolumeSourceArgs{
								Name: i.kowlConfigMap.Metadata.Name(),
							},
							Name: pulumi.String(KowlConfigVolumeName),
						},
					},
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String(i.kowlDeploymentName),
							Image: pulumi.String(DockerImage),
							Args: pulumi.ToStringArray([]string{
								//https://github.com/cloudhut/charts/blob/master/kowl/templates/deployment.yaml
								fmt.Sprintf("--config.filepath=%s", KowlConfigVolumeMountPath),
								fmt.Sprintf("--kafka.sasl.password=$%s", EnvVarNameKafkaSaslPassword),
							}),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									Name:          pulumi.String("http"),
									ContainerPort: pulumi.Int(ContainerPort),
								},
							},
							Env: corev1.EnvVarArray{
								corev1.EnvVarInput(corev1.EnvVarArgs{
									Name: pulumi.String(EnvVarNameKafkaSaslPassword),
									ValueFrom: &corev1.EnvVarSourceArgs{
										SecretKeyRef: &corev1.SecretKeySelectorArgs{
											Name: pulumi.String(adminuser.SaslPasswordSecretName),
											Key:  pulumi.String(adminuser.SaslPasswordKeyInSecret),
										},
									},
								}),
							},
							VolumeMounts: corev1.VolumeMountArray{
								corev1.VolumeMountArgs{
									MountPath: pulumi.String(KowlConfigVolumeMountPath),
									Name:      pulumi.String(KowlConfigVolumeName),
									SubPath:   pulumi.String(configmap.ConfigKeyName),
								},
							},
							Resources: corev1.ResourceRequirementsArgs{
								Limits: pulumi.ToStringMap(map[string]string{
									englishword.EnglishWord_cpu.String():    CpuLimits,
									englishword.EnglishWord_memory.String(): MemoryLimits,
								}),
								Requests: pulumi.ToStringMap(map[string]string{
									englishword.EnglishWord_cpu.String():    CpuRequests,
									englishword.EnglishWord_memory.String(): MemoryRequests,
								}),
							},
						},
					},
				},
			},
		},
	}, pulumi.IgnoreChanges([]string{"metadata", "status"}), pulumi.Parent(i.kowlConfigMap))
	if err != nil {
		return errors.Wrap(err, "failed to add deployment")
	}
	return nil
}
