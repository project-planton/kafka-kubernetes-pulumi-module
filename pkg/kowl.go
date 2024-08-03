package pkg

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/util/file"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kowl(ctx *pulumi.Context, locals *Locals, createdNamespace *kubernetescorev1.Namespace,
	labels map[string]string) error {

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

	configMap, err := kubernetescorev1.NewConfigMap(ctx, vars.KowlConfigMapName, &kubernetescorev1.ConfigMapArgs{
		Data: pulumi.ToStringMap(map[string]string{vars.KowlConfigKeyName: string(kowlConfig)}),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(vars.KowlConfigMapName),
			Namespace: createdNamespace.Metadata.Name(),
		},
	}, pulumi.Parent(createdNamespace))
	if err != nil {
		return errors.Wrap(err, "failed to add config-map")
	}

	labels[englishword.EnglishWord_app.String()] = vars.KowlDeploymentName
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
					englishword.EnglishWord_app.String(): pulumi.String(vars.KowlDeploymentName),
				},
			},
			Template: &kubernetescorev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						englishword.EnglishWord_app.String(): pulumi.String(vars.KowlDeploymentName),
					},
				},
				Spec: &kubernetescorev1.PodSpecArgs{
					Volumes: kubernetescorev1.VolumeArray{
						kubernetescorev1.VolumeArgs{
							ConfigMap: kubernetescorev1.ConfigMapVolumeSourceArgs{
								Name: configMap.Metadata.Name(),
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
									englishword.EnglishWord_cpu.String():    vars.KowlCpuLimits,
									englishword.EnglishWord_memory.String(): vars.KowlMemoryLimits,
								}),
								Requests: pulumi.ToStringMap(map[string]string{
									englishword.EnglishWord_cpu.String():    vars.KowlCpuRequests,
									englishword.EnglishWord_memory.String(): vars.KowlMemoryRequests,
								}),
							},
						},
					},
				},
			},
		},
	}, pulumi.IgnoreChanges([]string{"metadata", "status"}), pulumi.Parent(configMap))
	if err != nil {
		return errors.Wrap(err, "failed to add kowl deployment")
	}
	return nil
}
