package topics

import (
	code2cloudv1deploykfcmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
)

const (
	DefaultPartitions = 1
	DefaultReplicas   = 1
)

var defaultConfig = &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
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
}}

//	config.Prod: {
//		{TopicName: "comm-state", Partitions: 1, Replicas: 3},
//		{TopicName: "cur-state", Partitions: 1, Replicas: 3},
//		{TopicName: "cur-mkt-pair-state", Partitions: 1, Replicas: 3},
//		{TopicName: "order-cmd", Partitions: 6, Replicas: 3},
//		{TopicName: "order-state", Partitions: 6, Replicas: 3},
//		{TopicName: "order-repartition-mktpair", Partitions: 6, Replicas: 3},
//		{TopicName: "order-trade-cmd", Partitions: 6, Replicas: 3},
//		{TopicName: "order-trade-state", Partitions: 6, Replicas: 3},
//		{TopicName: "order-trade-repartition-windowed-mktpair", Partitions: 1, Replicas: 3},
//		{TopicName: "order-trade-timeseries-state", Partitions: 6, Replicas: 3},
//		{TopicName: "order-book-state", Partitions: 6, Replicas: 3},
//		{TopicName: "transfer-fiat-bank-state", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-fiat-bankacct-state", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-fiat-txn-cmd", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-fiat-txn-state", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-fiat-txn-wh-state", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-fiat-txn-repartition-uuid", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-crypto-txn-cmd", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-crypto-txn-state", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-crypto-txn-wh-state", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-crypto-txn-repartition-nwtxnid", Partitions: 1, Replicas: 3},
//		{TopicName: "transfer-crypto-wallet-state", Partitions: 1, Replicas: 3},
//		{TopicName: "user-acct-state", Partitions: 1, Replicas: 3},
//		{TopicName: "user-profile-state", Partitions: 1, Replicas: 3},
//		{TopicName: "user-kyc-state", Partitions: 1, Replicas: 3},
//		{TopicName: "wallet-cmd", Partitions: 6, Replicas: 3},
//		{TopicName: "wallet-state", Partitions: 6, Replicas: 3},
//	},
//}

//func getProdTopics() ([]*strimzitypes.KafkaTopicSpec, error) {
//	for _, t := range topics[config.Prod] {
//		if t.Config == nil {
//			t.Config = DefaultCfg
//		}
//		t.Replicas = 3
//		t.Config[MinInSyncReplicasCfgKey] = 2
//	}
//	topics, err := marshallTopics(topics[config.Prod])
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to marshal prod topics")
//	}
//	return topics, nil
//}

func getConfig(defaultConfig, inputConfig *code2cloudv1deploykfcmodel.KafkaTopicConfig) *code2cloudv1deploykfcmodel.KafkaTopicConfig {
	finalConfig := make(map[string]string, 0)
	if inputConfig == nil {
		return &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: defaultConfig.Value}
	}
	for k, v := range defaultConfig.Value {
		finalConfig[k] = v
	}
	for k, v := range inputConfig.Value {
		finalConfig[k] = v
	}
	return &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: finalConfig}
}

func isEqual(this, that *code2cloudv1deploykfcmodel.KafkaTopicConfig) bool {
	if this == nil && that == nil {
		return true
	}
	if (this == nil && that != nil) || (this != nil && that == nil) {
		return false
	}
	for thisKey, thisValue := range this.Value {
		if that.Value[thisKey] != thisValue {
			return false
		}
	}
	return true
}
