package hostname

import (
	"fmt"

	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kubernetes/broker"
	code2cloudv1deploykfcmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
)

func GetKafkaHostnames(productEnvName, endpointDomainName string, kafkaKubernetes *code2cloudv1deploykfcmodel.KafkaKubernetes) []string {
	hostnames := GetExternalHostnames(productEnvName, endpointDomainName, kafkaKubernetes)
	hostnames = append(hostnames, GetInternalHostnames(productEnvName, endpointDomainName, kafkaKubernetes)...)
	return hostnames
}

func GetExternalHostnames(productEnvName, domainName string, kafkaKubernetes *code2cloudv1deploykfcmodel.KafkaKubernetes) []string {
	hostnames := make([]string, 0)
	kafkaKubernetesId := kafkaKubernetes.Metadata.Id

	domainHostnames := getExternalHostnamesForDomain(kafkaKubernetesId, productEnvName, domainName, kafkaKubernetes.Spec.Kubernetes.KafkaBrokerContainer.Replicas)
	hostnames = append(hostnames, domainHostnames...)
	return hostnames
}

func GetInternalHostnames(productEnvName, domainName string, kafkaKubernetes *code2cloudv1deploykfcmodel.KafkaKubernetes) []string {
	hostnames := make([]string, 0)
	kafkaKubernetesId := kafkaKubernetes.Metadata.Id
	domainHostnames := getInternalHostnamesForDomain(kafkaKubernetesId, productEnvName, domainName, kafkaKubernetes.Spec.Kubernetes.KafkaBrokerContainer.Replicas)
	hostnames = append(hostnames, domainHostnames...)
	return hostnames
}

func getExternalHostnamesForDomain(kafkaKubernetesId, productEnvName, domainName string, brokerCount int32) []string {
	hostnames := make([]string, 0)
	hostnames = append(hostnames, GetExternalBootstrapHostname(kafkaKubernetesId, productEnvName, domainName))
	for i := 0; i < int(brokerCount); i++ {
		hostnames = append(hostnames, GetExternalBrokerHostname(kafkaKubernetesId, productEnvName, domainName, broker.Id(i)))
	}
	return hostnames
}

func getInternalHostnamesForDomain(kafkaKubernetesId, productEnvName, domainName string, brokerCount int32) []string {
	hostnames := make([]string, 0)
	hostnames = append(hostnames, GetInternalBootstrapHostname(kafkaKubernetesId, productEnvName, domainName))
	for i := 0; i < int(brokerCount); i++ {
		hostnames = append(hostnames, GetInternalBrokerHostname(kafkaKubernetesId, productEnvName, domainName, broker.Id(i)))
	}
	return hostnames
}

func GetExternalBootstrapHostname(kafkaKubernetesId, productEnvName, domainName string) string {
	return fmt.Sprintf("%s-bootstrap.%s.%s", kafkaKubernetesId, productEnvName, domainName)
}

func GetInternalBootstrapHostname(kafkaKubernetesId, productEnvName, domainName string) string {
	return fmt.Sprintf("%s-bootstrap.%s-internal.%s", kafkaKubernetesId, productEnvName, domainName)
}

func GetExternalBrokerHostname(kafkaKubernetesId, productEnvName, domainName string, brokerId broker.Id) string {
	return fmt.Sprintf("%s-broker-b%d.%s.%s", kafkaKubernetesId, brokerId, productEnvName, domainName)
}

func GetInternalBrokerHostname(kafkaKubernetesId, productEnvName, domainName string, brokerId broker.Id) string {
	return fmt.Sprintf("%s-broker-b%d.%s-internal.%s", kafkaKubernetesId, brokerId, productEnvName, domainName)
}

func GetExternalSchemaRegistryHostname(kafkaKubernetesId, productEnvName, domainName string) string {
	return fmt.Sprintf("%s-schema-registry.%s.%s", kafkaKubernetesId, productEnvName, domainName)
}

func GetInternalSchemaRegistryHostname(kafkaKubernetesId, productEnvName, domainName string) string {
	return fmt.Sprintf("%s-schema-registry.%s-internal.%s", kafkaKubernetesId, productEnvName, domainName)
}

func GetExternalKowlDashboardHostname(kafkaKubernetesId, productEnvName, domainName string) string {
	return fmt.Sprintf("%s-kowl.%s.%s", kafkaKubernetesId, productEnvName, domainName)
}

func GetInteralKowlDashboardHostname(kafkaKubernetesId, productEnvName, domainName string) string {
	return fmt.Sprintf("%s-kowl.%s-internal.%s", kafkaKubernetesId, productEnvName, domainName)
}
