package topics

import (
	"testing"

	code2cloudv1deploykfcmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kafkakubernetes/model"
)

func TestGetConfig(t *testing.T) {
	testCases := []struct {
		name          string
		defaultConfig *code2cloudv1deploykfcmodel.KafkaTopicConfig
		inputConfig   *code2cloudv1deploykfcmodel.KafkaTopicConfig
		expected      *code2cloudv1deploykfcmodel.KafkaTopicConfig
	}{
		{
			name: "input config nil",
			defaultConfig: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "compact",
			}},
			inputConfig: nil,
			expected: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "compact",
			}},
		},
		{
			name: "one key is passed in input config and is same as the default",
			defaultConfig: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "compact",
			}},
			inputConfig: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "compact",
			}},
			expected: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "compact",
			}},
		}, {
			name: "one key is passed in input config but is different from the default",
			defaultConfig: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "compact",
			}},
			inputConfig: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "delete",
			}},
			expected: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "delete",
			}},
		}, {
			name: "key not in the default config is specified in the input",
			defaultConfig: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"delete.retention.ms":                 "86400000",
				"max.message.bytes":                   "2097164",
				"message.timestamp.difference.max.ms": "9223372036854775807",
				"message.timestamp.type":              "CreateTime",
				"min.insync.replicas":                 "1",
				"retention.bytes":                     "-1",
				"retention.ms":                        "-1",
				"segment.bytes":                       "1073741824",
				"segment.ms":                          "604800000",
			}},
			inputConfig: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "compact",
			}},
			expected: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy":                      "compact",
				"delete.retention.ms":                 "86400000",
				"max.message.bytes":                   "2097164",
				"message.timestamp.difference.max.ms": "9223372036854775807",
				"message.timestamp.type":              "CreateTime",
				"min.insync.replicas":                 "1",
				"retention.bytes":                     "-1",
				"retention.ms":                        "-1",
				"segment.bytes":                       "1073741824",
				"segment.ms":                          "604800000",
			}},
		}, {
			name: "only one key in the default config is overridden in the input",
			defaultConfig: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
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
			}},
			inputConfig: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy": "compact",
			}},
			expected: &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{
				"cleanup.policy":                      "compact",
				"delete.retention.ms":                 "86400000",
				"max.message.bytes":                   "2097164",
				"message.timestamp.difference.max.ms": "9223372036854775807",
				"message.timestamp.type":              "CreateTime",
				"min.insync.replicas":                 "1",
				"retention.bytes":                     "-1",
				"retention.ms":                        "-1",
				"segment.bytes":                       "1073741824",
				"segment.ms":                          "604800000",
			}},
		},
	}
	t.Run("kafka topic configuration", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := getConfig(tc.defaultConfig, tc.inputConfig)
				if !isEqual(result, tc.expected) {
					t.Errorf("expected: %s got: %s", tc.expected, result)
				}
			})
		}
	})
}

func TestIsEqual(t *testing.T) {
	testCases := []struct {
		name     string
		this     *code2cloudv1deploykfcmodel.KafkaTopicConfig
		that     *code2cloudv1deploykfcmodel.KafkaTopicConfig
		expected bool
	}{
		{
			name:     "both are nil",
			this:     nil,
			that:     nil,
			expected: true,
		}, {
			name:     "this is nil",
			this:     nil,
			that:     &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: make(map[string]string, 0)},
			expected: false,
		}, {
			name:     "that is nil",
			this:     &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: make(map[string]string, 0)},
			that:     nil,
			expected: false,
		}, {
			name:     "different values for keys",
			this:     &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{"cleanup.policy": "delete"}},
			that:     &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{"cleanup.policy": "compact"}},
			expected: false,
		}, {
			name:     "same values for keys",
			this:     &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{"cleanup.policy": "delete"}},
			that:     &code2cloudv1deploykfcmodel.KafkaTopicConfig{Value: map[string]string{"cleanup.policy": "delete"}},
			expected: true,
		},
	}
	t.Run("compare topic configurations", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := isEqual(tc.this, tc.that)
				if result != tc.expected {
					t.Errorf("expected: %v got: %v", tc.expected, result)
				}
			})
		}
	})
}
