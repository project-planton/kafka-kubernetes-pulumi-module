package topics

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	strimzitypes "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	code2cloudv1deploykfcmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
	code2cloudv1deploykfcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/stack/kubernetes/model"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

const (
	KafkaKubernetesLabelKey = "strimzi.io/cluster"
)

type Input struct {
	KubernetesProvider               *pulumikubernetes.Provider
	WorkspaceDir                     string
	NamespaceName                    string
	KafkaKubernetesKubernetesStackInput *code2cloudv1deploykfcstackk8smodel.KafkaKubernetesKubernetesStackInput
	Labels                           map[string]string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	if input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.KafkaTopics == nil {
		return nil
	}
	for _, topic := range input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Spec.KafkaTopics {
		if topic == nil {
			continue
		}
		if err := addTopic(ctx, input, topic); err != nil {
			return errors.Wrap(err, "failed to add topic")
		}
	}
	return nil
}

func addTopic(ctx *pulumi.Context, input *Input, topic *code2cloudv1deploykfcmodel.KafkaTopic) error {
	yamlPath := filepath.Join(input.WorkspaceDir, fmt.Sprintf("kafka-topic-%s.yaml", topic.Name))
	ir, _ := buildTopicObject(input, topic)
	if err := manifest.Create(yamlPath, ir); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", yamlPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, topic.Name, &pulumik8syaml.ConfigFileArgs{
		File: yamlPath,
	}, pulumi.Provider(input.KubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to add topic kubernetes config file")
	}
	return nil
}

func buildTopicObject(input *Input, topic *code2cloudv1deploykfcmodel.KafkaTopic) (*strimzitypes.KafkaTopic, error) {
	input.Labels[KafkaKubernetesLabelKey] = input.KafkaKubernetesKubernetesStackInput.ResourceInput.KafkaKubernetes.Metadata.Id
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
			Namespace: input.NamespaceName,
			Labels:    input.Labels,
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
