package pkg

import (
	strimzitypes "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/strimzioperator/kafka/v1beta2"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kafkaCluster(ctx *pulumi.Context, locals *Locals, createdNamespace *kubernetescorev1.Namespace,
	labels map[string]string) error {

	listenersInput := v1beta2.KafkaSpecKafkaListenersArray{}
	listenersInput = append(listenersInput, v1beta2.KafkaSpecKafkaListenersArgs{
		Name: pulumi.String(vars.InternalListenerName),
		Port: pulumi.Int(vars.InternalListenerPortNumber),
		Authentication: &v1beta2.KafkaSpecKafkaListenersAuthenticationArgs{
			Type: pulumi.String(strimzitypes.KafkaSpecKafkaListenersElemAuthenticationTypeScramSha512),
		},
		Tls:  pulumi.Bool(false),
		Type: pulumi.String(strimzitypes.KafkaSpecKafkaListenersElemTypeInternal),
	})

	//internal listeners
	listenersInput = append(listenersInput, v1beta2.KafkaSpecKafkaListenersArgs{
		Name: pulumi.String(vars.ExternalPrivateListenerName),
		Port: pulumi.Int(vars.ExternalPrivateListenerPortNumber),
		Tls:  pulumi.Bool(true),
		Type: pulumi.String(strimzitypes.KafkaSpecKafkaListenersElemTypeIngress),
		Authentication: &v1beta2.KafkaSpecKafkaListenersAuthenticationArgs{
			Type: pulumi.String(strimzitypes.KafkaSpecKafkaListenersElemAuthenticationTypeScramSha512),
		},
		Configuration: &v1beta2.KafkaSpecKafkaListenersConfigurationArgs{
			Bootstrap: &v1beta2.KafkaSpecKafkaListenersConfigurationBootstrapArgs{
				Host: pulumi.String(locals.IngressInternalBootstrapHostname),
			},
			//private listeners should also advertise public listener port as the port translation is done by ingress controller
			Brokers: getBrokerListenersConfig(locals.IngressInternalBrokerHostnames),
			BrokerCertChainAndKey: &v1beta2.KafkaSpecKafkaListenersConfigurationBrokerCertChainAndKeyArgs{
				Certificate: pulumi.String("tls.crt"),
				Key:         pulumi.String("tls.key"),
				SecretName:  pulumi.String(vars.CertSecretName),
			},
		},
	})

	//external listeners
	listenersInput = append(listenersInput, v1beta2.KafkaSpecKafkaListenersArgs{
		Name: pulumi.String(vars.ExternalPublicListenerName),
		Port: pulumi.Int(vars.ExternalPublicListenerPortNumber),
		Tls:  pulumi.Bool(true),
		Type: pulumi.String(strimzitypes.KafkaSpecKafkaListenersElemTypeIngress),
		Authentication: &v1beta2.KafkaSpecKafkaListenersAuthenticationArgs{
			Type: pulumi.String(strimzitypes.KafkaSpecKafkaListenersElemAuthenticationTypeScramSha512),
		},
		Configuration: &v1beta2.KafkaSpecKafkaListenersConfigurationArgs{
			Bootstrap: &v1beta2.KafkaSpecKafkaListenersConfigurationBootstrapArgs{
				Host: pulumi.String(locals.IngressExternalBootstrapHostname),
			},
			Brokers: getBrokerListenersConfig(locals.IngressExternalBrokerHostnames),
			BrokerCertChainAndKey: &v1beta2.KafkaSpecKafkaListenersConfigurationBrokerCertChainAndKeyArgs{
				Certificate: pulumi.String("tls.crt"),
				Key:         pulumi.String("tls.key"),
				SecretName:  pulumi.String(vars.CertSecretName),
			},
		},
	})

	// create kafka cluster
	_, err := v1beta2.NewKafka(ctx, "kafka-cluster", &v1beta2.KafkaArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.KafkaKubernetes.Metadata.Id),
			Namespace: createdNamespace.Metadata.Name(),
			Labels:    pulumi.ToStringMap(labels),
		},
		Spec: v1beta2.KafkaSpecArgs{
			EntityOperator: v1beta2.KafkaSpecEntityOperatorArgs{
				//todo: understand resource usage pattern and allocate
				TopicOperator: v1beta2.KafkaSpecEntityOperatorTopicOperatorArgs{},
				UserOperator:  v1beta2.KafkaSpecEntityOperatorUserOperatorArgs{},
			},
			Kafka: v1beta2.KafkaSpecKafkaArgs{
				Authorization: v1beta2.KafkaSpecKafkaAuthorizationArgs{
					SuperUsers: pulumi.StringArray{pulumi.String(vars.AdminUsername)},
					Type:       pulumi.String("simple"),
				},
				BrokerRackInitImage: nil,
				Config:              vars.KafkaClusterDefaultConfig,
				Listeners:           listenersInput,
				Replicas:            pulumi.Int(locals.KafkaKubernetes.Spec.BrokerContainer.Replicas),
				Resources: v1beta2.KafkaSpecKafkaResourcesArgs{
					Limits: pulumi.Map{
						"cpu":    pulumi.String(locals.KafkaKubernetes.Spec.BrokerContainer.Resources.Limits.Cpu),
						"memory": pulumi.String(locals.KafkaKubernetes.Spec.BrokerContainer.Resources.Limits.Memory),
					},
					Requests: pulumi.Map{
						"cpu":    pulumi.String(locals.KafkaKubernetes.Spec.BrokerContainer.Resources.Requests.Cpu),
						"memory": pulumi.String(locals.KafkaKubernetes.Spec.BrokerContainer.Resources.Requests.Memory),
					},
				},
				Storage: v1beta2.KafkaSpecKafkaStorageArgs{
					Type: pulumi.String("jbod"),
					Volumes: v1beta2.KafkaSpecKafkaStorageVolumesArray{
						v1beta2.KafkaSpecKafkaStorageVolumesArgs{
							DeleteClaim: pulumi.Bool(false),
							Id:          pulumi.Int(0),
							Size:        pulumi.String(locals.KafkaKubernetes.Spec.BrokerContainer.DiskSize),
							Type:        pulumi.String(strimzitypes.KafkaSpecKafkaStorageVolumesElemTypePersistentClaim),
						},
					},
				},
				Template:      nil,
				TieredStorage: nil,
				Version:       nil,
			},
			KafkaExporter:          nil,
			MaintenanceTimeWindows: nil,
			Zookeeper: v1beta2.KafkaSpecZookeeperArgs{
				Replicas: pulumi.Int(locals.KafkaKubernetes.Spec.BrokerContainer.Replicas),
				Storage: v1beta2.KafkaSpecZookeeperStorageArgs{
					DeleteClaim: pulumi.Bool(false),
					Size:        pulumi.String(vars.ZookeeperDefaultDiskSizeInGb),
					Type:        pulumi.String(strimzitypes.KafkaSpecZookeeperStorageTypePersistentClaim),
				},
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to create kafka-cluster")
	}
	return nil
}

func getBrokerListenersConfig(hostnames []string) v1beta2.KafkaSpecKafkaListenersConfigurationBrokersArray {
	resp := make([]v1beta2.KafkaSpecKafkaListenersConfigurationBrokersInput, len(hostnames))
	for i, hostName := range hostnames {
		resp[i] = v1beta2.KafkaSpecKafkaListenersConfigurationBrokersArgs{
			Broker:         pulumi.Int(i),
			AdvertisedHost: pulumi.String(hostName),
			AdvertisedPort: pulumi.Int(vars.ExternalPublicListenerPortNumber),
			Host:           pulumi.String(hostName),
		}
	}
	return resp
}
