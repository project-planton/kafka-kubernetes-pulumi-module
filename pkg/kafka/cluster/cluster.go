package cluster

import (
	"encoding/json"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/network/istio/cert"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/network/istio/hostname"
	"path/filepath"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	strimzitypes "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/adminuser"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/broker"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/listener"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/zookeeper"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/pointer"
)

func Resources(ctx *pulumi.Context) error {
	if err := addCluster(ctx); err != nil {
		return errors.Wrap(err, "failed to add kafka kubernetes")
	}
	return nil
}

func addCluster(ctx *pulumi.Context) error {
	i := extractInput(ctx)
	strimziCacheLoc := filepath.Join(i.workspaceDir, "strimzi")
	clusterYamlPath := filepath.Join(strimziCacheLoc, "kafka-kubernetes.yaml")
	kafkaKubernetes, _ := getClusterObject(i)
	if err := manifest.Create(clusterYamlPath, kafkaKubernetes); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", clusterYamlPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, i.resourceId,
		&pulumik8syaml.ConfigFileArgs{File: clusterYamlPath}, pulumi.DependsOn([]pulumi.Resource{i.namespace}), pulumi.Parent(i.namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add kafka kubernetes config file")
	}
	return nil
}

func getClusterObject(i *input) (*strimzitypes.Kafka, error) {
	cfgBytes, err := json.Marshal(broker.DefaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to json encode kafka config")
	}
	brokerContainerResources := i.brokerContainerSpec.Resources

	containerResourcesLimitsJsonBytes, err := json.Marshal(brokerContainerResources.Limits)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert broker container resource limits to json")
	}
	containerResourcesRequestsJsonBytes, err := json.Marshal(brokerContainerResources.Requests)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert broker container resource requests to json")
	}

	return &strimzitypes.Kafka{
		TypeMeta: v1.TypeMeta{
			Kind:       "Kafka",
			APIVersion: "kafka.strimzi.io/v1beta2",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      i.resourceId,
			Namespace: i.namespaceName,
			Labels:    i.labels,
		},
		Spec: &strimzitypes.KafkaSpec{
			EntityOperator: &strimzitypes.KafkaSpecEntityOperator{
				//todo: understand resource usage pattern and allocate
				TopicOperator: &strimzitypes.KafkaSpecEntityOperatorTopicOperator{},
				UserOperator:  &strimzitypes.KafkaSpecEntityOperatorUserOperator{},
			},
			Kafka: strimzitypes.KafkaSpecKafka{
				Authorization: &strimzitypes.KafkaSpecKafkaAuthorization{
					SuperUsers: []string{adminuser.Username},
					Type:       "simple",
				},
				BrokerRackInitImage: nil,
				Config: &apiextensions.JSON{
					Raw: cfgBytes,
				},
				Listeners: getListenerElements(i.environmentInfo.EnvironmentName, i.endpointDomainName, i.resourceId, i.brokerContainerSpec.Replicas),
				Replicas:  i.brokerContainerSpec.Replicas,
				Resources: &strimzitypes.KafkaSpecKafkaResources{
					Limits:   &apiextensions.JSON{Raw: containerResourcesLimitsJsonBytes},
					Requests: &apiextensions.JSON{Raw: containerResourcesRequestsJsonBytes},
				},
				Storage: strimzitypes.KafkaSpecKafkaStorage{
					Type: "jbod",
					Volumes: []strimzitypes.KafkaSpecKafkaStorageVolumesElem{
						{
							DeleteClaim: pointer.Bool(false),
							Id:          pointer.Int32(0),
							Size:        pointer.String(i.brokerContainerSpec.DiskSize),
							Type:        strimzitypes.KafkaSpecKafkaStorageVolumesElemTypePersistentClaim,
						},
					},
				},
				Template: nil,
				Version:  nil,
			},
			Zookeeper: strimzitypes.KafkaSpecZookeeper{
				Replicas: i.zookeeperContainerSpec.Replicas,
				Storage: strimzitypes.KafkaSpecZookeeperStorage{
					DeleteClaim: pointer.Bool(false),
					Size:        pointer.String(zookeeper.DefaultDiskSizeInGb),
					Type:        strimzitypes.KafkaSpecZookeeperStorageTypePersistentClaim,
				},
			},
		},
	}, nil
}

