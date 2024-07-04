package topics

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	strimzitypes "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	code2cloudv1deploykfcmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

const (
	KafkaKubernetesLabelKey = "strimzi.io/cluster"
)

func Resources(ctx *pulumi.Context) error {
	i := extractInput(ctx)
	if i.topics == nil {
		return nil
	}
	for _, topic := range i.topics {
		if topic == nil {
			continue
		}
		if err := addTopic(ctx, topic); err != nil {
			return errors.Wrap(err, "failed to add topic")
		}
	}
	return nil
}

func addTopic(ctx *pulumi.Context, topic *code2cloudv1deploykfcmodel.KafkaTopic) error {
	i := extractInput(ctx)
	yamlPath := filepath.Join(i.workspaceDir, fmt.Sprintf("kafka-topic-%s.yaml", topic.Name))
	ir, _ := buildTopicObject(i, topic)
	if err := manifest.Create(yamlPath, ir); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", yamlPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, topic.Name, &pulumik8syaml.ConfigFileArgs{
		File: yamlPath,
	}, pulumi.Provider(i.kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to add topic kubernetes config file")
	}
	return nil
}

func buildTopicObject(i *input, topic *code2cloudv1deploykfcmodel.KafkaTopic) (*strimzitypes.KafkaTopic, error) {
	i.labels[KafkaKubernetesLabelKey] = i.resourceId
	kafkaTopicConfig := getConfig(defaultConfig, &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: topic.Config})
	configBytes, err := json.Marshal(kafkaTopicConfig.Value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to json encode kafka config")
	}
	return &strimzitypes.KafkaTopic{
		TypeMeta: v1.TypeMeta{
			APIVersion: "kafka.strimzi.io/v1beta2",
			Kind:       "KafkaTopic",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      topic.Name,
			Namespace: i.namespaceName,
			Labels:    i.labels,
		},
		Spec: &strimzitypes.KafkaTopicSpec{
			Config: &apiextensions.JSON{
				Raw: configBytes,
			},
			Partitions: pointer.Int32(getPartitions(topic)),
			Replicas:   pointer.Int32(getReplicas(topic)),
			TopicName:  &topic.Name,
		},
	}, nil
}

func getPartitions(topic *code2cloudv1deploykfcmodel.KafkaTopic) int32 {
	if topic.Partitions != 0 {
		return topic.Partitions
	}
	return DefaultPartitions
}

func getReplicas(topic *code2cloudv1deploykfcmodel.KafkaTopic) int32 {
	if topic.Replicas != 0 {
		return topic.Replicas
	}
	return DefaultReplicas
}
