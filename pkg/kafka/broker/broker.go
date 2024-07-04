package broker

import (
	"fmt"
)

type (
	Id int32
)

var DefaultConfig = map[string]interface{}{
	"offsets.topic.replication.factor":         1,
	"transaction.state.log.replication.factor": 1,
	"transaction.state.log.min.isr":            1,
	"auto.create.topics.enable":                true,
}

// GetKubeServiceName for each broker an external and and internal listener service
// <kafka-kubernetes-id>-kafka-extpub-0
// <kafka-kubernetes-id>-kafka-extpvt-0
func GetKubeServiceName(kafkaKubernetesId, listenerName string, brokerId Id) string {
	return fmt.Sprintf("%s-kafka-%s-%d", kafkaKubernetesId, listenerName, brokerId)
}
