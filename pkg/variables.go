package pkg

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

var vars = struct {
	ExternalPublicListenerName        string
	ExternalPublicListenerPortNumber  int
	ExternalPrivateListenerName       string
	ExternalPrivateListenerPortNumber int
	InternalListenerName              string
	InternalListenerPortNumber        int
	AdminUsername                     string
	ClusterLabelKey                   string
	SaslPasswordSecretName            string
	SaslJaasConfigKeyInSecret         string
	SaslPasswordKeyInSecret           string
	KafkaClusterDefaultConfig         pulumi.Map
	CertName                          string
	CertSecretName                    string
	IstioIngressNamespace             string
	IstioIngressSelectorLabels        map[string]string
	KafkaTopicDefaultConfig           map[string]string
	ZookeeperDefaultDiskSizeInGb      string
	SchemaRegistryDockerImage         string
	SchemaRegistryContainerPort       int
	SchemaRegistryKafkaStoreTopicName string
	SchemaRegistryDeploymentName      string
}{
	ExternalPublicListenerName:        "extpub",
	ExternalPublicListenerPortNumber:  9092, //this port is intended to be used by clients output the private network and outside the container cluster
	ExternalPrivateListenerName:       "extpvt",
	ExternalPrivateListenerPortNumber: 9093, //this port is intended to be used by clients inside the private network but outside the container cluster
	InternalListenerName:              "int",
	InternalListenerPortNumber:        9094, //this port is intended to be used by clients inside the container cluster
	AdminUsername:                     "admin",
	ClusterLabelKey:                   "strimzi.io/cluster",
	SaslPasswordSecretName:            "admin",
	SaslJaasConfigKeyInSecret:         "sasl.jaas.config",
	SaslPasswordKeyInSecret:           "password",

	KafkaClusterDefaultConfig: pulumi.Map{
		"offsets.topic.replication.factor":         pulumi.Int(1),
		"transaction.state.log.replication.factor": pulumi.Int(1),
		"transaction.state.log.min.isr":            pulumi.Int(1),
		"auto.create.topics.enable":                pulumi.Bool(true),
	},

	CertName:       "kafka-ingress",
	CertSecretName: "cert-kafka-ingress",

	IstioIngressNamespace: "istio-ingress",
	IstioIngressSelectorLabels: map[string]string{
		"app":   "istio-ingress",
		"istio": "ingress",
	},

	KafkaTopicDefaultConfig: map[string]string{
		"cleanup.policy":                      "delete",
		"delete.retention.ms":                 "86400000",
		"max.message.bytes":                   "2097164",
		"message.timestamp.difference.max.ms": "9223372036854775807",
		"message.timestamp.type":              "CreateTime",
		"min.insync.replicas":                 "1",
		"retention.bytes":                     "-1",
		"retention.ms":                        "-1",
		"segment.bytes":                       "1073741824",
		"segment.ms":                          "604800000",
	},
	ZookeeperDefaultDiskSizeInGb:      "1Gi",
	SchemaRegistryDockerImage:         "confluentinc/cp-schema-registry:7.2.6",
	SchemaRegistryContainerPort:       8081,
	SchemaRegistryKafkaStoreTopicName: "schema-registry",
	SchemaRegistryDeploymentName:      "schema-registry",
}
