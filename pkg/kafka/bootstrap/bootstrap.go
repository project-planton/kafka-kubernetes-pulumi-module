package bootstrap

import (
	"fmt"
)

// GetKubeServiceName returns one of the following service names based on the provided input
// <kafka-kubernetes-id>-kafka-bootstrap
// <kafka-kubernetes-id>-kafka-brokers
// <kafka-kubernetes-id>-zookeeper-client
// <kafka-kubernetes-id>-zookeeper-nodes
// <kafka-kubernetes-id>-kafka-extpub-bootstrap
// <kafka-kubernetes-id>-kafka-extpvt-bootstrap
func GetKubeServiceName(kafkaKubernetesId, listenerName string) string {
	return fmt.Sprintf("%s-kafka-%s-bootstrap", kafkaKubernetesId, listenerName)
}

// GetInternalListenerKubeServiceName returns the name of the kubernetes service created by strimzi for the internal listener.
func GetInternalListenerKubeServiceName(kafkaKubernetesId string) string {
	return fmt.Sprintf("%s-bootstrap", kafkaKubernetesId)
}
