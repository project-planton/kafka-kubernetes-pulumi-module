package cluster

import (
	"encoding/json"
	"path/filepath"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	strimzitypes "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/adminuser"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/broker"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/endpoint/cert"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/endpoint/hostname"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/listener"
	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/zookeeper"
	code2cloudv1deploykfcmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
	code2cloudv1deploykfcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/kubernetes/model"
	pulk8scv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/pointer"
)

type Input struct {
	Labels                           map[string]string
	WorkspaceDir                     string
	Namespace                        *pulk8scv1.Namespace
	NamespaceName                    string
	KafkaKubernetesKubernetesStackInput *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackInput
}

func Resources(ctx *pulumi.Context, input *Input) error {
	if err := addCluster(ctx, input); err != nil {
		return errors.Wrap(err, "failed to add kafka kubernetes")
	}
	return nil
}

func addCluster(ctx *pulumi.Context, input *Input) error {
	strimziCacheLoc := filepath.Join(input.WorkspaceDir, "strimzi")
	clusterYamlPath := filepath.Join(strimziCacheLoc, "kafka-kubernetes.yaml")
	kafkaKubernetes, _ := getClusterObject(input)
	if err := manifest.Create(clusterYamlPath, kafkaKubernetes); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", clusterYamlPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Metadata.Id,
		&pulumik8syaml.ConfigFileArgs{File: clusterYamlPath}, pulumi.DependsOn([]pulumi.Resource{input.Namespace}), pulumi.Parent(input.Namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add kafka kubernetes config file")
	}
	return nil
}

func getClusterObject(input *Input) (*strimzitypes.Kafka, error) {
	cfgBytes, err := json.Marshal(broker.DefaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to json encode kafka config")
	}
	brokerContainerResources := input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.Kubernetes.KafkaBrokerContainer.Resources

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
			Name:      input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Metadata.Id,
			Namespace: input.NamespaceName,
			Labels:    input.Labels,
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
				Listeners: getListenerElements(input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.EnvironmentInfo.EnvironmentName,
					input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.Kubernetes.Ingress.EndpointDomainName,
					input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes),
				Replicas: input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.Kubernetes.KafkaBrokerContainer.Replicas,
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
							Size:        pointer.String(input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.Kubernetes.KafkaBrokerContainer.DiskSize),
							Type:        strimzitypes.KafkaSpecKafkaStorageVolumesElemTypePersistentClaim,
						},
					},
				},
				Template: nil,
				Version:  nil,
			},
			Zookeeper: strimzitypes.KafkaSpecZookeeper{
				Replicas: input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.Kubernetes.ZookeeperContainer.Replicas,
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
func getListenerElements(productEnvName, domainName string, kafkaKubernetes *code2cloudv1deploykfcmodel.KafkaKubernetes) []strimzitypes.KafkaSpecKafkaListenersElem {
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
				Host: pointer.String(hostname.GetInternalBootstrapHostname(kafkaKubernetes.Metadata.Id,
					productEnvName, domainName)),
			},
			//private listeners should also advertise public listener port as the port translation is done by ingress controller
			Brokers: getInternalBrokerElements(kafkaKubernetes.Metadata.Id, productEnvName, domainName,
				kafkaKubernetes.Spec.Kubernetes.KafkaBrokerContainer.Replicas, listener.ExternalPublicListenerPortNumber),
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
				Host: pointer.String(hostname.GetExternalBootstrapHostname(kafkaKubernetes.Metadata.Id,
					productEnvName, domainName)),
			},
			Brokers: getExternalBrokerElements(kafkaKubernetes.Metadata.Id, productEnvName, domainName,
				kafkaKubernetes.Spec.Kubernetes.KafkaBrokerContainer.Replicas, listener.ExternalPublicListenerPortNumber),
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