// getListenerElements returns both internal and external listener elements to be added to kafka spec.
func getListenerElements(productEnvName, domainName, kafkaKubernetesId string, brokerContainerReplicas int32) []strimzitypes.KafkaSpecKafkaListenersElem {
	resp := make([]strimzitypes.KafkaSpecKafkaListenersElem, 0)
	resp = append(resp, strimzitypes.KafkaSpecKafkaListenersElem{
		Name:           listener.InternalListenerName,
		Port:           listener.InternalListenerPortNumber,
		Authentication: &strimzitypes.KafkaSpecKafkaListenersElemAuthentication{Type: strimzitypes.KafkaSpecKafkaListenersElemAuthenticationTypeScramSha512},
		Tls:            false,
		Type:           strimzitypes.KafkaSpecKafkaListenersElemTypeInternal,
	})

	//internal listeners
	resp = append(resp, strimzitypes.KafkaSpecKafkaListenersElem{
		Name: listener.ExternalPrivateListenerName,
		Port: listener.ExternalPrivateListenerPortNumber,
		Tls:  true,
		Type: strimzitypes.KafkaSpecKafkaListenersElemTypeIngress,
		Authentication: &strimzitypes.KafkaSpecKafkaListenersElemAuthentication{
			Type: strimzitypes.KafkaSpecKafkaListenersElemAuthenticationTypeScramSha512},
		Configuration: &strimzitypes.KafkaSpecKafkaListenersElemConfiguration{
			Bootstrap: &strimzitypes.KafkaSpecKafkaListenersElemConfigurationBootstrap{
				Host: pointer.String(hostname.GetInternalBootstrapHostname(kafkaKubernetesId,
					productEnvName, domainName)),
			},
			//private listeners should also advertise public listener port as the port translation is done by ingress controller
			Brokers: getInternalBrokerElements(kafkaKubernetesId, productEnvName, domainName,
				brokerContainerReplicas, listener.ExternalPublicListenerPortNumber),
			BrokerCertChainAndKey: &strimzitypes.KafkaSpecKafkaListenersElemConfigurationBrokerCertChainAndKey{
				Certificate: "tls.crt",
				Key:         "tls.key",
				SecretName:  cert.GetCertSecretName(cert.Name),
			},
		},
	})

	//external listeners
	resp = append(resp, strimzitypes.KafkaSpecKafkaListenersElem{
		Name:           listener.ExternalPublicListenerName,
		Port:           listener.ExternalPublicListenerPortNumber,
		Tls:            true,
		Type:           strimzitypes.KafkaSpecKafkaListenersElemTypeIngress,
		Authentication: &strimzitypes.KafkaSpecKafkaListenersElemAuthentication{Type: strimzitypes.KafkaSpecKafkaListenersElemAuthenticationTypeScramSha512},
		Configuration: &strimzitypes.KafkaSpecKafkaListenersElemConfiguration{
			Bootstrap: &strimzitypes.KafkaSpecKafkaListenersElemConfigurationBootstrap{
				Host: pointer.String(hostname.GetExternalBootstrapHostname(kafkaKubernetesId, productEnvName, domainName)),
			},
			Brokers: getExternalBrokerElements(kafkaKubernetesId, productEnvName, domainName,
				brokerContainerReplicas, listener.ExternalPublicListenerPortNumber),
			BrokerCertChainAndKey: &strimzitypes.KafkaSpecKafkaListenersElemConfigurationBrokerCertChainAndKey{
				Certificate: "tls.crt",
				Key:         "tls.key",
				SecretName:  cert.GetCertSecretName(cert.Name),
			},
		},
	})

	return resp
}

func getExternalBrokerElements(kafkaKubernetesName, productEnvName, domainName string, replicas int32, port int32) []strimzitypes.KafkaSpecKafkaListenersElemConfigurationBrokersElem {
	resp := make([]strimzitypes.KafkaSpecKafkaListenersElemConfigurationBrokersElem, 0)
	for i := 0; i < int(replicas); i++ {
		resp = append(resp, strimzitypes.KafkaSpecKafkaListenersElemConfigurationBrokersElem{
			Broker: int32(i),
			// advertised host is what the broker will advertise the hostname using which to reach the broker
			AdvertisedHost: pointer.String(hostname.GetExternalBrokerHostname(kafkaKubernetesName, productEnvName, domainName, broker.Id(i))),
			// advertised port is what the broker will advertise the port using which to reach the broker
			AdvertisedPort: pointer.Int32(port),
			// value of host attribute is used by strimzi operator to configure ingress resource
			Host: pointer.String(hostname.GetExternalBrokerHostname(kafkaKubernetesName, productEnvName, domainName, broker.Id(i))),
		})
	}
	return resp
}

// todo: this can be deduplicated by creating external hostname and internal hostname at the beginning of the stack.
// because of bad design, the internal and external domain code is polluted everywhere. this becomes a maintenance nightmare if not fixed.
func getInternalBrokerElements(kafkaKubernetesId, productEnvName, domainName string, replicas int32, port int32) []strimzitypes.KafkaSpecKafkaListenersElemConfigurationBrokersElem {
	resp := make([]strimzitypes.KafkaSpecKafkaListenersElemConfigurationBrokersElem, 0)
	for i := 0; i < int(replicas); i++ {
		resp = append(resp, strimzitypes.KafkaSpecKafkaListenersElemConfigurationBrokersElem{
			Broker: int32(i),
			// advertised host is what the broker will advertise the hostname using which to reach the broker
			AdvertisedHost: pointer.String(hostname.GetInternalBrokerHostname(kafkaKubernetesId, productEnvName, domainName, broker.Id(i))),
			// advertised port is what the broker will advertise the port using which to reach the broker
			AdvertisedPort: pointer.Int32(port),
			// value of host attribute is used by strimzi operator to configure ingress resource
			Host: pointer.String(hostname.GetInternalBrokerHostname(kafkaKubernetesId, productEnvName, domainName, broker.Id(i))),
		})
	}
	return resp
}
