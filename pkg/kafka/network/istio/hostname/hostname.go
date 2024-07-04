package hostname

import (
	"fmt"

	"github.com/plantoncloud/kafka-kubernetes-pulumi-blueprint/pkg/kafka/broker"
)

func GetKafkaHostnames(productEnvName, endpointDomainName, kafkaKubernetesId string, brokerReplicas int32) []string {
	hostnames := GetExternalHostnames(productEnvName, endpointDomainName, kafkaKubernetesId, brokerReplicas)
	hostnames = append(hostnames, GetInternalHostnames(productEnvName, endpointDomainName, kafkaKubernetesId, brokerReplicas)...)
	return hostnames
}

func GetExternalHostnames(productEnvName, domainName, kafkaKubernetesId string, brokerReplicas int32) []string {
	hostnames := make([]string, 0)

	domainHostnames := getExternalHostnamesForDomain(kafkaKubernetesId, productEnvName, domainName, brokerReplicas)
	hostnames = append(hostnames, domainHostnames...)
	return hostnames
}

func GetInternalHostnames(productEnvName, domainName, kafkaKubernetesId string, brokerReplicas int32) []string {
	hostnames := make([]string, 0)
	domainHostnames := getInternalHostnamesForDomain(kafkaKubernetesId, productEnvName, domainName, brokerReplicas)
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
